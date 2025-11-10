package funcGen

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/stretchr/testify/assert"
	"math"
	"strconv"
	"strings"
	"testing"
)

type Value interface {
	Float() (float64, error)
}

type Float float64

func (f Float) Float() (float64, error) {
	return float64(f), nil
}

func (f Float) Sqrt() Float {
	return Float(math.Sqrt(float64(f)))
}

func (f Float) String() string {
	return fmt.Sprintf("%f", float64(f))
}

type vClosure Function[Value]

func (v vClosure) Float() (float64, error) {
	return 0, errors.New("a function is not a float value")
}

func (v vClosure) String() string {
	return "function"
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
		AddOp("+", true, func(st Stack[Value], a Value, b Value) (Value, error) {
			aVal, err := a.Float()
			if err != nil {
				return nil, err
			}
			bVal, err := b.Float()
			if err != nil {
				return nil, err
			}
			return Float(aVal + bVal), nil
		}).
		AddOp("*", true, func(st Stack[Value], a Value, b Value) (Value, error) {
			aVal, err := a.Float()
			if err != nil {
				return nil, err
			}
			bVal, err := b.Float()
			if err != nil {
				return nil, err
			}
			return Float(aVal * bVal), nil
		}).
		AddUnary("-", UnaryOperatorFunc[Value](func(a Value) (Value, error) {
			fl, err := a.Float()
			if err != nil {
				return nil, err
			}
			return Float(-fl), nil
		})).
		SetToBool(func(c Value) (bool, bool) {
			fl, err := c.Float()
			if err != nil {
				return false, false
			}
			return fl != 0, true
		}).
		SetClosureHandler(th).
		SetKeyWords("let", "const", "if", "then", "else", "try", "catch", "func").
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
		{
			args:     []string{"a"},
			argsVals: []Value{Float(3)},
			exp:      "a*(3*3)",
			result:   27,
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(3)},
			exp:      "3*a*3",
			result:   27,
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(3)},
			exp:      "a*3*3",
			result:   27,
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(3)},
			exp:      "let c=3; a*(-c)",
			result:   -9,
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(2)},
			exp:      "a.sqrt()",
			result:   math.Sqrt(2),
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(2)},
			exp:      "try a+(a->a) catch 1",
			result:   1,
		},
		{
			args:     []string{"a"},
			argsVals: []Value{Float(2)},
			exp:      "try a catch 1",
			result:   2,
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
					fl, err := res.Float()
					assert.NoError(t, err)
					assert.InDelta(t, test.result, fl, 1e-6)
				}
			}
		})
	}
}

func TestReflectionError(t *testing.T) {
	f, err := NewGen().Generate("a.doesNotExist()", "a")
	assert.NoError(t, err)
	_, err = f(NewStack[Value](Float(2)))
	assert.Error(t, err)
	errStr := err.Error()
	assert.True(t, strings.Contains(errStr, "method DoesNotExist not found"))
	assert.True(t, strings.Contains(errStr, "available are: Sqrt()"))
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

func TestFunctionDescription_String(t *testing.T) {
	tests := []struct {
		name string
		fu   Function[int]
		want string
	}{
		{
			name: "map",
			fu:   Function[int]{Args: 2}.SetMethodDescription("func([item])", "Converts  a  \nlist to a new list. The new list items are created by calling the function with the old item as argument."),
			want: "map(func([item]))\n\tConverts a list to a new list. The new list items are created by\n\tcalling the function with the old item as argument.",
		},
		{
			name: "z",
			fu:   Function[int]{Args: 2}.SetMethodDescription("func([item])", "A a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a."),
			want: "z(func([item]))\n\tA a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a\n\ta a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a\n\ta.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			tt.fu.Description.WriteTo(&b, tt.name)
			assert.Equalf(t, tt.want, b.String(), "ToString(%v)", tt.name)
		})
	}
}
