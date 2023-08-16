package query

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

type DefaultExpression struct {
	Contains   string
	Expression string
}

type QueryParser struct {
	Default   DefaultExpression
	Tables    map[string]string
	TzOffset  int
	Dialector string
}

type ParsedQuery interface {
	Filter(*gorm.DB) *gorm.DB
}

type parsedQuery struct {
	qp         *QueryParser
	joins      map[string]join
	conditions []clause.Expression
}

type callable func(args []ast.Expr) (any, error)

type attributeGetter func(attr string) (any, error)

type subscriptSlicer func(index ast.Slicer) (any, error)

type join struct {
	alias string
	query string
	args  []any
}

type SyntaxError struct {
	Statement string `json:"statement"`
	Line      int    `json:"line"`
	Offset    int    `json:"offset"`
	EndOffset int    `json:"end_offset"`
	Err       string `json:"error,omitempty"`
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at (%d, %d) in %q: %s", s.Line, s.Offset, s.Statement, s.Err)
}

func (s SyntaxError) Detail() any {
	return s
}

func (s SyntaxError) Message() string {
	return "SyntaxError"
}

func (s SyntaxError) Code() int {
	return fiber.StatusBadRequest
}

func (s SyntaxError) Is(target error) bool {
	_, ok := target.(SyntaxError)
	return ok
}

func wrapError(e error, q string) error {
	switch e := e.(type) {
	case *py.Exception:
		if py.SyntaxError.IsSubtype(e.Base.Type()) {
			s := SyntaxError{
				Statement: q,
				Err:       "invalid syntax",
			}
			if l, ok := e.Dict["lineno"]; ok {
				if l, ok := l.(py.Int); ok {
					if l, err := l.GoInt(); err == nil {
						s.Line = l
					}
				}
			}
			if o, ok := e.Dict["offset"]; ok {
				if o, ok := o.(py.Int); ok {
					if o, err := o.GoInt(); err == nil {
						s.Offset = o
					}
				}
			}
			return s
		}
	case SyntaxError:
		e.Statement = q
		return e
	}
	return e
}

func (qp *QueryParser) Parse(q string) (ParsedQuery, error) {
	pq := &parsedQuery{
		qp:    qp,
		joins: make(map[string]join),
	}

	if q == "" {
		if qp.Default.Expression == "" {
			return pq, nil
		}
		q = qp.Default.Expression
	}

	if !strings.Contains(q, qp.Default.Contains) {
		q = fmt.Sprintf("(%s) and (%s)", q, qp.Default.Expression)
	}

	a, err := parser.ParseString(q, py.EvalMode)
	if err != nil {
		return nil, wrapError(err, q)
	}

	e, ok := a.(*ast.Expression)
	if !ok {
		return nil, fmt.Errorf("not a valid Python expression: %#v", a)
	}

	cl, err := pq.parseNode(e.Body)
	if err != nil {
		return nil, wrapError(err, q)
	}

	cond, ok := cl.(clause.Expression)
	if !ok {
		return nil, fmt.Errorf("not a valid SQL expression: %#v", cl)
	}

	pq.conditions = append(pq.conditions, cond)

	return pq, nil
}

func (pq *parsedQuery) Filter(tx *gorm.DB) *gorm.DB {
	for _, j := range pq.joins {
		tx.Joins(j.query, j.args...)
	}
	if len(pq.conditions) > 0 {
		tx.Where(clause.And(pq.conditions...))
	}
	return tx
}

func (pq *parsedQuery) parseNode(node ast.Expr) (any, error) {
	ret, err := pq._parseNode(node)
	if err != nil && !errors.Is(err, SyntaxError{}) {
		return nil, SyntaxError{
			Line:   node.GetLineno(),
			Offset: node.GetColOffset() + 3,
			Err:    err.Error(),
		}
	}
	return ret, err
}

func (pq *parsedQuery) _parseNode(node ast.Expr) (any, error) {
	switch n := node.(type) {
	case *ast.BoolOp:
		return pq.parseBoolOp(n)
	case *ast.Call:
		return pq.parseCall(n)
	case *ast.List:
		return pq.parseList(n)
	case *ast.Name:
		return pq.parseName(n)
	case *ast.NameConstant:
		return pq.parseNameConstant(n)
	case *ast.Num:
		return pq.parseNum(n)
	case *ast.Str:
		return pq.parseStr(n)
	case *ast.Subscript:
		return pq.parseSubscript(n)
	case *ast.UnaryOp:
		return pq.parseUnaryOp(n)
	case *ast.Attribute:
		return pq.parseAttribute(n)
	case *ast.Compare:
		return pq.parseCompare(n)
	default:
		return nil, fmt.Errorf("unsupported expression %q", ast.Dump(n))
	}
}

