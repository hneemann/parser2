package example

import (
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMinimal(t *testing.T) {
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
		{"3^2", 9, "9"},
		{"a-1", 1, "a-1"},
		{"1+a", 3, "1+a"},
		{"4*4+a", 18, "16+a"},
		{"8/2/2", 2, "2"},
		{"(a*4)/2/2", 2, "((a*4)/2)/2"},
		{"a+4*4", 18, "a+16"},
		{"(a+4)*4", 24, "(a+4)*4"},
		{"4*(a+4)", 24, "4*(a+4)"},
		{"2*2*a", 8, "4*a"},
		{"2*a*2", 8, "4*a"},
		{"a*2*2", 8, "a*4"},
		{"2+2+a", 6, "4+a"},
		{"2+a+2", 6, "4+a"},
		{"a+2+2", 6, "a+4"},
		{"a-2-2", -2, "(a-2)-2"},
		{"a+a+2+2", 8, "(a+a)+4"},
		{"a-a-2-2", -4, "((a-a)-2)-2"},
		{"2+2+a+a", 8, "(4+a)+a"},
		{"2-2-a-a", -4, "(0-a)-a"},
		{"sin(pi/2)", 1, "1"},
		{"sin(pi/a)", 1, "sin(3.141592653589793/a)"},
		{"2a²-2a+1", 5, "((2*(a^2))-(2*a))+1"},
		{"(a+1)(1-a)", -3, "(a+1)*(1-a)"},
		{"2 a cos(a)", -1.6645873461885696, "(2*a)*cos(a)"},
		{"2 cos(a) a", -1.6645873461885696, "(2*cos(a))*a"},
		{"a 2 cos(a)", -1.6645873461885696, "(a*2)*cos(a)"},
		{"a cos(a) 2", -1.6645873461885696, "(a*cos(a))*2"},
		{"cos(a) 2 a", -1.6645873461885696, "(cos(a)*2)*a"},
		{"cos(a) a 2", -1.6645873461885696, "(cos(a)*a)*2"},
		{"a (a+1)", 6, "a*(a+1)"},
	}

	vars := funcGen.NewStack[float64](2)
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			// check result
			// create the function which evaluates the given expression
			f, err := minimal.Generate(test.exp, "a")
			assert.NoError(t, err)
			// evaluate the function using the given variables
			r, err := f(vars)
			assert.NoError(t, err, test.exp)
			assert.InDelta(t, test.result, r, 1e-6, test.exp)

			// check optimizer
			// not required in production usage
			var idents parser2.Identifiers[float64]
			idents = idents.Add("a")
			ast, err := minimal.CreateAst(test.exp, idents, false)
			assert.NoError(t, err, test.exp)
			assert.EqualValues(t, test.optimized, ast.String(), test.exp)
		})
	}
}
