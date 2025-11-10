package value

import (
	"testing"
)

func TestConst(t *testing.T) {
	runTest(t, []testType{
		{exp: "let a=3; a", res: Int(3)},
		{exp: "let a=3; let b=4; a+b", res: Int(7)},
		{exp: "let c=3; let f=c->c*c; f(c+1)", res: Int(16)},
		{exp: "let c=3; func f(c) c*c; f(c+1)", res: Int(16)},
		{exp: "func f(pi) pi*2; f(2)", res: Int(4)},
	})
}
