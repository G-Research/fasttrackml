package query

import (
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
		// #nosec G104
		builder.WriteString(" ~ ")
	default:
		// #nosec G104
		builder.WriteString(" REGEXP ")
	}
	builder.AddVar(builder, regexp.Value)
}

// NegationBuild builds negative statement.
func (regexp Regexp) NegationBuild(builder clause.Builder) {
	regexp.writeColumn(builder)
	switch regexp.Dialector {
	case postgres.Dialector{}.Name():
		// #nosec G104
		builder.WriteString(" !~ ")
	default:
		// #nosec G104
		builder.WriteString(" NOT REGEXP ")
	}
	builder.AddVar(builder, regexp.Value)
}

func (regexp Regexp) writeColumn(builder clause.Builder) {
	switch regexp.Dialector {
	case sqlite.Dialector{}.Name():
		// #nosec G104
		builder.WriteString("IFNULL(")
		builder.WriteQuoted(regexp.Column)
		// #nosec G104
		builder.WriteString(", '')")
	default:
		builder.WriteQuoted(regexp.Column)
	}
}
