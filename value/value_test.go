package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/stretchr/testify/assert"
	"math"
	"strings"
	"testing"
	"unicode"
)

type testType struct {
	exp string
	res Value
}

func TestBasic(t *testing.T) {
	runTest(t, []testType{
		{exp: "1e-7", res: Float(1e-7)},
		{exp: "1e7", res: Float(1e7)},
		{exp: "1e+7", res: Float(1e+7)},
		{exp: "1+2", res: Int(3)},
		{exp: "2-1", res: Int(1)},
		{exp: "1.0+2.0", res: Float(3.0)},
		{exp: "2.0-1.0", res: Float(1.0)},
		{exp: "1<<2", res: Int(4)},
		{exp: "8>>2", res: Int(2)},
		{exp: "1<2", res: Bool(true)},
		{exp: "1>2", res: Bool(false)},
		{exp: "1>2", res: Bool(false)},
		{exp: "1<2", res: Bool(true)},
		{exp: "2=2", res: Bool(true)},
		{exp: "1=2", res: Bool(false)},
		{exp: "1.0+2.0", res: Float(3.0)},
		{exp: "3.0*2.0", res: Float(6.0)},
		{exp: "-3.0", res: Float(-3.0)},
		{exp: "3.0^3.0", res: Float(27.0)},
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
		{exp: "let a=2;abs(-a)", res: Int(2)},
		{exp: "let a=-2;abs(a)", res: Int(2)},
		{exp: "let a=2.0;abs(a)", res: Float(2.0)},
		{exp: "let a=2.0;abs(-a)", res: Float(2.0)},
		{exp: "let a=-2.0;abs(a)", res: Float(2.0)},
		{exp: "let a=5;12%a", res: Int(2)},
		{exp: "let a=2;sqr(a)", res: Int(4)},
		{exp: "let a=2.0;sqr(a)", res: Float(4.0)},
		{exp: "let a=2;let b=3;exp(b*ln(a))", res: Float(8.0)},
		{exp: "\"a\"=\"a\"", res: Bool(true)},
		{exp: "\"a\">=\"a\"", res: Bool(true)},
		{exp: "\"a\"<=\"a\"", res: Bool(true)},
		{exp: "\"a\"=\"b\"", res: Bool(false)},
		{exp: "\"a\">\"b\"", res: Bool(false)},
		{exp: "\"a\"<\"b\"", res: Bool(true)},
		{exp: "\"test\">\"hello\"", res: Bool(true)},
		{exp: "nil=nil", res: Bool(true)},
		{exp: "nil=0", res: Bool(false)},
		{exp: "0=nil", res: Bool(false)},
		{exp: "nil=0.0", res: Bool(false)},
		{exp: "0.0=nil", res: Bool(false)},
		{exp: "{a:0}=nil", res: Bool(false)},
		{exp: "0.0=nil", res: Bool(false)},
		{exp: "\"test\"+\"hello\"", res: String("testhello")},
		{exp: "sqrt(2)", res: Float(math.Sqrt(2))},
		{exp: "let x=2;sqrt(x)", res: Float(math.Sqrt(2))},
		{exp: "let a=1;a", res: Int(1)},
		{exp: "let a=1;min(a,2)", res: Int(1)},
		{exp: "let a=1;min(a,2,3)", res: Int(1)},
		{exp: "let a=1;min(3,2,a)", res: Int(1)},
		{exp: "let a=1;max(a,2)", res: Int(2)},
		{exp: "let a=1;max(a,2,3)", res: Int(3)},
		{exp: "let a=1;max(3,2,a)", res: Int(3)},
		{exp: "let sqr=x->x*x;sqr(2)", res: Int(4)},
		{exp: "let x=2;sqr(2)", res: Int(4)},
		{exp: "let x=pi;sin(x)", res: Float(0.0)},
		{exp: "let x=pi;cos(x/2)", res: Float(0.0)},
		{exp: "let s=3; let f=x->x*x*s;f(2)", res: Int(12)},
		{exp: "func inv(x) -x; inv(2)", res: Int(-2)},
		{exp: "func fib(n) if n<=2 then 1 else fib(n-1)+fib(n-2);[fib(10),fib(15)]", res: NewList(Int(55), Int(610))},
		{exp: "if 1<2 then 1 else 2", res: Int(1)},
		{exp: "if 1>2 then 1 else 2", res: Int(2)},
		{exp: "let a=2; if 1<a then 1 else 2", res: Int(1)},
		{exp: "let a=2; if 1>a then 1 else 2", res: Int(2)},
		{exp: "const a=2;const b=3; a*b", res: Int(6)},
		{exp: "func g(a) switch a case 0:\"Test\" case 1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: NewList(String("Test"), String("Hello"), String("World"))},
		{exp: "func g(a) switch true case a=0:\"Test\" case a=1:\"Hello\" default \"World\"; [g(0),g(1),g(100)]", res: NewList(String("Test"), String("Hello"), String("World"))},
		{exp: "int(3.2)", res: Int(3)},
		{exp: "int(3)", res: Int(3)},
		{exp: "float(3.2)", res: Float(3.2)},
		{exp: "float(3)", res: Float(3)},
		{exp: "string(true)", res: String("true")},
		{exp: "true.string()", res: String("true")},
		{exp: "string(\"test\")", res: String("test")},
		{exp: "string(3.2)", res: String("3.2")},
		{exp: "(3.2).string()", res: String("3.2")},
		{exp: "string(3)", res: String("3")},
		{exp: "(3).string()", res: String("3")},
		{exp: "let a=1;sprintf()", res: String("")},
		{exp: "let a=1;sprintf(\"Hello World\")", res: String("Hello World")},
		{exp: "let a=1;sprintf(\"%v->%v\",a,2)", res: String("1->2")},
		{exp: "let a=1;sprintf(\"%v->\",a)", res: String("1->")},

		{exp: "let a=2; func cl(b) x->x*a*b; cl(4)(3)", res: Int(24)},
		{exp: "let a=2; func cl(b) let f=a*b;x->x*f; cl(4)(3)", res: Int(24)},

		{exp: `func bool(a)  
                 if a then true else false;
               {ff:bool(0.0), 
                ft:bool(1.5),
                if:bool(0),
                it:bool(1)}.string()`, res: String("{ff:false, ft:true, if:false, it:true}")},

		{exp: `func mySqrt(a)  
                 if a<0 then throw("sqrt of neg value") else sqrt(a);

               try 2*mySqrt(-1)+1 catch e-> "sqrt of neg value" ~ e`, res: Bool(true)},

		{exp: "let p={a:1,b:2}; try p.a catch 5", res: Int(1)},
		{exp: "let p={a:1,b:2}; try p.c catch 5", res: Int(5)},
		{exp: "let p={a:1,b:2}; try p.c catch e->\"caught error: \"+e", res: String("caught error: key 'c' not found in map; available are: a, b")},

		{exp: "func sqr(a) a*a; sqr.args()", res: Int(1)},
		{exp: "func sqr(a) a*a; sqr.invoke([2])", res: Int(4)},
		{exp: "func mul(a,b) a*b; mul.invoke([2,3])", res: Int(6)},

		// Currying
		{exp: "let m=a->b->a*b; [m(2)(3),m(4)(5),m(4.5)(5.5)].string()", res: String("[6, 20, 24.75]")},
		{exp: "func mul(a) b->a*b; [mul(2)(3),mul(4)(5),mul(4.5)(5.5)].string()", res: String("[6, 20, 24.75]")},
	})
}

func runTest(t *testing.T, tests []testType) {
	valueParser := New()
	for _, test := range tests {
		test := test
		t.Run(shrinkSpace(test.exp), func(t *testing.T) {
			fu, err := valueParser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[Value]())
				assert.NoError(t, err, test.exp)
				if tr, ok := test.res.(Float); ok {
					float, ok := res.(Float)
					assert.True(t, ok)
					assert.InDelta(t, float64(tr), float64(float), 1e-6, test.exp)
				} else if expList, ok := test.res.(*List); ok {
					actList, ok := res.(*List)
					assert.True(t, ok)
					st := funcGen.NewEmptyStack[Value]()
					slice, err := expList.ToSlice(st)
					assert.NoError(t, err)
					toSlice, err := actList.ToSlice(st)
					assert.NoError(t, err)
					assert.Equal(t, slice, toSlice, test.exp)
				} else {
					assert.Equal(t, test.res, res, test.exp)
				}
			}
		})
	}
}

