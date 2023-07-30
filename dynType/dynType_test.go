package dynType

import (
	"github.com/hneemann/parser2"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func Test(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "1+2", res: float64(3)},
		{exp: "1<2", res: true},
		{exp: "\"test\">\"hello\"", res: true},
		{exp: "\"test\"+\"hello\"", res: vString("testhello")},
		{exp: "sqrt(2)", res: math.Sqrt(2)},
		{exp: "let x=2;sqrt(x)", res: math.Sqrt(2)},
		{exp: "{a:1,b:2,c:3}", res: vMap{"a": vFloat(1), "b": vFloat(2), "c": vFloat(3)}},
		{exp: "{a:1,b:2,c:3}.b", res: vFloat(2)},
		{exp: "[1,2,3]", res: vList{vFloat(1), vFloat(2), vFloat(3)}},
		{exp: "let a=2; [1,a,3]", res: vList{vFloat(1), vFloat(2), vFloat(3)}},
		{exp: "[1,2,3][2]", res: 3},
		{exp: "let a=1;a", res: 1},
		{exp: "let sqr=x->x*x;sqr(2)", res: 4},
		{exp: "let s=3; let f=x->x*x*s;f(2)", res: 12},
		{exp: "ite(1<2,1,2)", res: vFloat(1)},
		{exp: "ite(1>2,1,2)", res: vFloat(2)},
		{exp: "ite(1<2,1,notAvail)", res: vFloat(1)},
		{exp: "ite(1>2,notAvail,2)", res: vFloat(2)},
		{exp: "true | (notAvail<1)", res: vBool(true)},
		{exp: "false & (notAvail<1)", res: vBool(false)},
		{exp: "[1,2,3].size()", res: vFloat(3)},
		{exp: "[1,2,3].map(e->e*2)", res: vList{vFloat(2), vFloat(4), vFloat(6)}},
		{exp: "let lp=lowPass(-1/ln(0.5)); [lp(0,0),lp(1,1),lp(2,1),lp(3,1)]", res: vList{vFloat(0), vFloat(0.5), vFloat(0.75), vFloat(0.875)}},
		{exp: "let a=1;sprintf(\"%v->%v\",a,2)", res: vString("1->2")},
		{exp: "let a=1;sprintf(\"%v->\",a)", res: vString("1->")},
		{exp: "{a:x->x*2,b:x->x*3}.b(4)", res: vFloat(12)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			fu, err := DynType.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(parser2.VarMap[Value]{})
				assert.NoError(t, err, test.exp)
				assert.EqualValues(t, test.res, res, test.exp)
			}
		})
	}
}

func TestOptimizer(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "1+2", res: float64(3)},
		{exp: "\"test\"+\"hello\"", res: vString("testhello")},
		{exp: "[1+2,8/4]", res: vList{vFloat(3), vFloat(2)}},
		{exp: "{a:1+2,b:8/4}", res: vMap{"a": vFloat(3), "b": vFloat(2)}},
		{exp: "(1+pi)/(pi+1)", res: vFloat(1)},
		{exp: "sqrt(4/2)", res: vFloat(math.Sqrt(2))},
		{exp: "(1<2) & (2<3)", res: vBool(true)},
		{exp: "-2/(-1)", res: vFloat(2)},
		{exp: "sprintf(\"%v->%v\",1,2)", res: vString("1->2")},
		{exp: "sprintf(\"%v->\",1)", res: vString("1->")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			ast, err := DynType.CreateAst(test.exp)
			assert.NoError(t, err, test.exp)
			if c, ok := ast.(parser2.Const[Value]); ok {
				assert.EqualValues(t, test.res, c.Value)
			} else {
				t.Errorf("not a constant: %v -> %v", test.exp, ast)
			}
		})
	}
}
