package example

import (
	"github.com/hneemann/parser2"
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
		{"a& !b", true, "a&!b"},
	}

	vars := parser2.Variables[bool]{"a": true, "b": false}
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			//check result
			f, err := boolParser.Generate(test.exp)
			assert.NoError(t, err)
			r, err := f(vars)
			assert.NoError(t, err)
			assert.Equal(t, test.result, r)

			// check optimizer
			ast, err := boolParser.CreateAst(test.exp)
			assert.NoError(t, err)
			assert.EqualValues(t, test.optimizes, ast.String())
		})
	}
}
