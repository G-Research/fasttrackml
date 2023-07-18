package query

import (
	"fmt"

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
	builder.WriteQuoted(regexp.Column)
	operator := ""
	switch regexp.Dialector {
	case sqlite.Dialector{}.Name():
		operator = "regexp"
	case postgres.Dialector{}.Name():
		operator = "~"
	}
	builder.WriteString(fmt.Sprintf(" %s ", operator))
	builder.AddVar(builder, fmt.Sprintf("%s", regexp.Value))
}

// NegationBuild builds negative statement.
func (regexp Regexp) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(regexp.Column)
	switch regexp.Dialector {
	case sqlite.Dialector{}.Name():
		builder.WriteString(" NOT regexp ")
	case postgres.Dialector{}.Name():
		builder.WriteString(" !~ ")
	}
	builder.AddVar(builder, regexp.Value)
}
