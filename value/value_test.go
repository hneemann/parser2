package value

import (
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestValueType(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "1e-7", res: float64(1e-7)},
		{exp: "1e7", res: float64(1e7)},
		{exp: "1e+7", res: float64(1e+7)},
		{exp: "1+2", res: Int(3)},
		{exp: "2-1", res: Int(1)},
		{exp: "1.0+2.0", res: 3.0},
		{exp: "2.0-1.0", res: 1.0},
		{exp: "1<2", res: Bool(true)},
		{exp: "1>2", res: Bool(false)},
		{exp: "1>2", res: Bool(false)},
		{exp: "1<2", res: Bool(true)},
		{exp: "2=2", res: Bool(true)},
		{exp: "1=2", res: Bool(false)},
		{exp: "1.0+2.0", res: 3.0},
		{exp: "3.0*2.0", res: 6.0},
		{exp: "-3.0", res: -3.0},
		{exp: "3.0^3.0", res: 27.0},
		{exp: "3^4", res: Int(81)},
		{exp: "2^12", res: Int(4096)},
		{exp: "1.0<2.0", res: Bool(true)},
		{exp: "1.0>2.0", res: Bool(false)},
		{exp: "1.0>2.0", res: Bool(false)},
		{exp: "1.0<2.0", res: Bool(true)},
		{exp: "2.0=2.0", res: Bool(true)},
		{exp: "1.0=2.0", res: Bool(false)},
		{exp: "2!=2", res: Bool(false)},
		{exp: "1!=2", res: Bool(true)},
		{exp: "2>=2", res: Bool(true)},
		{exp: "2<=2", res: Bool(true)},
		{exp: "2.0>=2.0", res: Bool(true)},
		{exp: "2.0<=2.0", res: Bool(true)},
		{exp: "!(1<2)", res: Bool(false)},
		{exp: "1<2 & 3<4", res: Bool(true)},
		{exp: "1<2 & 3>4", res: Bool(false)},
		{exp: "1<2 | 3>4", res: Bool(true)},
		{exp: "1<2 | 3>4", res: Bool(true)},
		{exp: "1>2 | 3>4", res: Bool(false)},
		{exp: "let a=2;1<a & 3<4", res: Bool(true)},
		{exp: "let a=2;1<a & 3>4", res: Bool(false)},
		{exp: "let a=2;1<a | 3>4", res: Bool(true)},
		{exp: "let a=2;1<a | 3>4", res: Bool(true)},
		{exp: "let a=2;1>a | 3>4", res: Bool(false)},
		{exp: "let a=2;abs(a)", res: Int(2)},
		{exp: "let a= -2;abs(a)", res: Int(2)},
		{exp: "let a=2.0;abs(a)", res: 2.0},
		{exp: "let a= -2.0;abs(a)", res: 2.0},
		{exp: "let a=2;sqr(a)", res: Int(4)},
		{exp: "let a=2.0;sqr(a)", res: 4.0},
		{exp: "\"a\"=\"a\"", res: Bool(true)},
		{exp: "\"a\">=\"a\"", res: Bool(true)},
		{exp: "\"a\"<=\"a\"", res: Bool(true)},
		{exp: "\"a\"=\"b\"", res: Bool(false)},
		{exp: "\"a\">\"b\"", res: Bool(false)},
		{exp: "\"a\"<\"b\"", res: Bool(true)},
		{exp: "\"test\">\"hello\"", res: Bool(true)},
		{exp: "\"test\"+\"hello\"", res: String("testhello")},
		{exp: "sqrt(2)", res: math.Sqrt(2)},
		{exp: "let x=2;sqrt(x)", res: math.Sqrt(2)},
		{exp: "{a:1,b:2,c:3}", res: Map{M: listMap.ListMap[Value]{
			{Key: "a", Value: Int(1)},
			{Key: "b", Value: Int(2)},
			{Key: "c", Value: Int(3)},
		}}},
		{exp: "{a:1,b:2,c:3}.b", res: Int(2)},
		{exp: "[1,2,3]", res: NewList(Int(1), Int(2), Int(3))},
		{exp: "let a=2; [1,a,3]", res: NewList(Int(1), Int(2), Int(3))},
		{exp: "let a=2;[1,a]+[3,4]", res: NewList(Int(1), Int(2), Int(3), Int(4))},
		{exp: "[1,2,3][2]", res: Int(3)},
		{exp: "let a=1;a", res: Int(1)},
		{exp: "let sqr=x->x*x;sqr(2)", res: Int(4)},
		{exp: "let x=2;sqr(2)", res: Int(4)},
		{exp: "let x=pi;sin(x)", res: 0.0},
		{exp: "let x=pi;cos(x/2)", res: 0.0},
		{exp: "let s=3; let f=x->x*x*s;f(2)", res: Int(12)},
		{exp: "func fib(n) if n<=2 then 1 else fib(n-1)+fib(n-2);[fib(10),fib(15)]", res: NewList(Int(55), Int(610))},
		{exp: "if 1<2 then 1 else 2", res: Int(1)},
		{exp: "if 1>2 then 1 else 2", res: Int(2)},
		{exp: "let a=2; if 1<a then 1 else 2", res: Int(1)},
		{exp: "let a=2; if 1>a then 1 else 2", res: Int(2)},
		{exp: "[1,2].replace(l->l[0]+l[1])", res: Int(3)},
		{exp: "[1,2,3].indexOf(2)", res: Int(1)},
		{exp: "[1,2,3].indexOf(7)", res: Int(-1)},
		{exp: "2 ~ [1,2,3]", res: Bool(true)},
		{exp: "7 ~ [1,2,3]", res: Bool(false)},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "[1,2,3]=[1,2,3]", res: Bool(true)},
		{exp: "[1,2,3]=[1,2,4]", res: Bool(false)},
		{exp: "[1,2,3]=[1,2]", res: Bool(false)},
		{exp: "[1,2,3].map(e->e*2)", res: NewList(Int(2), Int(4), Int(6))},
		{exp: "[1,2,3,4,5].reduce((a,b)->a+b)", res: Int(15)},
		{exp: "{a:x->x*2,b:x->x*3}.b(4)", res: Int(12)},
		{exp: "const a=2;const b=3; a*b", res: Int(6)},
		{exp: "func g(a) switch a case 0:\"Test\" case 1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: NewList(String("Test"), String("Hello"), String("World"))},
		{exp: "func g(a) switch true case a=0:\"Test\" case a=1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: NewList(String("Test"), String("Hello"), String("World"))},
		{exp: "[1,2,3].map(i->i*i)", res: NewList(Int(1), Int(4), Int(9))},
		{exp: "[1,2,3].accept(i->i>1)", res: NewList(Int(2), Int(3))},
		{exp: "[1,2,3].accept(i->i>1)", res: NewList(Int(2), Int(3))},
		{exp: "[1,2,3,3].reduce((a,b)->a+b)", res: Int(9)},
		{exp: "(3.2).int()", res: Int(3)},
		{exp: "(3).int()", res: Int(3)},
		// Prefix Sum
		{exp: "[1,2,3,4,4].iir(i->i,(i,l)->i+l)", res: NewList(Int(1), Int(3), Int(6), Int(10), Int(14))},
		// Fibonacci Sequence
		{exp: "list(12).iir(i->[1,1],(i,l)->[l[1],l[0]+l[1]]).map(l->l[0])",
			res: NewList(Int(1), Int(1), Int(2), Int(3), Int(5), Int(8), Int(13), Int(21), Int(34), Int(55), Int(89), Int(144))},
		// Low-pass Filter
		{exp: "list(11).iir(i->0,(i,l)->(1024+l)>>1)",
			res: NewList(Int(0), Int(512), Int(768), Int(896), Int(960), Int(992), Int(1008), Int(1016), Int(1020), Int(1022), Int(1023))},
		{exp: "list(6).combine((a,b)->a+b)", res: NewList(Int(1), Int(3), Int(5), Int(7), Int(9))},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "{a:1,b:2,c:3}.map((k,v)->v*v)", res: Map{M: listMap.ListMap[Value]{
			{Key: "a", Value: Int(1)},
			{Key: "b", Value: Int(4)},
			{Key: "c", Value: Int(9)},
		}}},
		{exp: "{a:1,b:2,c:3}.accept((k,v)->v>1)", res: Map{M: listMap.ListMap[Value]{
			{Key: "b", Value: Int(2)},
			{Key: "c", Value: Int(3)},
		}}},
		{exp: "{a:1,b:2,c:3}.list()", res: NewList(
			Map{M: listMap.ListMap[Value]{
				{Key: "key", Value: String("a")},
				{Key: "value", Value: Int(1)},
			}},
			Map{M: listMap.ListMap[Value]{
				{Key: "key", Value: String("b")},
				{Key: "value", Value: Int(2)},
			}},
			Map{M: listMap.ListMap[Value]{
				{Key: "key", Value: String("c")},
				{Key: "value", Value: Int(3)},
			}},
		)},
		{exp: "{a:1,b:2,c:3,d:-1}.size()", res: Int(4)},
		{exp: "{a:1,b:2}.replace(m->m.a+m.b)", res: Int(3)},
		{exp: "\"\"+list(12).group(i->\"n\"+round(i/4),i->i).list().order((a,b)->a.key<b.key)",
			res: String("[{key:n0, value:[0, 1]}, {key:n1, value:[2, 3, 4, 5]}, {key:n2, value:[6, 7, 8, 9]}, {key:n3, value:[10, 11]}]")},

		{exp: "\"Hello World\".len()", res: Int(11)},
		{exp: "\"Hello World\".indexOf(\"Wo\")", res: Int(6)},
		{exp: "\"Hello World\".toLower()", res: String("hello world")},
		{exp: "\"Hello World\".toUpper()", res: String("HELLO WORLD")},
		{exp: "\"Hello World\".contains(\"Wo\")", res: Bool(true)},
		{exp: "\"Hello World\".contains(\"wo\")", res: Bool(false)},

		{exp: "{a:1,b:2}.isAvail(\"a\")", res: Bool(true)},
		{exp: "{a:1,b:2}.isAvail(\"c\")", res: Bool(false)},
		{exp: "{a:1,b:2}.get(\"a\")", res: Int(1)},
		{exp: "\"\"+{a:1,b:2}.put(\"c\",3)", res: String("{c:3, a:1, b:2}")},
		{exp: "{a:1,b:2}.put(\"c\",3).c", res: Int(3)},
		{exp: "{a:1,b:2}.put(\"c\",3).b", res: Int(2)},
		{exp: "{a:1,b:2}.put(\"c\",3).size()", res: Int(3)},
	}

	valueParser := SetUpParser(New())
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			fu, err := valueParser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[Value]())
				assert.NoError(t, err, test.exp)
				if _, ok := test.res.(float64); ok {
					float, ok := res.(Float)
					assert.True(t, ok)
					assert.InDelta(t, test.res, float64(float), 1e-6, test.exp)
				} else if expList, ok := test.res.(*List); ok {
					actList, ok := res.(*List)
					assert.True(t, ok)
					assert.EqualValues(t, expList.ToSlice(), actList.ToSlice(), test.exp)
				} else {
					assert.Equal(t, test.res, res, test.exp)
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
		{exp: "1+2", res: Int(3)},
		{exp: "\"test\"+\"hello\"", res: String("testhello")},
		{exp: "[1+2,8/4]", res: NewList(Int(3), Float(2))},
		{exp: "{a:1+2,b:8/4}", res: Map{M: *listMap.NewP[Value](3).Put("a", Int(3)).Put("b", Float(2))}},
		{exp: "(1+pi)/(pi+1)", res: Float(1)},
		{exp: "sqrt(4/2)", res: Float(math.Sqrt(2))},
		{exp: "sqr(2)", res: Float(4)},
		{exp: "2^3", res: Float(8)},
		{exp: "(1<2) & (2<3)", res: Bool(true)},
		{exp: "-2/(-1)", res: Float(2)},
		{exp: "const a=sqrt(2);const b=a*a; b", res: Float(2)},
	}

	valueParser := SetUpParser(New())
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			ast, err := valueParser.CreateAst(test.exp)
			assert.NoError(t, err, test.exp)
			if c, ok := ast.(*parser2.Const[Value]); ok {
				if f, ok := test.res.(Float); ok {
					fl, ok := c.Value.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(f), fl, 1e-7)
				} else if expList, ok := test.res.(*List); ok {
					actList, ok := c.Value.(*List)
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

	valueParser := SetUpParser(New())
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			f, err := valueParser.Generate(test.exp, "a")
			assert.NoError(t, err, test.name)
			if f != nil {
				r, err := f(funcGen.NewStack[Value](Float(2)))
				assert.NoError(t, err, test.name)
				res, ok := r.ToFloat()
				assert.True(t, ok)
				assert.InDelta(t, math.Sqrt(2), res, 1e-6, test.name)
			}
		})
	}
}

func BenchmarkRegulaFalsi(b *testing.B) {
	f, _ := SetUpParser(New()).Generate(regulaFalsi, "a")
	args := funcGen.NewStack[Value](Float(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkCall(b *testing.B) {
	f, _ := SetUpParser(New()).Generate("x+(2*y/x)", "x", "y")
	args := funcGen.NewStack[Value](Float(3), Float(3))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc(b *testing.B) {
	f, _ := SetUpParser(New()).Generate("func f(x) x*x;f(a)+f(2*a)", "a")
	args := funcGen.NewStack[Value](Float(3))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc2(b *testing.B) {
	f, _ := SetUpParser(New()).Generate("let c=1.5;func mul(x) y->y*x*c;mul(b)(a)", "a", "b")
	args := funcGen.NewStack[Value](Float(3), Float(2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkList(b *testing.B) {
	f, err := SetUpParser(New()).Generate("l.map(e->e*e).map(e->e/100)", "l")
	if err != nil {
		fmt.Println(err)
	}

	l := make([]Value, 1000)
	for i := range l {
		l[i] = Float(i)
	}

	args := funcGen.NewStack[Value](NewList(l...))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}
