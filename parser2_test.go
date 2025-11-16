package parser2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

type simpleOptimizer struct{}

func (so simpleOptimizer) Optimize(ast AST) (AST, error) {
	if o, ok := ast.(*Operate); ok {
		if an, aOk := o.A.(*Const[int]); aOk {
			if bn, bOk := o.B.(*Const[int]); bOk {
				switch o.Operator {
				case "+":
					return &Const[int]{an.Value + bn.Value, o.Line}, nil
				case "-":
					return &Const[int]{an.Value - bn.Value, o.Line}, nil
				case "*":
					return &Const[int]{an.Value * bn.Value, o.Line}, nil
				case "/":
					return &Const[int]{an.Value / bn.Value, o.Line}, nil
				}
			}
		}
	}
	return nil, nil
}

type numberParser struct{}

func (np numberParser) ParseNumber(n string) (int, error) {
	atoi, err := strconv.Atoi(n)
	return atoi, err
}

var parser = NewParser[int]().
	SetKeyWords("let", "switch", "case", "default", "func").
	SetNumberParser(numberParser{}).
	SetOptimizer(&simpleOptimizer{}).
	Op("+", "-", "*", "/", "^").
	Unary("-")

var parserComfort = NewParser[int]().
	Comfort(true).
	SetKeyWords("let", "switch", "case", "default").
	SetNumberParser(numberParser{}).
	SetOptimizer(&simpleOptimizer{}).
	Op("+", "-", "*", "/", "^").
	Unary("-")

func TestParser(t *testing.T) {
	tests := []struct {
		exp  string
		args []string
		opt  string
	}{
		{exp: "(1+1)*(2+2)", opt: "8"},
		{exp: "(a,b)->a*b*(1+1)", opt: "(a, b)->(a*b)*2", args: []string{"a", "b"}},
		{exp: "a->a*(1+1)", opt: "a->a*2", args: []string{"a"}},
		{exp: "f(1+1,2+2)", opt: "f(2, 4)", args: []string{"f"}},
		{exp: "a[1+1](2+2)", opt: "a[2](4)", args: []string{"a"}},
		{exp: "(1+1)[2+2]", opt: "2[4]"},
		{exp: "(1+1).m[2+2]", opt: "2.m[4]"},
		{exp: "(2+4)/(1+10/2)", opt: "1"},
		{exp: "[1+1,2+2,3+3]", opt: "[2, 4, 6]"},
		{exp: "[1+1,2+2,3+3,]", opt: "[2, 4, 6]"},
		{exp: "let v=1+1; 2+v", opt: "4"},
		{exp: "switch a case 0:1 case 1:10 default 100", opt: "switch a case 0 : 1 case 1 : 10 default 100", args: []string{"a"}},
		{exp: "func sqr(x) x*x; sqr(x)", opt: "let sqr=x->x*x; sqr(x)", args: []string{"x"}},
		{exp: "func mul(a,b) a*b; mul(1,2)", opt: "let mul=(a, b)->a*b; mul(1, 2)", args: []string{"a", "b"}},
		{exp: "-(2*2)", opt: "-4"},
		{exp: "{a:1+1, b:2*2}", opt: "{a:2, b:4}"},
		{exp: "{a:1+1, b:2*2,}", opt: "{a:2, b:4}"},
		{exp: "a.m(1+1,2+2)", opt: "a.m(2, 4)", args: []string{"a"}},
		{exp: "2*x³-4*x²+2*x+1", opt: "(((2*(x^3))-(4*(x^2)))+(2*x))+1", args: []string{"x"}},
		{exp: "x^3", opt: "x^3", args: []string{"x"}},
		{exp: "xˆ3", opt: "x^3", args: []string{"x"}},
		{exp: "xˆ3+xˆ2", opt: "(x^3)+(x^2)", args: []string{"x"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			var idents Identifiers[int]
			for _, arg := range test.args {
				idents = idents.Add(arg)
			}
			ast, err := parser.Parse(test.exp, idents, false)
			assert.NoError(t, err, test.exp)
			if ast != nil {
				assert.EqualValues(t, test.opt, ast.String())
			}
		})
	}
}

