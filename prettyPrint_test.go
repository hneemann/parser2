package parser2

import (
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
		{"fu2", "a(2)", "a(2)"},
		{"method", "a.bla(5)", "a\n  .bla(5)"},
		{"method2", "(1+a).toFloat()", "(1+a)\n  .toFloat()"},
		{"method3", "(x->x*x).toFloat()", "(x -> x*x)\n  .toFloat()"},
		{"attr", "a.bla", "a.bla"},
		{"index", "a[1].bla", "a[1].bla"},
		{"cl1", "a(x->x^2)", "a(x -> x^2)"},
		{"cl2", "a((x,y)->x^2+y^2)", "a((x, y) -> x^2+y^2)"},
		{"if", "if a=0 then 0 else if a<0 then -1 else 1", "if a=0\nthen\n  0\nelse\n  if a<0\n  then\n    -1\n  else\n    1"},
		{"if2", "a(if a=0 then 0 else if a<0 then -1 else 1)", "a(if a=0\nthen\n  0\nelse\n  if a<0\n  then\n    -1\n  else\n    1)"},
		{"switch", "switch a=0 case 0: 1 case 1: 3 default -1", "switch a=0\ncase 0:\n  1\ncase 1:\n  3\ndefault\n  -1"},
		{"switch2", "a(switch a=0 case 0: 1 case 1: 3 default -1)", "a(switch a=0\ncase 0:\n  1\ncase 1:\n  3\ndefault\n  -1)"},
		{"switch2", "a(switch a=0 case 0: 1 case 1: 3 default -1)", "a(switch a=0\ncase 0:\n  1\ncase 1:\n  3\ndefault\n  -1)"},
		{"try", "try a+1 catch 1", "try a+1 catch 1"},
		{"listLit", "[1,2,3]", "[1, 2, 3]"},
		{"mapLit", "{a:1,b:2}", "{a:1, b:2}"},
	}

	var idents Identifiers[int]
	idents = idents.Add("a")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.exp, idents)
			assert.NoError(t, err)
			printAST := PrettyPrint[int](ast)
			assert.Equalf(t, tt.want, printAST, "PrettyPrintAST(%v)", tt.exp)
		})
	}
}
