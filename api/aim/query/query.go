package query

import (
	"fmt"
	"reflect"
	"time"

	"github.com/G-Resarch/fasttrack/database"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueryParser struct {
	Tables   map[string]Table
	TzOffset int
}

type parsedQuery struct {
	qp         *QueryParser
	joins      map[string]string
	columns    map[string]clause.Column
	conditions []clause.Expression
}
type ParsedQuery interface {
	Filter(*gorm.DB) *gorm.DB
}

type Table map[string]any

type function func(pq *parsedQuery, args ...ast.Expr) (any, error)

var functions map[ast.Identifier]function

func (qp *QueryParser) Parse(q string) (ParsedQuery, error) {
	pq := &parsedQuery{
		qp: qp,
	}

	if q == "" {
		return pq, nil
	}

	a, err := parser.ParseString(q, py.EvalMode)
	if err != nil {
		return nil, err
	}

	e, ok := a.(*ast.Expression)
	if !ok {
		return nil, fmt.Errorf("not a valid Python expression: %#v", a)
	}

	// TODO this is just for debugging
	fmt.Println(ast.Dump(e))

	cl, err := pq.parseNode(e.Body)
	if err != nil {
		return nil, err
	}

	cond, ok := cl.(clause.Expression)
	if !ok {
		return nil, fmt.Errorf("not a valid SQL expression: %#v", cl)
	}

	pq.conditions = append(pq.conditions, cond)

	return pq, nil
}

func (pq *parsedQuery) Filter(tx *gorm.DB) *gorm.DB {
	for _, c := range pq.columns {
		tx.Select(c)
	}
	for _, j := range pq.joins {
		tx.Joins(j)
	}
	if len(pq.conditions) > 0 {
		tx.Where(clause.And(pq.conditions...))
	}
	return tx
}

func (pq *parsedQuery) parseNode(node ast.Expr) (any, error) {
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
	case *ast.UnaryOp:
		return pq.parseUnaryOp(n)
	case *ast.Attribute:
		v, err := pq.parseNode(n.Value)
		if err != nil {
			return nil, err
		}
		a := string(n.Attr)
		switch v := v.(type) {
		case Table:
			c, ok := v[a]
			if ok {
				return c, nil
			}
			c, ok = v["*"]
			if ok {
				return c, nil
			}
			return nil, fmt.Errorf("no mapping for attribute %q in %q", a, n.Value.(*ast.Name).Id)
		default:
			return nil, fmt.Errorf("unsupported attribute value %#v", v)
		}
		// c := string(n.Attr)
		// switch c {
		// case "created_at":
		// 	c = "start_time"
		// case "finalized_at":
		// 	c = "end_time"
		// case "hash":
		// 	c = "run_uuid"
		// case "name":
		// case "experiment":
		// 	t = "Experiment"
		// 	c = "name"
		// case "active":
		// 	return clause.Eq{
		// 		Column: clause.Column{
		// 			Table: t,
		// 			Name:  "status",
		// 		},
		// 		Value: database.StatusRunning,
		// 	}, nil
		// case "archived":
		// 	return clause.Eq{
		// 		Column: clause.Column{
		// 			Table: t,
		// 			Name:  "lifecycle_stage",
		// 		},
		// 		Value: database.LifecycleStageDeleted,
		// 	}, nil
		// case "duration":
		// 	return clause.Column{
		// 		Name: "runs.end_time - runs.start_time",
		// 		Raw:  true,
		// 	}, nil
		// case "metrics":
		// default:
		// 	return clause.And(
		// 		clause.Eq{
		// 			Column: clause.Column{
		// 				Table: "Params",
		// 				Name:  "key",
		// 			},
		// 			Value: c,
		// 		},
		// 		clause.Eq{
		// 			Column: clause.Column{
		// 				Table: "Params",
		// 				Name:  "value",
		// 			},
		// 			Value: "1",
		// 		},
		// 	), nil
		// }
		// return clause.Column{
		// 	Table: t,
		// 	Name:  c,
		// }, nil
	case *ast.Compare:
		exprs := make([]clause.Expression, len(n.Ops))
		for i, o := range n.Ops {
			// lrOrder := true
			// var equality clause.Expression
			leftAst := n.Left
			if i > 0 {
				leftAst = n.Comparators[i-1]
			}
			left, err := pq.parseNode(leftAst)
			if err != nil {
				return nil, err
			}
			right, err := pq.parseNode(n.Comparators[i])
			if err != nil {
				return nil, err
			}
			if reflect.TypeOf(left) != reflect.TypeOf(clause.Column{}) {
				if reflect.TypeOf(right) == reflect.TypeOf(clause.Column{}) {
					o = reverseComparison(o)
					t := left
					left = right
					right = t
				} else {
					left = clause.Column{
						Name: database.DB.Statement.Quote(left),
						Raw:  true,
					}
				}
			}
			// if reflect.TypeOf(left) == reflect.TypeOf(clause.Eq{}) {
			// 	if reflect.TypeOf(right) == reflect.TypeOf(true) {
			// 		equality = left.(clause.Eq)
			// 		if !right.(bool) {
			// 			equality = clause.Not(equality)
			// 		}
			// 	} else {
			// 		return nil, fmt.Errorf("unsupported comparison %q", ast.Dump(n))
			// 	}
			// }
			// if reflect.TypeOf(right) == reflect.TypeOf(clause.Eq{}) {
			// 	if reflect.TypeOf(left) == reflect.TypeOf(true) {
			// 		equality = right.(clause.Eq)
			// 		if !left.(bool) {
			// 			equality = clause.Not(equality)
			// 		}
			// 	} else {
			// 		return nil, fmt.Errorf("unsupported comparison %q", ast.Dump(n))
			// 	}
			// }
			// if equality != nil {
			// 	switch o {
			// 	case ast.Eq, ast.Is:
			// 		exprs[i] = equality
			// 	case ast.NotEq, ast.IsNot:
			// 		exprs[i] = clause.Not(equality)
			// 	default:
			// 		return nil, fmt.Errorf("unsupported comparison %q", ast.Dump(n))
			// 	}
			// 	break
			// }
			switch o {
			case ast.Eq, ast.Is:
				exprs[i] = clause.Eq{
					Column: left,
					Value:  right,
				}
			case ast.NotEq, ast.IsNot:
				exprs[i] = clause.Neq{
					Column: left,
					Value:  right,
				}
			case ast.Lt:
				exprs[i] = clause.Lt{
					Column: left,
					Value:  right,
				}
			case ast.LtE:
				exprs[i] = clause.Lte{
					Column: left,
					Value:  right,
				}
			case ast.Gt:
				exprs[i] = clause.Gt{
					Column: left,
					Value:  right,
				}
			case ast.GtE:
				exprs[i] = clause.Gte{
					Column: left,
					Value:  right,
				}
			case ast.In:
				r, ok := right.([]any)
				if !ok {
					return nil, fmt.Errorf("right value in \"in\" comparison is not a list: %#v", right)
				}
				exprs[i] = clause.IN{
					Column: left,
					Values: r,
				}
			case ast.NotIn:
				r, ok := right.([]any)
				if !ok {
					return nil, fmt.Errorf("right value in \"not in\" comparison is not a list: %#v", right)
				}
				exprs[i] = clause.NotConditions{
					Exprs: []clause.Expression{
						clause.IN{
							Column: left,
							Values: r,
						},
					},
				}
			default:
				return nil, fmt.Errorf("unsupported compare operation %s", o)
			}
		}
		return clause.AndConditions{
			Exprs: exprs,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported expression %#v", n)
	}
}

func reverseComparison(op ast.CmpOp) ast.CmpOp {
	switch op {
	case ast.Lt:
		return ast.Gt
	case ast.LtE:
		return ast.GtE
	case ast.Gt:
		return ast.Lt
	case ast.GtE:
		return ast.LtE
	default:
		return op
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
		return nil, fmt.Errorf("unsupported boolean operation %s", node.Op)
	}
}

func (pq *parsedQuery) parseCall(node *ast.Call) (any, error) {
	f, err := pq.parseNode(node.Func)
	if err != nil {
		return nil, err
	}

	fu, ok := f.(function)
	if !ok {
		return nil, fmt.Errorf("unsupported call to function %#v", node.Func)
	}

	return fu(pq, node.Args...)
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
		t, ok := pq.qp.Tables[string(node.Id)]
		if ok {
			return t, nil
		}
		f, ok := functions[node.Id]
		if ok {
			return f, nil
		}
		return nil, fmt.Errorf("unsupported name identifier %q", node.Id)
	default:
		return nil, fmt.Errorf("unsupported name context %s", node.Ctx)
	}
}

func (pq *parsedQuery) parseNameConstant(node *ast.NameConstant) (any, error) {
	switch node.Value.Type() {
	case py.NoneTypeType:
		return nil, nil
	case py.BoolType:
		return bool(node.Value.(py.Bool)), nil
	default:
		return nil, fmt.Errorf("unsupported name constant type %s", node.Value.Type())
	}
}
func (pq *parsedQuery) parseNum(node *ast.Num) (any, error) {
	switch node.N.Type() {
	case py.IntType:
		return node.N.(py.Int).GoInt()
	case py.FloatType:
		return py.FloatAsFloat64(node.N.(py.Float))
	default:
		return nil, fmt.Errorf("unsupported num type %s", node.N.Type())
	}
}

func (pq *parsedQuery) parseStr(node *ast.Str) (any, error) {
	return string(node.S), nil
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
		return nil, fmt.Errorf("unsupported unary operation %s", node.Op)
	}
}

func initFunctions() {
	functions = map[ast.Identifier]function{
		"datetime": func(pq *parsedQuery, args ...ast.Expr) (any, error) {
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
				time.FixedZone("custom", pq.qp.TzOffset*60),
			).UnixMilli(), nil
		},
	}
}

func init() {
	initFunctions()
}
