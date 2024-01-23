package query

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	"github.com/rotisserie/eris"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/clause"
)

// Regexp whether string matches regular expression
type Regexp struct {
	clause.Eq
	Dialector string
}

// Build builds positive statement.
func (regexp Regexp) Build(builder clause.Builder) {
	regexp.writeColumn(builder)
	switch regexp.Dialector {
	case postgres.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString(" ~ ")
	default:
		//nolint:errcheck,gosec
		builder.WriteString(" REGEXP ")
	}
	builder.AddVar(builder, regexp.Value)
}

// NegationBuild builds negative statement.
func (regexp Regexp) NegationBuild(builder clause.Builder) {
	regexp.writeColumn(builder)
	switch regexp.Dialector {
	case postgres.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString(" !~ ")
	default:
		//nolint:errcheck,gosec
		builder.WriteString(" NOT REGEXP ")
	}
	builder.AddVar(builder, regexp.Value)
}

func (regexp Regexp) writeColumn(builder clause.Builder) {
	switch regexp.Dialector {
	case sqlite.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("IFNULL(")
		builder.WriteQuoted(regexp.Column)
		//nolint:errcheck,gosec
		builder.WriteString(", '')")
	default:
		builder.WriteQuoted(regexp.Column)
	}
}

// Json clause for string match at a json path.
type Json struct {
	clause.Column
	JsonPath  string
	Dialector string
}

// Build builds positive statement.
func (json Json) Build(builder clause.Builder) {
	json.writeColumn(builder)
	jsonPath := json.JsonPath
	switch json.Dialector {
	case postgres.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("#>>")
		jsonPath = "{" + strings.ReplaceAll(jsonPath, ",", ".") + "}"
	default:
		//nolint:errcheck,gosec
		builder.WriteString("->>")
	}
	builder.AddVar(builder, jsonPath)
}

// NegationBuild builds negative statement.
func (json Json) NegationBuild(builder clause.Builder) {
	json.writeColumn(builder)
	jsonPath := json.JsonPath
	switch json.Dialector {
	case postgres.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("#>>")
		jsonPath = "{" + strings.ReplaceAll(jsonPath, ",", ".") + "}"
	default:
		//nolint:errcheck,gosec
		builder.WriteString("->>")
	}
	builder.AddVar(builder, jsonPath)
}

func (json Json) writeColumn(builder clause.Builder) {
	switch json.Dialector {
	case sqlite.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("IFNULL(")
		builder.WriteQuoted(json.Column)
		//nolint:errcheck,gosec
		builder.WriteString(", JSON('{}'))")
	default:
		builder.WriteQuoted(json.Column)
	}
}

type JsonEq struct {
	Left      Json
	Value     any
	Dialector string
}

func (eq JsonEq) Build(builder clause.Builder) {
	eq.Left.Build(builder)
	switch eq.Value.(type) {
	case []JsonEq:
		rv := reflect.ValueOf(eq.Value)
		if rv.Len() == 0 {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" = ")
			//nolint:errcheck,gosec
			renderDictValue(builder, eq.Dialector, rv)
		}
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []interface{}:
		rv := reflect.ValueOf(eq.Value)
		if rv.Len() == 0 {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" = ")
			//nolint:errcheck
			renderArrayValue(builder, eq.Dialector, rv)
		}
	default:
		if eqNil(eq.Value) {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" = ")
			builder.AddVar(builder, eq.Value)
		}
	}
}

func (eq JsonEq) NegationBuild(builder clause.Builder) {
	JsonNeq(eq).Build(builder)
}

// JsonNeq not equal to for where
type JsonNeq JsonEq