func (pq *parsedQuery) parseAttribute(node *ast.Attribute) (any, error) {
	switch node.Ctx {
	case ast.Load:
		parsedNode, err := pq.parseNode(node.Value)
		if err != nil {
			return nil, err
		}
		attribute := string(node.Attr)
		switch strings.ToLower(attribute) {
		case "endswith":
			return callable(func(args []ast.Expr) (any, error) {
				if len(args) != 1 {
					return nil, errors.New("`endwith` function support exactly one argument")
				}
				c, ok := parsedNode.(clause.Column)
				if !ok {
					return nil, errors.New("unsupported node type. has to be clause.Column")
				}

				arg, ok := args[0].(*ast.Str)
				if !ok {
					return nil, errors.New("unsupported argument type. has to be `string` only")
				}
				return clause.Like{
					Value: fmt.Sprintf("%%%s", arg.S),
					Column: clause.Column{
						Table: c.Table,
						Name:  c.Name,
					},
				}, nil
			}), nil
		case "startswith":
			return callable(func(args []ast.Expr) (any, error) {
				if len(args) != 1 {
					return nil, errors.New("`startwith` function support exactly one argument")
				}
				c, ok := parsedNode.(clause.Column)
				if !ok {
					return nil, errors.New("unsupported node type. has to be clause.Column")
				}

				arg, ok := args[0].(*ast.Str)
				if !ok {
					return nil, errors.New("unsupported argument type. has to be `string` only")
				}
				return clause.Like{
					Value: fmt.Sprintf("%s%%", arg.S),
					Column: clause.Column{
						Table: c.Table,
						Name:  c.Name,
					},
				}, nil
			}), nil
		}

		switch value := parsedNode.(type) {
		case attributeGetter:
			return value(attribute)
		default:
			return nil, fmt.Errorf("unsupported attribute value %#parsedNode", value)
		}
	default:
		return nil, fmt.Errorf("unsupported attribute context %q", node.Ctx)
	}
}

func (pq *parsedQuery) parseBoolOp(node *ast.BoolOp) (any, error) {
	exprs := make([]clause.Expression, len(node.Values))
	for i, v := range node.Values {
		e, err := pq.parseNode(v)
		if err != nil {
			return nil, err
		}
		c, ok := e.(clause.Expression)
		if !ok {
			return nil, fmt.Errorf("not a valid SQL expression: %#v", e)
		}
		exprs[i] = c
	}
	switch node.Op {
	case ast.And:
		return clause.And(exprs...), nil
	case ast.Or:
		return clause.Or(exprs...), nil
	default:
		return nil, fmt.Errorf("unsupported boolean operation %q", node.Op)
	}
}

func (pq *parsedQuery) parseCall(node *ast.Call) (any, error) {
	f, err := pq.parseNode(node.Func)
	if err != nil {
		return nil, err
	}

	switch f := f.(type) {
	case callable:
		return f(node.Args)
	default:
		return nil, fmt.Errorf("unsupported call to function %#v", node.Func)
	}
}

func (pq *parsedQuery) parseCompare(node *ast.Compare) (any, error) {
	exprs := make([]clause.Expression, len(node.Ops))

	for i, op := range node.Ops {
		leftAst := node.Left
		if i > 0 {
			leftAst = node.Comparators[i-1]
		}
		left, err := pq.parseNode(leftAst)
		if err != nil {
			return nil, err
		}
		right, err := pq.parseNode(node.Comparators[i])
		if err != nil {
			return nil, err
		}

		switch left := left.(type) {
		case clause.Column:
			exprs[i], err = newSqlComparison(op, left, right)
			if err != nil {
				return nil, err
			}
		case clause.Eq:
			switch right := right.(type) {
			case bool:
				exprs[i], err = newSqlBoolComparison(op, left, right)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unsupported comparison %q", ast.Dump(node))
			}
		default:
			switch right := right.(type) {
			case clause.Column:
				switch op {
				case ast.In:
					// for `IN` statement, left parameter has to be always `string`.
					if _, ok := left.(string); !ok {
						return nil, errors.New("left parameter has to be a string")
					}
					return clause.Like{
						Value:  fmt.Sprintf("%%%s%%", left),
						Column: right,
					}, nil
				case ast.NotIn:
					// for `NOT IN` statement, left parameter has to be always `string`.
					if _, ok := left.(string); !ok {
						return nil, errors.New("left parameter has to be a string")
					}
					return negativeClause(clause.Like{
						Value:  fmt.Sprintf("%%%s%%", left),
						Column: right,
					}), nil
				default:
					o, l, r, err := reverseComparison(op, left, right)
					if err != nil {
						return nil, err
					}
					exprs[i], err = newSqlComparison(o, l, r)
					if err != nil {
						return nil, err
					}
				}
			case clause.Eq:
				switch left := left.(type) {
				case bool:
					exprs[i], err = newSqlBoolComparison(op, right, left)
					if err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("unsupported comparison %q", ast.Dump(node))
				}
			}
		}
	}

	return clause.AndConditions{
		Exprs: exprs,
	}, nil
}

