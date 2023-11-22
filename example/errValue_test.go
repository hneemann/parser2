package example

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrValue(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "2+3", res: value.Int(5)},
		{exp: "2-3", res: value.Int(-1)},
		{exp: "2*3", res: value.Int(6)},
		{exp: "6/2", res: value.Float(3)},
		{exp: "10+err(2)", res: ErrValue{10, 2}},
		{exp: "10+err(-2)", res: ErrValue{10, 2}},
		{exp: "(10+err(2))+2", res: ErrValue{12, 2}},
		{exp: "2+(10+err(2))", res: ErrValue{12, 2}},
		{exp: "(10+err(2))-2", res: ErrValue{8, 2}},
		{exp: "2-(10+err(2))", res: ErrValue{-8, 2}},
		{exp: "(10+err(2))*2", res: ErrValue{20, 4}},
		{exp: "2*(10+err(2))", res: ErrValue{20, 4}},
		{exp: "(10+err(2))/2", res: ErrValue{5, 1}},
		{exp: "10/(2+err(1))", res: ErrValue{5, 5}},
		{exp: "(10+err(2))-(11+err(3))", res: ErrValue{-1, 5}},
		{exp: "(10+err(2))+(11+err(3))", res: ErrValue{21, 5}},
		{exp: "(10+err(2))*(11+err(3))", res: ErrValue{110, 58}},
		{exp: "(100+err(8))/(10+err(1))", res: ErrValue{10, 2}},
		{exp: "(10+err(2)).val()", res: value.Float(10)},
		{exp: "(10+err(2)).err()", res: value.Float(2)},
		{exp: `let a=10+err(1);
               let b=20+err(1);
               let c=30+err(1);
               (a+b)/c`, res: ErrValue{val: 1, err: 32.0/29.0 - 1}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			fu, err := ErrValueParser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)
				switch expected := test.res.(type) {
				case value.Float:
					f, ok := res.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(expected), f, 1e-6, test.exp)
				case ErrValue:
					e, ok := res.(ErrValue)
					assert.True(t, ok)
					assert.InDelta(t, expected.val, e.val, 1e-6, test.exp)
					assert.InDelta(t, expected.err, e.err, 1e-6, test.exp)
				default:
					assert.Equal(t, test.res, res, test.exp)
				}
			}
		})
	}
}
