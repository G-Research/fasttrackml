package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueAny(t *testing.T) {
	tests := []struct {
		name  string
		param Param
		want  any
	}{
		{
			name:  "IntegerValue",
			param: Param{Value: "123"},
			want:  int64(123),
		},
		{
			name:  "FloatValue",
			param: Param{Value: "123.45"},
			want:  float64(123.45),
		},
		{
			name:  "StringValue",
			param: Param{Value: "abc"},
			want:  "abc",
		},
		{
			name:  "StringValue2",
			param: Param{Value: "123.45n"},
			want:  "123.45n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueTyped := tt.param.ValueAny()
			assert.Equal(t, tt.want, valueTyped)
			assert.IsType(t, tt.want, valueTyped)
		})
	}
}