func (pq *parsedQuery) parseList(node *ast.List) (any, error) {
	var err error
	list := make([]any, len(node.Elts))
	for i, e := range node.Elts {
		list[i], err = pq.parseNode(e)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (pq *parsedQuery) parseName(node *ast.Name) (any, error) {
	switch node.Ctx {
	case ast.Load:
		switch string(node.Id) {
		case "run":
			table, ok := pq.qp.Tables["runs"]
			if !ok {
				return nil, errors.New("unsupported name identifier 'run'")
			}
			return attributeGetter(
				func(attr string) (any, error) {
					switch attr {
					case "creation_time", "created_at":
						return clause.Column{
							Table: table,
							Name:  "start_time",
						}, nil
					case "end_time", "finalized_at":
						return clause.Column{
							Table: table,
							Name:  "end_time",
						}, nil
					case "hash":
						return clause.Column{
							Table: table,
							Name:  "run_uuid",
						}, nil
					case "name":
						return clause.Column{
							Table: table,
							Name:  "name",
						}, nil
					case "experiment":
						e, ok := pq.qp.Tables["experiments"]
						if !ok {
							return nil, errors.New("unsupported attribute 'experiment'")
						}
						return clause.Column{
							Table: e,
							Name:  "name",
						}, nil
					case "archived":
						return clause.Eq{
							Column: clause.Column{
								Table: table,
								Name:  "lifecycle_stage",
							},
							Value: models.LifecycleStageDeleted,
						}, nil
					case "active":
						return clause.Eq{
							Column: clause.Column{
								Table: table,
								Name:  "status",
							},
							Value: models.StatusRunning,
						}, nil
					case "duration":
						return clause.Column{
							Name: fmt.Sprintf("(%s.end_time - %s.start_time) / 1000", table, table),
							Raw:  true,
						}, nil
					case "metrics":
						return subscriptSlicer(func(s ast.Slicer) (any, error) {
							switch s := s.(type) {
							case *ast.Index:
								v, err := pq.parseNode(s.Value)
								if err != nil {
									return nil, err
								}
								switch v := v.(type) {
								case string:
									j, ok := pq.joins[fmt.Sprintf("metrics:%s", v)]
									if !ok {
										alias := fmt.Sprintf("metrics_%d", len(pq.joins))
										j = join{
											alias: alias,
											query: fmt.Sprintf("LEFT JOIN latest_metrics %s ON %s.run_uuid = %s.run_uuid AND %s.key = ?", alias, table, alias, alias),
											args:  []any{v},
										}
										pq.joins[fmt.Sprintf("metrics:%s", v)] = j
									}
									return attributeGetter(func(attr string) (any, error) {
										var name string
										switch attr {
										case "last":
											name = "value"
										case "last_step":
											name = "last_iter"
										case "first_step":
											return 0, nil
										default:
											return nil, fmt.Errorf("unsupported metrics attribute %q", attr)
										}
										return clause.Column{
											Table: j.alias,
											Name:  name,
										}, nil
									}), nil
								default:
									return nil, fmt.Errorf("unsupported index value type %t", v)
								}
							default:
								return nil, fmt.Errorf("unsupported slicer %q", ast.Dump(s))
							}
						}), nil
					case "tags":
						return subscriptSlicer(func(s ast.Slicer) (any, error) {
							switch s := s.(type) {
							case *ast.Index:
								v, err := pq.parseNode(s.Value)
								if err != nil {
									return nil, err
								}
								switch v := v.(type) {
								case string:
									j, ok := pq.joins[fmt.Sprintf("tags:%s", v)]
									if !ok {
										alias := fmt.Sprintf("tags_%d", len(pq.joins))
										j = join{
											alias: alias,
											query: fmt.Sprintf("LEFT JOIN tags %s ON %s.run_uuid = %s.run_uuid AND %s.key = ?", alias, table, alias, alias),
											args:  []any{v},
										}
										pq.joins[fmt.Sprintf("tags:%s", v)] = j
									}
									return clause.Column{
										Table: j.alias,
										Name:  "value",
									}, nil
								default:
									return nil, fmt.Errorf("unsupported index value type %t", v)
								}
							default:
								return nil, fmt.Errorf("unsupported slicer %q", ast.Dump(s))
							}
						}), nil
					default:
						j, ok := pq.joins[fmt.Sprintf("params:%s", attr)]
						if !ok {
							alias := fmt.Sprintf("params_%d", len(pq.joins))
							j = join{
								alias: alias,
								query: fmt.Sprintf("LEFT JOIN params %s ON %s.run_uuid = %s.run_uuid AND %s.key = ?", alias, table, alias, alias),
								args:  []any{attr},
							}
							pq.joins[fmt.Sprintf("params:%s", attr)] = j
						}
						return clause.Column{
							Table: j.alias,
							Name:  "value",
						}, nil
					}
				},
			), nil
		case "metric":
			table, ok := pq.qp.Tables["metrics"]
			if !ok {
				return nil, errors.New("unsupported name identifier 'metric'")
			}
			return attributeGetter(
				func(attr string) (any, error) {
					switch attr {
					case "name":
						return clause.Column{
							Table: table,
							Name:  "key",
						}, nil
					case "last":
						return clause.Column{
							Table: table,
							Name:  "value",
						}, nil
					case "last_step":
						return clause.Column{
							Table: table,
							Name:  "last_iter",
						}, nil
					case "first_step":
						return 0, nil
					default:
						return nil, fmt.Errorf("unsupported metrics attribute %q", attr)
					}
				},
			), nil
		case "re":
			return attributeGetter(
				func(attr string) (any, error) {
					switch attr {
					case "match":
						fallthrough
					case "search":
						return callable(
							func(args []ast.Expr) (any, error) {
								if len(args) != 2 {
									return nil, errors.New("re.match function support exactly 2 arguments")
								}

								parsedNode, err := pq.parseNode(args[0])
								if err != nil {
									return nil, err
								}
								str, ok := parsedNode.(string)
								if !ok {
									return nil, errors.New("first argument type for re.match function has to be a string")
								}

								parsedNode, err = pq.parseNode(args[1])
								if err != nil {
									return nil, err
								}
								column, ok := parsedNode.(clause.Column)
								if !ok {
									return nil, errors.New(
										"second argument type for re.match function has to be clause.Column",
									)
								}

								// handle difference between `match` and `search`.
								if attr == "match" {
									str = fmt.Sprintf("^%s", str)
								}

								return Regexp{
									Eq: clause.Eq{
										Column: column,
										Value:  str,
									},
									Dialector: pq.qp.Dialector,
								}, nil
							},
						), nil
					default:
						return nil, fmt.Errorf("unsupported re function %s", attr)
					}
				},
			), nil
		case "datetime":
			return callable(
				func(args []ast.Expr) (any, error) {
					if len(args) > 7 {
						return nil, fmt.Errorf("too many arguments for datetime: %d", len(args))
					}
					intArgs := make([]int, 7)
					for i, a := range args {
						e, err := pq.parseNode(a)
						if err != nil {
							return nil, err
						}
						n, ok := e.(int)
						if !ok {
							return nil, fmt.Errorf("unsupported argument %d to datetime: %#v", i, a)
						}
						intArgs[i] = n
					}
					return time.Date(
						intArgs[0],
						time.Month(intArgs[1]),
						intArgs[2],
						intArgs[3],
						intArgs[4],
						intArgs[5],
						intArgs[6]*1000,
						time.FixedZone("custom", -pq.qp.TzOffset*60),
					).UnixMilli(), nil
				},
			), nil
		default:
			return nil, fmt.Errorf("unsupported name identifier %q", node.Id)
		}
	default:
		return nil, fmt.Errorf("unsupported name context %q", node.Ctx)
	}
}

func (pq *parsedQuery) parseNameConstant(node *ast.NameConstant) (any, error) {
	switch node.Value.Type() {
	case py.NoneTypeType:
		return nil, nil
	case py.BoolType:
		return bool(node.Value.(py.Bool)), nil
	default:
		return nil, fmt.Errorf("unsupported name constant type %q", node.Value.Type())
	}
}

func (pq *parsedQuery) parseNum(node *ast.Num) (any, error) {
	switch node.N.Type() {
	case py.IntType:
		return node.N.(py.Int).GoInt()
	case py.FloatType:
		return py.FloatAsFloat64(node.N.(py.Float))
	default:
		return nil, fmt.Errorf("unsupported num type %q", node.N.Type())
	}
}

func (pq *parsedQuery) parseStr(node *ast.Str) (any, error) {
	return string(node.S), nil
}

func (pq *parsedQuery) parseSubscript(node *ast.Subscript) (any, error) {
	switch node.Ctx {
	case ast.Load:
		v, err := pq.parseNode(node.Value)
		if err != nil {
			return nil, err
		}
		switch v := v.(type) {
		case subscriptSlicer:
			return v(node.Slice)
		default:
			return nil, fmt.Errorf("unsupported attribute value %#v", v)
		}
	default:
		return nil, fmt.Errorf("unsupported attribute context %q", node.Ctx)
	}
}

func (pq *parsedQuery) parseUnaryOp(node *ast.UnaryOp) (any, error) {
	e, err := pq.parseNode(node.Operand)
	if err != nil {
		return nil, err
	}
	c, ok := e.(clause.Expression)
	if !ok {
		return nil, fmt.Errorf("not a valid SQL expression: %#v", e)
	}
	switch node.Op {
	case ast.Not:
		return clause.Not(c), nil
	default:
		return nil, fmt.Errorf("unsupported unary operation %q", node.Op)
	}
}

func newSqlBoolComparison(op ast.CmpOp, left clause.Eq, right bool) (clause.Expression, error) {
	switch op {
	case ast.Eq, ast.Is:
		if right {
			return left, nil
		}
		return clause.Not(left), nil
	case ast.NotEq, ast.IsNot:
		if !right {
			return left, nil
		}
		return clause.Not(left), nil
	default:
		return nil, fmt.Errorf("comparison operation incompatible with bool %q", op)
	}
}

func newSqlComparison(op ast.CmpOp, left clause.Column, right any) (clause.Expression, error) {
	switch op {
	case ast.Eq, ast.Is:
		return clause.Eq{
			Column: left,
			Value:  right,
		}, nil
	case ast.NotEq, ast.IsNot:
		return clause.Neq{
			Column: left,
			Value:  right,
		}, nil
	case ast.Lt:
		return clause.Lt{
			Column: left,
			Value:  right,
		}, nil
	case ast.LtE:
		return clause.Lte{
			Column: left,
			Value:  right,
		}, nil
	case ast.Gt:
		return clause.Gt{
			Column: left,
			Value:  right,
		}, nil
	case ast.GtE:
		return clause.Gte{
			Column: left,
			Value:  right,
		}, nil
	case ast.In:
		r, ok := right.([]any)
		if !ok {
			return nil, fmt.Errorf("right value in \"in\" comparison is not a list: %#v", right)
		}
		return clause.IN{
			Column: left,
			Values: r,
		}, nil
	case ast.NotIn:
		r, ok := right.([]any)
		if !ok {
			return nil, fmt.Errorf("right value in \"not in\" comparison is not a list: %#v", right)
		}
		return negativeClause(clause.IN{
			Column: left,
			Values: r,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported comparison operation %q", op)
	}
}

func reverseComparison(op ast.CmpOp, left any, right clause.Column) (ast.CmpOp, clause.Column, any, error) {
	switch op {
	case ast.Lt:
		return ast.Gt, right, left, nil
	case ast.LtE:
		return ast.GtE, right, left, nil
	case ast.Gt:
		return ast.Lt, right, left, nil
	case ast.GtE:
		return ast.LtE, right, left, nil
	case ast.Eq, ast.Is, ast.NotEq, ast.IsNot:
		return op, right, left, nil
	default:
		return op, right, left, fmt.Errorf("unable to reverse comparison operator %q", op)
	}
}

func negativeClause(expression clause.Expression) clause.Expression {
	return clause.NotConditions{
		Exprs: []clause.Expression{
			expression,
		},
	}
}