func (neq JsonNeq) Build(builder clause.Builder) {
	neq.Left.Build(builder)
	switch neq.Value.(type) {
	case []JsonEq:
		rv := reflect.ValueOf(neq.Value)
		if rv.Len() == 0 {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NOT NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" <> ")
			//nolint:errcheck,gosec
			renderDictValue(builder, neq.Dialector, rv)
		}
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []interface{}:
		rv := reflect.ValueOf(neq.Value)
		if rv.Len() == 0 {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" <> ")
			renderArrayValue(builder, neq.Dialector, rv)
		}
	default:
		if eqNil(neq.Value) {
			//nolint:errcheck,gosec
			builder.WriteString(" IS NOT NULL")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" <> ")
			builder.AddVar(builder, neq.Value)
		}
	}
}

func (neq JsonNeq) NegationBuild(builder clause.Builder) {
	JsonEq(neq).Build(builder)
}

// JsonLike like for where clause.
type JsonLike struct {
	Json  Json
	Value any
}

// Build renders the Json like expression.
func (jl JsonLike) Build(builder clause.Builder) {
	jl.Json.Build(builder)
	//nolint:errcheck,gosec
	builder.WriteString(" LIKE ")
	builder.AddVar(builder, jl.Value)
}

// NegationBuild renders the Json not-like expression.
func (jl JsonLike) NegationBuild(builder clause.Builder) {
	JsonNotLike(jl).Build(builder)
}

// JsonNeq not like for where.
type JsonNotLike JsonLike

// Build renders the Json not-like expression.
func (jnl JsonNotLike) Build(builder clause.Builder) {
	jnl.Json.Build(builder)
	//nolint:errcheck,gosec
	builder.WriteString(" NOT LIKE ")
	builder.AddVar(builder, jnl.Value)
}

// NegationBuild renders the Json like expression.
func (neq JsonNotLike) NegationBuild(builder clause.Builder) {
	JsonLike(neq).Build(builder)
}

func eqNil(value interface{}) bool {
	if valuer, ok := value.(driver.Valuer); ok && !eqNilReflect(valuer) {
		//nolint:errcheck
		value, _ = valuer.Value()
	}

	return value == nil || eqNilReflect(value)
}

func eqNilReflect(value interface{}) bool {
	reflectValue := reflect.ValueOf(value)
	return reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil()
}

func renderArrayValue(builder clause.Builder, dialector string, rv reflect.Value) {
	//nolint:errcheck,gosec
	builder.WriteString("'[")
	tmpl := strings.Repeat("%v,", rv.Len()-1) + "%v"

	switch dialector {
	case postgres.Dialector{}.Name():
		tmpl = strings.ReplaceAll(tmpl, ",", ", ")
	}

	vals := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		vals[i] = rv.Index(i).Interface()
	}
	//nolint:errcheck,gosec
	builder.WriteString(fmt.Sprintf(tmpl, vals...))

	//nolint:errcheck,gosec
	builder.WriteString("]'")
}

func renderDictValue(builder clause.Builder, dialector string, rv reflect.Value) error {
	//nolint:errcheck,gosec
	builder.WriteString("'{")
	tmpl := strings.Repeat(`"%v":"%v",`, rv.Len()-1) + `"%v":"%v"`

	switch dialector {
	case postgres.Dialector{}.Name():
		tmpl = strings.ReplaceAll(tmpl, ":", ": ")
		tmpl = strings.ReplaceAll(tmpl, ",", ", ")
	}

	vals := make([]any, rv.Len()*2)
	dictIndex := 0
	for i := 0; i < rv.Len(); i++ {
		jsonEq, ok := rv.Index(i).Interface().(JsonEq)
		if !ok {
			return eris.New("unable to cast reflect value to JsonEq")
		}
		vals[dictIndex] = jsonEq.Left.JsonPath
		vals[dictIndex+1] = jsonEq.Value
		dictIndex = dictIndex + 2
	}
	//nolint:errcheck,gosec
	builder.WriteString(fmt.Sprintf(tmpl, vals...))
	//nolint:errcheck,gosec
	builder.WriteString("}'")
	return nil
}
