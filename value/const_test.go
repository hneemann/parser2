package value

import (
	"testing"
)

func TestConst(t *testing.T) {
	runTest(t, []testType{
		{exp: "const a=3; a", res: Int(3)},
		{exp: "const a=3; const b=4; a+b", res: Int(7)},
		{exp: "const a=3; const a=4; a", res: Int(4)},
		{exp: "const c=3; let f=c->c*c; f(c+1)", res: Int(16)},
		{exp: "const c=3; func f(c) c*c; f(c+1)", res: Int(16)},
		{exp: "func f(pi) pi*2; f(2)", res: Int(4)},
	})
}
