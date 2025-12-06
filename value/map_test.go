package value

import (
	"errors"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	runTest(t, []testType{
		{exp: "{a:x->x*2,b:x->x*3}.b(4)", res: Int(12)},
		{exp: "{a:1,b:2,c:3}", res: NewMap(listMap.New[Value](3).
			Append("a", Int(1)).
			Append("b", Int(2)).
			Append("c", Int(3)))},
		{exp: "{a:1,b:2,c:3}.b", res: Int(2)},
		{exp: "{a:1,b:2,c:3}={a:1,b:2,c:3}", res: Bool(true)},
		{exp: "{a:1,b:2,c:3}={a:1,b:2,c:4}", res: Bool(false)},
		{exp: "{a:1,b:2,c:3}={a:1,b:2,d:3}", res: Bool(false)},
		{exp: "{a:1,b:2,c:3}.map((k,v)->v*v)", res: NewMap(listMap.New[Value](3).
			Append("a", Int(1)).
			Append("b", Int(4)).
			Append("c", Int(9)))},
		{exp: "{a:1,b:2,c:3}.accept((k,v)->v>1)", res: NewMap(listMap.New[Value](2).
			Append("b", Int(2)).
			Append("c", Int(3)))},
		{exp: "{a:1,b:2,c:3}.list()", res: NewList(
			NewMap(listMap.New[Value](2).
				Append("key", String("a")).
				Append("value", Int(1))),
			NewMap(listMap.New[Value](2).
				Append("key", String("b")).
				Append("value", Int(2))),
			NewMap(listMap.New[Value](2).
				Append("key", String("c")).
				Append("value", Int(3))),
		)},
		{exp: "{a:1,b:2,c:3,d:-1}.size()", res: Int(4)},
		{exp: "{a:1,b:2}.replaceMap(m->m.a+m.b)", res: Int(3)},
		{exp: "{a:1,b:2}.isAvail()", res: Bool(true)},
		{exp: "{a:1,b:2}.isAvail(\"a\")", res: Bool(true)},
		{exp: "{a:1,b:2}.isAvail(\"c\")", res: Bool(false)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"a\",\"b\")", res: Bool(true)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"b\",\"c\")", res: Bool(true)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"b\",\"d\")", res: Bool(false)},
		{exp: "\"a\" ~ {a:1,b:2}", res: Bool(true)},
		{exp: "\"c\" ~ {a:1,b:2}", res: Bool(false)},
		{exp: "{a:1,b:2}.get(\"a\")", res: Int(1)},
		{exp: "\"\"+{a:1,b:2}.put(\"c\",3)", res: String("{c:3, a:1, b:2}")},
		{exp: "({a:1,b:2}+{c:3,d:4}).string()", res: String("{a:1, b:2, c:3, d:4}")},
		{exp: "{a:1,b:2}.put(\"c\",3).string()", res: String("{c:3, a:1, b:2}")},
		{exp: "{a:1,b:2}.put(\"c\",3).c", res: Int(3)},
		{exp: "{a:1,b:2}.put(\"c\",3).b", res: Int(2)},
		{exp: "{a:1,b:2}.put(\"c\",3).size()", res: Int(3)},
		{exp: "{a:1,b:2}.combine({a:3,b:4},(a,b)->a+b).string()", res: String("{a:4, b:6}")},

		{exp: "{a:1,b:2,c:3}.replace(m->{b:m.b+5}).string()", res: String("{a:1, b:7, c:3}")},
		{exp: "{a:1,b:2,c:3}.replace(m->{b:m.b+5,c:m.c+5}).string()", res: String("{a:1, b:7, c:8}")},
	})
}

func TestMap_Equals(t *testing.T) {
	tests := []struct {
		name string
		a, b Map
		want bool
	}{
		{name: "empty", a: Map{RealMap{}}, b: Map{RealMap{}}, want: true},
		{name: "one empty", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{}}, want: false},
		{name: "one empty", a: Map{RealMap{}}, b: Map{RealMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{RealMap{"a": Int(2)}}, b: Map{RealMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{"b": Int(1)}}, want: false},
		{name: "equal", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{"a": Int(1)}}, want: true},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			equals, err := test.a.Equals(funcGen.NewEmptyStack[Value](), test.b, func(st funcGen.Stack[Value], a, b Value) (bool, error) {
				if aa, ok := a.(Int); ok {
					if bb, ok := b.(Int); ok {
						return aa == bb, nil
					}
				}
				return false, errors.New("not an int")
			})
			assert.NoError(t, err)
			assert.Equalf(t, test.want, equals, "Equals(%v)", test.b)
		})
	}
}

func TestFuncMap(t *testing.T) {
	mf := NewFuncMapFactory(func(i Int, key string) (Value, bool) {
		if key == "a" {
			return i, true
		} else if key == "b" {
			return i, true
		}
		return Int(0), false
	}, "a", "b")
	m := mf.Create(1)

	v, ok := m.Get("a")
	assert.True(t, ok)
	assert.EqualValues(t, 1, v)
	v, ok = m.Get("b")
	assert.True(t, ok)
	assert.EqualValues(t, 1, v)
	_, ok = m.Get("v")
	assert.False(t, ok)

	n := 0
	m.Iter(func(key string, value Value) bool {
		assert.True(t, key == "a" || key == "b")
		assert.EqualValues(t, 1, value)
		n++
		return true
	})
	assert.EqualValues(t, 2, n)

}
