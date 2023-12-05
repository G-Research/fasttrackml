package query

import (
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

// Json clause for string match at a json path
type Json struct {
	clause.Eq
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
		builder.WriteString("#>>?")
		jsonPath = "{" + strings.ReplaceAll(jsonPath, ",", ".") + "}"
	default:
		//nolint:errcheck,gosec
		builder.WriteString("->>?")
	}
	//nolint:errcheck,gosec
	builder.WriteString(" = ")
	builder.AddVar(builder, jsonPath)
	builder.AddVar(builder, json.Value)
}

// NegationBuild builds negative statement.
func (json Json) NegationBuild(builder clause.Builder) {
	json.writeColumn(builder)
	jsonPath := json.JsonPath
	switch json.Dialector {
	case postgres.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("#>>?")
		jsonPath = "{" + strings.ReplaceAll(jsonPath, ",", ".") + "}"
	default:
		//nolint:errcheck,gosec
		builder.WriteString("->>?")
	}
	//nolint:errcheck,gosec
	builder.WriteString(" <> ")
	builder.AddVar(builder, jsonPath)
	builder.AddVar(builder, json.Value)
}

func (json Json) writeColumn(builder clause.Builder) {
	switch json.Dialector {
	case sqlite.Dialector{}.Name():
		//nolint:errcheck,gosec
		builder.WriteString("IFNULL(")
		builder.WriteQuoted(json.Column)
		//nolint:errcheck,gosec
		builder.WriteString(", '{}'::json)")
	default:
		builder.WriteQuoted(json.Column)
	}
}
