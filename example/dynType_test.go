package example

import (
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynType(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "let a=1;sprintf()", res: value.String("")},
		{exp: "let a=1;sprintf(\"Hello World\")", res: value.String("Hello World")},
		{exp: "let a=1;sprintf(\"%v->%v\",a,2)", res: value.String("1->2")},
		{exp: "let a=1;sprintf(\"%v->\",a)", res: value.String("1->")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			fu, err := DynTypeParser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)
				if _, ok := test.res.(float64); ok {
					float, ok := res.(value.Float)
					assert.True(t, ok)
					assert.InDelta(t, test.res, float64(float), 1e-6, test.exp)
				} else if expList, ok := test.res.(value.List); ok {
					actList, ok := res.(*value.List)
					assert.True(t, ok)
					assert.EqualValues(t, expList.ToSlice(), actList.ToSlice(), test.exp)
				} else {
					assert.EqualValues(t, test.res, res, test.exp)
				}
			}
		})
	}
}

func TestOptimizer(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "sprintf(\"%v->%v\",1,2)", res: value.String("1->2")},
		{exp: "sprintf(\"%v->\",1)", res: value.String("1->")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			ast, err := DynTypeParser.CreateAst(test.exp)
			assert.NoError(t, err, test.exp)
			if c, ok := ast.(*parser2.Const[value.Value]); ok {
				if f, ok := test.res.(value.Float); ok {
					fl, ok := c.Value.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(f), fl, 1e-7)
				} else if expList, ok := test.res.(value.List); ok {
					actList, ok := c.Value.(*value.List)
					assert.True(t, ok)
					assert.EqualValues(t, expList.ToSlice(), actList.ToSlice(), test.exp)
				} else {
					assert.EqualValues(t, test.res, c.Value)
				}
			} else {
				t.Errorf("not a constant: %v -> %v", test.exp, ast)
			}
		})
	}
}
