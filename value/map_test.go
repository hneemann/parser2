package value

import (
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
		{exp: "({a:1,b:2}+{c:3,d:4}).string()", res: String("{a:1, b:2, c:3, d:4}")},
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
		{exp: "{a:1,b:2}.replace(m->m.a+m.b)", res: Int(3)},
		{exp: "{a:1,b:2}.isAvail(\"a\")", res: Bool(true)},
		{exp: "{a:1,b:2}.isAvail(\"c\")", res: Bool(false)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"a\",\"b\")", res: Bool(true)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"b\",\"c\")", res: Bool(true)},
		{exp: "{a:1,b:2,c:3}.isAvail(\"b\",\"d\")", res: Bool(false)},
		{exp: "\"a\" ~ {a:1,b:2}", res: Bool(true)},
		{exp: "\"c\" ~ {a:1,b:2}", res: Bool(false)},
		{exp: "{a:1,b:2}.get(\"a\")", res: Int(1)},
		{exp: "\"\"+{a:1,b:2}.append(\"c\",3)", res: String("{c:3, a:1, b:2}")},
		{exp: "{a:1,b:2}.append(\"c\",3).string()", res: String("{c:3, a:1, b:2}")},
		{exp: "{a:1,b:2}.append(\"c\",3).c", res: Int(3)},
		{exp: "{a:1,b:2}.append(\"c\",3).b", res: Int(2)},
		{exp: "{a:1,b:2}.append(\"c\",3).size()", res: Int(3)},
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
			equals, err := test.a.Equals(test.b)
			assert.NoError(t, err)
			assert.Equalf(t, test.want, equals, "Equals(%v)", test.b)
		})
	}
}
