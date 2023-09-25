package example

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBool(t *testing.T) {
	tests := []struct {
		exp       string
		result    bool
		optimizes string
	}{
		{"true", true, "true"},
		{"true | false", true, "true"},
		{"true & false", false, "false"},
		{"a|b", true, "a|b"},
		{"a&b", false, "a&b"},
		{"a & !b", true, "a&!b"},
		{"let c=true; if c then a&b else a|b", false, "let c=true; if c then a&b else a|b"},
	}

	vars := []bool{true, false}
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			//check result
			// create the function which evaluates the given expression
			f, err := boolParser.Generate([]string{"a", "b"}, test.exp)
			assert.NoError(t, err)
			// evaluate the function using the given variables
			r, err := f(vars)
			assert.NoError(t, err)
			assert.Equal(t, test.result, r)

			// check optimizer
			// not required in production usage
			ast, err := boolParser.CreateAst(test.exp)
			assert.NoError(t, err)
			assert.EqualValues(t, test.optimizes, ast.String())
		})
	}
}