func TestParserComfort(t *testing.T) {
	tests := []struct {
		exp  string
		args []string
		opt  string
	}{
		{exp: "(1+1)*(2+2)", opt: "8"},
		{exp: "(a,b)->a*b*(1+1)", opt: "(a, b)->(a*b)*2", args: []string{"a", "b"}},
		{exp: "a->a*(1+1)", opt: "a->a*2", args: []string{"a"}},
		{exp: "f(1+1,2+2)", opt: "f(2, 4)", args: []string{"f"}},
		{exp: "a[1+1](2+2)", opt: "a[2](4)", args: []string{"a"}},
		{exp: "(1+1)[2+2]", opt: "2[4]"},
		{exp: "(1+1).m[2+2]", opt: "2.m[4]"},
		{exp: "(2+4)/(1+10/2)", opt: "1"},
		{exp: "[1+1,2+2,3+3]", opt: "[2, 4, 6]"},
		{exp: "let v=1+2; 1+v", opt: "4"},
		{exp: "switch a case 0:1 case 1:10 default 100", opt: "switch a case 0 : 1 case 1 : 10 default 100", args: []string{"a"}},
		{exp: "-(2*2)", opt: "-4"},
		{exp: "{a:1+1, b:2*2}", opt: "{a:2, b:4}"},
		{exp: "a.m(1+1,2+2)", opt: "a.m(2, 4)", args: []string{"a"}},
		{exp: "2x³-4x²+2x+1", opt: "(((2*(x^3))-(4*(x^2)))+(2*x))+1", args: []string{"x"}},
		{exp: "2x cos(x) ", opt: "(2*x)*cos(x)", args: []string{"x", "cos"}},
		{exp: "x 2", opt: "x*2", args: []string{"x"}},
		{exp: "cos(x) x", opt: "cos(x)*x", args: []string{"x", "cos"}},
		{exp: "cos(x) 2 x", opt: "(cos(x)*2)*x", args: []string{"x", "cos"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			var idents Identifiers[int]
			for _, arg := range test.args {
				idents = idents.Add(arg)
			}
			ast, err := parserComfort.Parse(test.exp, idents, false)
			assert.NoError(t, err, test.exp)
			if ast != nil {
				assert.EqualValues(t, test.opt, ast.String())
			}
		})
	}
}

type vars map[string]int

type fu func(vars) (int, error)

func codeGen(ast AST) (fu, error) {
	switch a := ast.(type) {
	case *Const[int]:
		return func(vars) (int, error) {
			return a.Value, nil
		}, nil
	case *Ident:
		return func(v vars) (int, error) {
			if i, ok := v[a.Name]; ok {
				return i, nil
			}
			return 0, fmt.Errorf("variable not found: %v", a)
		}, nil
	case *Unary:
		inner, err := codeGen(a.Value)
		if err != nil {
			return nil, err
		}
		switch a.Operator {
		case "-":
			return func(v vars) (int, error) {
				i, err := inner(v)
				if err != nil {
					return 0, err
				}
				return -i, err
			}, nil
		}
		return nil, fmt.Errorf("unsupported unary operator %v", a.Operator)
	case *Operate:
		fA, err := codeGen(a.A)
		if err != nil {
			return nil, err
		}
		fB, err := codeGen(a.B)
		if err != nil {
			return nil, err
		}
		switch a.Operator {
		case "+":
			return func(v vars) (int, error) {
				a, err := fA(v)
				if err != nil {
					return 0, err
				}
				b, err := fB(v)
				if err != nil {
					return 0, err
				}
				return a + b, nil
			}, nil
		case "*":
			return func(v vars) (int, error) {
				av, err := fA(v)
				if err != nil {
					return 0, err
				}
				bv, err := fB(v)
				if err != nil {
					return 0, err
				}
				return av * bv, nil
			}, nil
		}
		return nil, fmt.Errorf("unsupported operator %v", a.Operator)
	default:
		return nil, fmt.Errorf("unsupported %v", a)
	}
}

func TestCodeGen(t *testing.T) {
	tests := []struct {
		exp  string
		args []string
		res  int
	}{
		{exp: "1+2*2", res: 5},
		{exp: "a+b*2", res: 8},
		{exp: "-2", res: -2},
	}

	v := vars{"a": 2, "b": 3}
	for _, test := range tests {
		test := test
		t.Run(test.exp, func(t *testing.T) {
			var idents Identifiers[int]
			for n := range v {
				idents = idents.Add(n)
			}
			ast, err := parser.Parse(test.exp, idents, false)
			assert.NoError(t, err, test.exp)
			fu, err := codeGen(ast)
			assert.NoError(t, err)
			got, err := fu(v)
			assert.NoError(t, err)
			assert.EqualValues(t, test.res, got)
		})
	}
}
