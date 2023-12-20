package query

import (
	"database/sql/driver"
	"reflect"
	"strings"

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
	Left  Json
	Value any
}

func (eq JsonEq) Build(builder clause.Builder) {
	eq.Left.Build(builder)
	switch eq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []interface{}:
		rv := reflect.ValueOf(eq.Value)
		if rv.Len() == 0 {
			//nolint:errcheck,gosec
			builder.WriteString(" IN (NULL)")
		} else {
			//nolint:errcheck,gosec
			builder.WriteString(" IN (")
			for i := 0; i < rv.Len(); i++ {
				if i > 0 {
					//nolint:errcheck,gosec
					builder.WriteByte(',')
				}
				builder.AddVar(builder, rv.Index(i).Interface())
			}
			//nolint:errcheck,gosec
			builder.WriteByte(')')
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
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []interface{}:
		//nolint:errcheck,gosec
		builder.WriteString(" NOT IN (")
		rv := reflect.ValueOf(neq.Value)
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				//nolint:errcheck,gosec
				builder.WriteByte(',')
			}
			builder.AddVar(builder, rv.Index(i).Interface())
		}
		//nolint:errcheck,gosec
		builder.WriteByte(')')
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
