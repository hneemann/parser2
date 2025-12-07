package parser2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrettyPrintAST(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		want string
	}{
		{"simple", "let x=a+2;x+a", "let x = a+2;\n\nx+a"},
		{"pri", "(a+2)*3", "(a+2)*3"},
		{"pri2", "a+2*a", "a+2*a"},
		{"unary", "-a", "-a"},
		{"fu", "func f(n) n^2; f(5)", "func f(n)\n  n^2;\n\nf(5)"},
		{"fuOpt", "func f(n) 2*2*n; f(5)", "func f(n)\n  4*n;\n\nf(5)"},
		{"fu2", "a(2)", "a(2)"},
		{"fu3", "a(2,3,4)", "a(2, 3, 4)"},
		{"fu4", "a(2,a.abs(),4)", "a(2, a\n      .abs(), 4)"},
		{"fu5", "a(a.abs().sqrt(),a.abs().sqrt(),a.abs().sqrt())", "a(a\n   .abs()\n   .sqrt(),\n  a\n   .abs()\n   .sqrt(),\n  a\n   .abs()\n   .sqrt()\n  )"},
		{"method", "a.bla(5)", "a\n .bla(5)"},
		{"method2", "(1+a).toFloat()", "(1+a)\n .toFloat()"},
		{"method3", "(x->x*x).toFloat()", "(x -> x*x)\n .toFloat()"},
		{"method4", "(x->x*x.abs().neg()).toFloat()", "(x -> x*x\n         .abs()\n         .neg())\n .toFloat()"},
		{"attr", "a.bla", "a.bla"},
		{"index", "a[1].bla", "a[1].bla"},
		{"cl1", "a(x->x^2)", "a(x -> x^2)"},
		{"cl2", "a((x,y)->x^2+y^2)", "a((x, y) -> x^2+y^2)"},
		{"if", "if a=0 then 0 else if a<0 then -1 else 1", "if a=0\nthen 0\nelse if a<0\n     then -1\n     else 1"},
		{"if2", "a(if a=0 then 0 else if a<0 then -1 else 1)", "a(if a=0\n  then 0\n  else if a<0\n       then -1\n       else 1)"},
		{"switch", "switch a=0 case 0: 1 case 1: 3 default -1", "switch a=0\n  case 0: 1\n  case 1: 3\n  default -1"},
		{"switch2", "a(switch a=0 case 0: 1 case 1: 3 default -1)", "a(switch a=0\n    case 0: 1\n    case 1: 3\n    default -1)"},
		{"try", "try a+1 catch 1", "try a+1 catch 1"},
		{"listLit", "[1,2,3]", "[1, 2, 3]"},
		{"mapLit", "{a:1,b:2}", "{a: 1,\n b: 2}"},
		{"mapLit2", "a({a:1,b:2})", "a({a: 1,\n   b: 2})"},
		{"e", "(((f->f(f))(h->f->f(x->(f->f(f))(h)(f)(x))))(f->a->b->x->if x=0 then a else f(b)(a + b)(x-1))(0)(1))(12)", "f -> f(f)(h -> f -> f(x -> f -> f(f)(h)(f)(x)))(f -> a -> b -> x -> if x=0\n                                                                    then a\n                                                                    else f(b)(a+b)(x-1))(0)(1)(12)"},
	}

	var idents Identifiers[int]
	idents = idents.Add("a")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.exp, idents)
			assert.NoError(t, err)
			printAST := PrettyPrint[int](ast)
			assert.Equalf(t, tt.want, printAST, "PrettyPrintAST(%v)", tt.exp)
			fmt.Println(printAST)
		})
	}
}
