package query

import (
	"fmt"

	"gorm.io/driver/postgres"
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
	switch regexp.Dialector {
	case postgres.Dialector{}.Name():
		// #nosec G104
		builder.WriteString(" ~ ")
	default:
		// #nosec G104
		builder.WriteString(" regexp ")
	}
	builder.AddVar(builder, fmt.Sprintf("%s", regexp.Value))
}

// NegationBuild builds negative statement.
func (regexp Regexp) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(regexp.Column)
	switch regexp.Dialector {
	case postgres.Dialector{}.Name():
		// #nosec G104
		builder.WriteString(" !~ ")
	default:
		// #nosec G104
		builder.WriteString(" NOT regexp ")
	}
	builder.AddVar(builder, regexp.Value)
}
