package example

import (
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestDynType(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "1e-7", res: float64(1e-7)},
		{exp: "1e7", res: float64(1e7)},
		{exp: "1e+7", res: float64(1e+7)},
		{exp: "1+2", res: 3},
		{exp: "2-1", res: 1},
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
		{exp: "\"test\"+\"hello\"", res: value.String("testhello")},
		{exp: "sqrt(2)", res: math.Sqrt(2)},
		{exp: "let x=2;sqrt(x)", res: math.Sqrt(2)},
		{exp: "{a:1,b:2,c:3}", res: value.Map{M: listMap.ListMap[value.Value]{
			{Key: "a", Value: value.Int(1)},
			{Key: "b", Value: value.Int(2)},
			{Key: "c", Value: value.Int(3)},
		}}},
		{exp: "{a:1,b:2,c:3}.b", res: value.Int(2)},
		{exp: "[1,2,3]", res: value.NewList(value.Int(1), value.Int(2), value.Int(3))},
		{exp: "let a=2; [1,a,3]", res: value.NewList(value.Int(1), value.Int(2), value.Int(3))},
		{exp: "[1,2,3][2]", res: 3},
		{exp: "let a=1;a", res: 1},
		{exp: "let sqr=x->x*x;sqr(2)", res: 4},
		{exp: "let s=3; let f=x->x*x*s;f(2)", res: 12},
		{exp: "func fib(n) if n<=2 then 1 else fib(n-1)+fib(n-2);[fib(10),fib(15)]", res: value.NewList(value.Int(55), value.Int(610))},
		{exp: "if 1<2 then 1 else 2", res: value.Int(1)},
		{exp: "if 1>2 then 1 else 2", res: value.Int(2)},
		{exp: "let a=2; if 1<a then 1 else 2", res: value.Int(1)},
		{exp: "let a=2; if 1>a then 1 else 2", res: value.Int(2)},
		{exp: "[1,2,3].size()", res: value.Int(3)},
		{exp: "[1,2,3].map(e->e*2)", res: value.NewList(value.Int(2), value.Int(4), value.Int(6))},
		{exp: "[1,2,3,4,5].reduce((a,b)->a+b)", res: value.Int(15)},
		{exp: "let a=1;sprintf(\"%v->%v\",a,2)", res: value.String("1->2")},
		{exp: "let a=1;sprintf(\"%v->\",a)", res: value.String("1->")},
		{exp: "{a:x->x*2,b:x->x*3}.b(4)", res: value.Int(12)},
		{exp: "const a=2;const b=3; a*b", res: value.Int(6)},
		{exp: "func g(a) switch a case 0:\"Test\" case 1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: value.NewList(value.String("Test"), value.String("Hello"), value.String("World"))},
		{exp: "func g(a) switch true case a=0:\"Test\" case a=1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: value.NewList(value.String("Test"), value.String("Hello"), value.String("World"))},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			fu, err := DynType.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
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
		{exp: "1+2", res: value.Int(3)},
		{exp: "\"test\"+\"hello\"", res: value.String("testhello")},
		{exp: "[1+2,8/4]", res: value.NewList(value.Int(3), value.Int(2))},
		{exp: "{a:1+2,b:8/4}", res: value.Map{M: *listMap.NewP[value.Value](3).Put("a", value.Int(3)).Put("b", value.Int(2))}},
		{exp: "(1+pi)/(pi+1)", res: value.Float(1)},
		{exp: "sqrt(4/2)", res: value.Float(math.Sqrt(2))},
		{exp: "(1<2) & (2<3)", res: value.Bool(true)},
		{exp: "-2/(-1)", res: value.Float(2)},
		{exp: "sprintf(\"%v->%v\",1,2)", res: value.String("1->2")},
		{exp: "sprintf(\"%v->\",1)", res: value.String("1->")},
		{exp: "const a=sqrt(2);const b=a*a; b", res: value.Float(2)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			ast, err := DynType.CreateAst(test.exp)
			assert.NoError(t, err, test.exp)
			if c, ok := ast.(*parser2.Const[value.Value]); ok {
				if f, ok := test.res.(value.Float); ok {
					fl, ok := c.Value.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(f), fl, 1e-7)
				} else {
					assert.EqualValues(t, test.res, c.Value)
				}
			} else {
				t.Errorf("not a constant: %v -> %v", test.exp, ast)
			}
		})
	}
}

// The power of closures and recursion.
// Recursive implementation of the sqrt function using the Regula-Falsi algorithm.
const regulaFalsi = `
      func regulaFalsi(rf)
          let xn = (rf.x0*rf.f1 - rf.x1*rf.f0) / (rf.f1 - rf.f0);
          let fn = rf.f(xn);

          let next = if abs(rf.f0) > abs(rf.f1)
                       then {x0:xn, f0:fn, x1:rf.x1, f1:rf.f1, f:rf.f}
                       else {x0:rf.x0, f0:rf.f0, x1:xn, f1:fn, f:rf.f};

          if abs(fn)<1e-7 
            then next 
            else regulaFalsi(next);

      func solve(x0, x1, f)
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
      func newton(x,a) 
         if abs(x*x-a)<1e-7 
         then x 
         else newton(x+(a-x*x)/(2*x), a);
      func mySqrt(a) 
         newton(2,a); 

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

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			f, err := DynType.Generate(test.exp, "a")
			assert.NoError(t, err, test.name)
			if f != nil {
				r, err := f(funcGen.NewStack[value.Value](value.Float(2)))
				assert.NoError(t, err, test.name)
				res, ok := r.ToFloat()
				assert.True(t, ok)
				assert.InDelta(t, math.Sqrt(2), res, 1e-6, test.name)
			}
		})
	}
}

func BenchmarkRegulaFalsi(b *testing.B) {
	f, _ := DynType.Generate(regulaFalsi, "a")
	args := funcGen.NewStack[value.Value](value.Float(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkCall(b *testing.B) {
	f, _ := DynType.Generate("x+(2*y/x)", "x", "y")
	args := funcGen.NewStack[value.Value](value.Float(3), value.Float(3))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc(b *testing.B) {
	f, _ := DynType.Generate("func f(x) x*x;f(a)+f(2*a)", "a")
	args := funcGen.NewStack[value.Value](value.Float(3))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc2(b *testing.B) {
	f, _ := DynType.Generate("let c=1.5;func mul(x) y->y*x*c;mul(b)(a)", "a", "b")
	args := funcGen.NewStack[value.Value](value.Float(3), value.Float(2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkList(b *testing.B) {
	f, err := DynType.Generate("l.map(e->e*e).map(e->e/100)", "l")
	if err != nil {
		fmt.Println(err)
	}

	l := make([]value.Value, 1000)
	for i := range l {
		l[i] = value.Float(i)
	}

	args := funcGen.NewStack[value.Value](value.NewList(l...))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}
