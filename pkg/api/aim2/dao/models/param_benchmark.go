package models

import (
	"math/rand"
	"strconv"
	"testing"
)

// ParamValueTyped has fields for each possible value type.
type ParamValueTyped struct {
	Key         string
	ValueString *string
	ValueInt    *int64
	ValueFloat  *float64
}

func (p ParamValueTyped) Value() any {
	if p.ValueString != nil {
		return *p.ValueString
	}
	if p.ValueInt != nil {
		return *p.ValueInt
	}
	if p.ValueFloat != nil {
		return *p.ValueFloat
	}
	return nil
}

func BenchmarkValueCast(b *testing.B) {
	// Generate 100,000 random inputs
	inputs := make([]Param, 100000)
	for i := range inputs {
		switch rand.Intn(3) {
		case 0:
			// int64
			inputs[i] = Param{Value: strconv.FormatInt(rand.Int63(), 10)}
		case 1:
			// float64
			inputs[i] = Param{Value: strconv.FormatFloat(rand.Float64()*float64(rand.Intn(100)), 'f', 6, 64)}
		case 2:
			// string
			inputs[i] = Param{Value: strconv.Itoa(rand.Int())}
		}
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			input.ValueTyped()
		}
	}
}

func BenchmarkValueTyped(b *testing.B) {
	// Generate 100,000 random inputs
	inputs := make([]ParamValueTyped, 100000)
	for i := range inputs {
		randInt := rand.Int63()
		randFloat := rand.Float64() * float64(rand.Intn(100))
		randString := strconv.Itoa(rand.Int())
		switch rand.Intn(3) {
		case 0:
			// int64
			inputs[i] = ParamValueTyped{ValueInt: &randInt}
		case 1:
			// float64
			inputs[i] = ParamValueTyped{ValueFloat: &randFloat}
		case 2:
			// string
			inputs[i] = ParamValueTyped{ValueString: &randString}
		}
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			input.Value()
		}
	}
}
