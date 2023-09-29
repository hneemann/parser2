package funcGen

import (
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

type Value interface {
	Float() float64
}

type Float float64

func (f Float) Float() float64 {
	return float64(f)
}

func (f Float) String() string {
	return fmt.Sprintf("%f", float64(f))
}

type vClosure Function[Value]

func (v vClosure) Float() float64 {
	panic("a closure is not a float value")
}

func (v vClosure) String() string {
	return "closure"
}

type typeHandler struct{}

var th typeHandler

func (th typeHandler) FromClosure(closure Function[Value]) Value {
	return vClosure(closure)
}

func (th typeHandler) ToClosure(fu Value) (Function[Value], bool) {
	cl, ok := fu.(vClosure)
	return Function[Value](cl), ok
}

func NewGen() *FunctionGenerator[Value] {
	return New[Value]().
		AddOp("+", true, func(a Value, b Value) Value { return Float(a.Float() + b.Float()) }).
		AddOp("*", true, func(a Value, b Value) Value { return Float(a.Float() * b.Float()) }).
		SetClosureHandler(th).
		SetNumberParser(
			parser2.NumberParserFunc[Value](
				func(n string) (Value, error) {
					f, err := strconv.ParseFloat(n, 64)
					return Float(f), err
				},
			),
		)
}

func TestFunctionGenerator_Generate(t *testing.T) {
	fg := NewGen()

	tests := []struct {
		args     []string
		exp      string
		argsVals []Value
		result   float64
	}{
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(2), Float(3)},
			exp:      "a*(b+2)",
			result:   10,
		},
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(3), Float(2)},
			exp:      "a*(b+2)",
			result:   12,
		},
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(3), Float(2)},
			exp:      "let c=2;a*(b+c)",
			result:   12,
		},
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(3), Float(2)},
			exp:      "let sqr=x->x*x;a*sqr(b)",
			result:   12,
		},
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(3), Float(2)},
			exp:      "func mul(x) y->y*x;mul(b)(a)",
			result:   6,
		},
		{
			args:     []string{"a", "b"},
			argsVals: []Value{Float(3), Float(2)},
			exp:      "let c=1.5;func mul(x) y->y*x*c;mul(b)(a)",
			result:   9,
		},
	}

	for _, te := range tests {
		test := te
		t.Run(test.exp, func(t *testing.T) {
			f, err := fg.Generate(test.exp, test.args...)
			assert.NoError(t, err)
			assert.NotNil(t, f)
			if f != nil {
				res, err := f(NewStack(test.argsVals...))
				assert.NoError(t, err)
				if res != nil {
					assert.InDelta(t, test.result, res.Float(), 1e-6)
				}
			}
		})
	}
}

func BenchmarkFunc(b *testing.B) {
	f, _ := NewGen().Generate("func f(x) x*x;f(a)+f(2*a)", "a")
	argVals := []Value{Float(2)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(NewStack(argVals...))
	}
}

func BenchmarkFunc2(b *testing.B) {
	f, _ := NewGen().Generate("let c=1.5;func mul(x) y->y*x*c;mul(b)(a)", "a", "b")
	argVals := []Value{Float(3), Float(2)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(NewStack(argVals...))
	}
}
