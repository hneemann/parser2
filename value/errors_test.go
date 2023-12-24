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
			if f, ok := st.Get(0).ToInt(); ok {
				if f < n {
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
		{"notFound(a)", "not found: notFound"},
		{"let a=1;notFound(a)", "not found: notFound"},
		{"[].notFound()", "method 'notFound' not found"},
		{"sin(1,2)", "number of args wrong"},
		{"{a:sin(1,2)}", "number of args wrong"},
		{"[].first()", "no items"},
		{"list(0).first()", "no items"},
		{"list(10).multiUse(3)", "needs to be a map"},
		{"list(10).multiUse({a:3})", "contain functions"},
		{"list(10).multiUse({a:(a,b)->a*b})", "one argument"},
		{"list(10).multiUse({a:a->a.map(a->a.a)})", "not a map"},
		{"list(10).multiUse({a:l->l.reduce((a,b)->a.e+b), b:l->l.reduce((a,b)->a+b)})", "not a map"},
		{"list(10).multiUse({a:l->1, b:l->l->2})", "affected function(s): a, b"},
		{"list(10).multiUse({a:l->l.reduce((a,b)->a+b)+l.reduce((a,b)->a*b)})", "function a can only be used once"},
		{"list(10).map(e->e.e).multiUse({a:l->l.reduce((a,b)->a+b)})", "not a map"},
		{"list(10).multiUse({a:l->l.mapReduce(0,(s,i)->s+i), b:l->l.notFound(i->i+1)})", "notFound"},
		{"list(1e9).multiUse({a:l->l.mapReduce(0,(s,i)->s+error(i)), b:l->l.mapReduce((s,i)->s+i)})", "toLarge"},
		{"list(10).multiUse({a:l->l.notFound(0,(s,i)->s+i), b:l->l.notFound(i->i+1)})", "notFound"},
		{"{a:1,b:2}.put(\"b\", 3)", "key 'b' already present in map"},
		{"{a:1,b:2}+{b:3,c:4}", "first map already contains key 'b'"},
		{"{a:1,b:2,c:3}.d", "available are: a, b, c"},
		{"true.d", "not possible; Bool is not a map"},
		{"(2).d", "not possible; Int is not a map"},
		{"(2.2).d", "not possible; Float is not a map"},
		{"[1,2,3,4].d", "not possible; List is not a map"},
		{"[1,2,3,4]+\"test\"", "not allowed on List, String"},
		{"[1,2,3,4].set(-1,0)", "index -1 out of range"},
		{"[1,2,3,4].set(4,0)", "index 4 out of range"},
		{"true-2", "not allowed on Bool, Int"},
		{"func f(x) x+b; f(2)", "outer value 'b' not found"},
	}

	fg := New().AddStaticFunction("error", toLargeErrorFunc(100))
	for _, tt := range tests {
		test := tt
		t.Run(test.exp, func(t *testing.T) {
			f, err := fg.Generate(test.exp)
			if err == nil {
				_, err = f(funcGen.NewEmptyStack[Value]())
			}
			if err == nil {
				t.Errorf("expected an error containing '%v'", test.err)
			} else {
				assert.True(t, strings.Contains(err.Error(), test.err), "expected error containing '%v', got: %v", test.err, err.Error())
			}
		})
	}
}
