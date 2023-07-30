package example

import (
	"github.com/hneemann/parser2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	tests := []struct {
		exp       string
		result    float64
		optimized string
	}{
		{"-2", -2, "-2"},
		{"-(2+1)", -3, "-3"},
		{"1+2", 3, "3"},
		{"4*4+2", 18, "18"},
		{"2+4*4", 18, "18"},
		{"(2+4)*4", 24, "24"},
		{"4*(2+4)", 24, "24"},
		{"1+a", 3, "1+a"},
		{"4*4+a", 18, "16+a"},
		{"a+4*4", 18, "a+16"},
		{"(a+4)*4", 24, "(a+4)*4"},
		{"4*(a+4)", 24, "4*(a+4)"},
		{"2*2*a", 8, "4*a"},
		{"sin(pi/2)", 1, "1"},
		{"sin(pi/a)", 1, "sin(3.141592653589793/a)"},
	}

	vars := parser2.VarMap[float64]{"a": 2}
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			// check result
			// create the function which evaluates the given expression
			f, err := minimal.Generate(test.exp)
			assert.NoError(t, err)
			// evaluate the function using the given variables
			r, err := f(vars)
			assert.NoError(t, err)
			assert.InDelta(t, test.result, r, 1e-6)

			// check optimizer
			// not required in production usage
			ast, err := minimal.CreateAst(test.exp)
			assert.NoError(t, err)
			assert.EqualValues(t, test.optimized, ast.String())
		})
	}
}
