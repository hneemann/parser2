package value

import (
	"errors"
	"github.com/hneemann/parser2/funcGen"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func toLargeErrorFunc(n int) funcGen.Function[Value] {
	return funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			if f, ok := st.Get(0).(Int); ok {
				if int(f) < n {
					return st.Get(0), nil
				} else {
					return nil, errors.New("toLarge")
				}
			} else {
				return nil, errors.New("not an int")
			}
		},
		Args:   1,
		IsPure: true,
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		exp string
		err string
	}{
		{"notFound(a)", "identifier 'notFound' not found"},
		{"let a=1;notFound(a)", "identifier 'notFound' not found"},
		{"[].notFound()", "method 'notFound' not found"},
		{exp: "sin(1,2)", err: ", required 1, found 2"},
		{"{a:sin(1,2)}", ", required 1, found 2"},
		{"[].first()", "no items"},
		{"[].single()", "not a single"},
		{"[1,2].single()", "not a single"},
		{"numbers(0).single()", "no item in list"},
		{"numbers(2).single()", "more than one"},
		{"numbers(0).first()", "no items"},
		{"numbers(10).multiUse(3)", "needs to be a map"},
		{"numbers(10).multiUse({a:3})", "contain functions"},
		{"numbers(10).multiUse({a:(a,b)->a*b})", "one argument"},
		{"numbers(10).multiUse({a:a->a.map(a->a.a)})", "not a map"},
		{"numbers(10).multiUse({a:l->l.reduce((a,b)->a.e+b), b:l->l.reduce((a,b)->a+b)})", "not a map"},
		{"numbers(10).multiUse({a:l->1, b:l->l->2})", "timed out"},
		{"numbers(10).multiUse({a:l->l.reduce((a,b)->a+b)+l.reduce((a,b)->a*b)})", "copied iterator a can only be used once"},
		{"numbers(10).map(e->e.e).multiUse({a:l->l.reduce((a,b)->a+b)})", "not a map"},
		{"numbers(10).multiUse({a:l->l.mapReduce(0,(s,i)->s+i), b:l->l.notFound(i->i+1)})", "notFound"},
		{"numbers(1000000000).multiUse({a:l->l.mapReduce(0,(s,i)->s+error(i)), b:l->l.mapReduce((s,i)->s+i)})", "wrong number of arguments"},
		{"numbers(10).multiUse({a:l->l.notFound(0,(s,i)->s+i), b:l->l.notFound(i->i+1)})", "notFound"},
		{"{a:1,b:2}.put(\"b\", 3)", "key 'b' already present in map"},
		{"{a:1,b:2}+{b:3,c:4}", "first map already contains key 'b'"},
		{"{a:1,b:2,c:3}.d", "available are: a, b, c"},
		{"true.d", "not possible; Bool is not a map"},
		{"(2).d", "not possible; Int is not a map"},
		{"(2.2).d", "not possible; Float is not a map"},
		{"[1,2,3,4].d", "not possible; List is not a map"},
		{"[1,2,3,4]+\"test\"", "operation '+' not defined on list, string"},
		{"[1,2,3,4].set(-1,0)", "index -1 out of range"},
		{"[1,2,3,4].set(4,0)", "index 4 out of range"},
		{"true-2", "operation '-' not defined on bool, int"},
		{"func f(x) x+b; f(2)", "identifier 'b' not found"},
		{"func f(x,x) x+x; f(2,2)", "'x' used twice"},
		{"let f=(x,x)-> x+x; f(2,2)", "'x' used twice"},
		{exp: "throw(\"error: zzzz\")", err: "error: zzzz"},
		{exp: "func mul(a,b) a*b; mul.invoke([2,3,4])", err: "wrong number of arguments in invoke: 3 instead of 2"},
		{exp: "func mul(a,b) a*b; mul(2)", err: "wrong number of arguments at call of function, required 2, found 1 in line 1"},
		{exp: "let m={a:(x,y)->x*y};m.a(2)", err: "wrong number of arguments at call of \"a\", required 2, found 1"},
		{exp: "[].size(1)", err: ", required 0, found 1"},
	}

	fg := New().AddStaticFunction("error", toLargeErrorFunc(100))
	for _, tt := range tests {
		test := tt
		t.Run(test.exp, func(t *testing.T) {
			f, err := fg.Generate(test.exp)
			var r Value
			if err == nil {
				r, err = f(funcGen.NewEmptyStack[Value]())
			}
			if err == nil {
				t.Errorf("expected an error containing '%v', result was: %v", test.err, r)
			} else {
				assert.True(t, strings.Contains(err.Error(), test.err), "expected error containing '%v', got: %v", test.err, err.Error())
			}
		})
	}
}
