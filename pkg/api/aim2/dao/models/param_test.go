package models

import (
	"testing"

	"github.com/G-Research/fasttrackml/pkg/common"
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
			param: Param{ValueInt: common.GetPointer[int64](int64(123))},
			want:  int64(123),
		},
		{
			name:  "FloatValue",
			param: Param{ValueFloat: common.GetPointer[float64](float64(123.45))},
			want:  float64(123.45),
		},
		{
			name:  "StringValue",
			param: Param{ValueStr: common.GetPointer[string]("abc")},
			want:  "abc",
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