func shrinkSpace(str string) string {
	var b bytes.Buffer
	lastWasSpace := true
	for _, r := range str {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				lastWasSpace = true
				b.WriteRune('_')
			}
		} else {
			b.WriteRune(r)
			lastWasSpace = false
		}
	}
	return b.String()
}

func TestOptimizer(t *testing.T) {
	tests := []struct {
		exp string
		res any
	}{
		{exp: "1+2", res: Int(3)},
		{exp: "\"test\"+\"hello\"", res: String("testhello")},
		{exp: "[1+2,8/4]", res: NewList(Int(3), Float(2))},
		{exp: "{a:1+2,b:8/4}", res: Map{m: listMap.New[Value](3).Append("a", Int(3)).Append("b", Float(2))}},
		{exp: "(1+pi)/(pi+1)", res: Float(1)},
		{exp: "sqrt(4/2)", res: Float(math.Sqrt(2))},
		{exp: "sqr(2)", res: Float(4)},
		{exp: "2^3", res: Float(8)},
		{exp: "(1<2) & (2<3)", res: Bool(true)},
		{exp: "-2/(-1)", res: Float(2)},
		{exp: "const a=sqrt(2);const b=a*a; b", res: Float(2)},
	}

	valueParser := New()
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
					st := funcGen.NewEmptyStack[Value]()
					assert.True(t, ok)
					slice, err := expList.ToSlice(st)
					assert.NoError(t, err)
					toSlice, err := actList.ToSlice(st)
					assert.NoError(t, err)
					assert.EqualValues(t, slice, toSlice, test.exp)
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

func TestMethodError(t *testing.T) {
	valueParser := New()
	f, err := valueParser.Generate("a.notFound()", "a")
	assert.NoError(t, err)
	_, err = f(funcGen.NewStack[Value](Float(2)))
	assert.Error(t, err)
	es := err.Error()
	assert.True(t, strings.Contains(es, "method 'notFound' not found"))
}

func TestSolve(t *testing.T) {
	tests := []struct {
		name string
		exp  string
	}{
		{name: "regulaFalsi", exp: regulaFalsi},
		{name: "newtonRaphson", exp: newtonRaphson},
	}

	valueParser := New()
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
	f, _ := New().Generate(regulaFalsi, "a")
	args := funcGen.NewStack[Value](Float(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkCall(b *testing.B) {
	f, _ := New().Generate("x+(2*y/x)", "x", "y")
	args := funcGen.NewStack[Value](Float(3), Float(3))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc(b *testing.B) {
	f, _ := New().Generate("func f(x) x*x;f(a)+f(2*a)", "a")
	args := funcGen.NewStack[Value](Float(3))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkFunc2(b *testing.B) {
	f, _ := New().Generate("let c=1.5;func mul(x) y->y*x*c;mul(b)(a)", "a", "b")
	args := funcGen.NewStack[Value](Float(3), Float(2))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(args)
	}
}

func BenchmarkList(b *testing.B) {
	f, err := New().Generate("l.map(e->e*e).map(e->e/100)", "l")
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
