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
		{exp: "1e-7", res: float64(1e-7)},
		{exp: "1e7", res: float64(1e7)},
		{exp: "1e+7", res: float64(1e+7)},
		{exp: "1+2", res: float64(3)},
		{exp: "2-1", res: float64(1)},
		{exp: "1<2", res: true},
		{exp: "1>2", res: false},
		{exp: "1>2", res: false},
		{exp: "1<2", res: true},
		{exp: "2=2", res: true},
		{exp: "1=2", res: false},
		{exp: "2!=2", res: false},
		{exp: "1!=2", res: true},
		{exp: "2>=2", res: true},
		{exp: "2<=2", res: true},
		{exp: "\"a\"=\"a\"", res: true},
		{exp: "\"a\">=\"a\"", res: true},
		{exp: "\"a\"<=\"a\"", res: true},
		{exp: "\"a\"=\"b\"", res: false},
		{exp: "\"a\">\"b\"", res: false},
		{exp: "\"a\"<\"b\"", res: true},
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
		{exp: "let fib=n->if n<=2 then 1 else fib(n-1)+fib(n-2);[fib(10),fib(15)]", res: vList{vFloat(55), vFloat(610)}},
		{exp: "if 1<2 then 1 else 2", res: vFloat(1)},
		{exp: "if 1>2 then 1 else 2", res: vFloat(2)},
		{exp: "if 1<2 then 1 else notAvail", res: vFloat(1)},
		{exp: "if 1>2 then notAvail else 2", res: vFloat(2)},
		{exp: "let a=2; if 1<a then 1 else 2", res: vFloat(1)},
		{exp: "let a=2; if 1>a then 1 else 2", res: vFloat(2)},
		{exp: "let a=2; if 1<a then 1 else notAvail", res: vFloat(1)},
		{exp: "let a=2; if 1>a then notAvail else 2", res: vFloat(2)},
		{exp: "true | (notAvail<1)", res: vBool(true)},
		{exp: "false & (notAvail<1)", res: vBool(false)},
		{exp: "[1,2,3].size()", res: vFloat(3)},
		{exp: "[1,2,3].map(e->e*2)", res: vList{vFloat(2), vFloat(4), vFloat(6)}},
		{exp: "[1,2,3,4,5].reduce((a,b)->a+b)", res: vFloat(15)},
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
			if c, ok := ast.(*parser2.Const[Value]); ok {
				assert.EqualValues(t, test.res, c.Value)
			} else {
				t.Errorf("not a constant: %v -> %v", test.exp, ast)
			}
		})
	}
}

// The power of closures and recursion.
// Recursive implementation of the sqrt function using the Regula-Falsi algorithm.
const regulaFalsi = `
      let regulaFalsi = rf->
          let xn = (rf.x0*rf.f1 - rf.x1*rf.f0) / (rf.f1 - rf.f0);
          let fn = rf.f(xn);

          let next = if abs(rf.f0) > abs(rf.f1)
                       then {x0:xn, f0:fn, x1:rf.x1, f1:rf.f1, f:rf.f}
                       else {x0:rf.x0, f0:rf.f0, x1:xn, f1:fn, f:rf.f};

          if abs(fn)<1e-7 
            then next 
            else regulaFalsi(next);

      let solve = (x0, x1, f)->
          let r = regulaFalsi({x0:x0, f0:f(x0), x1:x1, f1:f(x1), f:f});
          if abs(r.f0)<abs(r.f1) 
            then r.x0 
            else r.x1;

      let mySqrt = a->solve(1, 2, x->x*x-a);

      mySqrt(a)
    `

// Recursive implementation of the sqrt function using the Newton-Raphson algorithm.
// Since the first derivative is required, no solver for arbitrary functions can be implemented.
const newtonRaphson = `
      let newton = (x,a) -> if abs(x*x-a)<1e-7 
                              then x 
                              else newton(x+(a-x*x)/(2*x), a);
      let mySqrt = a -> newton(2,a); 

      mySqrt(a)
    `

func TestSolve(t *testing.T) {
	tests := []struct {
		name string
		exp  string
	}{
		{name: "regulaFalsi", exp: regulaFalsi},
		{name: "newtonRaphson", exp: newtonRaphson},
	}

	v := parser2.VarMap[Value]{"a": vFloat(2)}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			f, err := DynType.Generate(test.exp)
			assert.NoError(t, err, test.name)
			if f != nil {
				r, err := f(v)
				assert.NoError(t, err, test.name)
				assert.InDelta(t, math.Sqrt(2), r.Float(), 1e-6, test.name)
			}
		})
	}
}

func BenchmarkRegulaFalsi(b *testing.B) {
	f, _ := DynType.Generate(regulaFalsi)
	v := parser2.VarMap[Value]{"a": vFloat(2)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(v)
	}
}

func BenchmarkCall(b *testing.B) {
	f, _ := DynType.Generate("x+(2*y/x)")
	v := parser2.VarMap[Value]{"x": vFloat(3), "y": vFloat(3)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(v)
	}
}

func BenchmarkFunc(b *testing.B) {
	f, _ := DynType.Generate("let f=x->x*x;f(a)+f(2*a)")
	v := parser2.VarMap[Value]{"a": vFloat(3)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(v)
	}
}
