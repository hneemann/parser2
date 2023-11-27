package value

import (
	"github.com/hneemann/parser2/listMap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	runTest(t, []testType{
		{exp: "{a:1,b:2,c:3}", res: Map{m: listMap.ListMap[Value]{
			{Key: "a", Value: Int(1)},
			{Key: "b", Value: Int(2)},
			{Key: "c", Value: Int(3)},
		}}},
		{exp: "{a:1,b:2,c:3}.b", res: Int(2)},
		{exp: "{a:x->x*2,b:x->x*3}.b(4)", res: Int(12)},
		{exp: "{a:1,b:2,c:3}.map((k,v)->v*v)", res: Map{m: listMap.ListMap[Value]{
			{Key: "a", Value: Int(1)},
			{Key: "b", Value: Int(4)},
			{Key: "c", Value: Int(9)},
		}}},
		{exp: "{a:1,b:2,c:3}.accept((k,v)->v>1)", res: Map{m: listMap.ListMap[Value]{
			{Key: "b", Value: Int(2)},
			{Key: "c", Value: Int(3)},
		}}},
		{exp: "{a:1,b:2,c:3}.list()", res: NewList(
			Map{m: listMap.ListMap[Value]{
				{Key: "key", Value: String("a")},
				{Key: "value", Value: Int(1)},
			}},
			Map{m: listMap.ListMap[Value]{
				{Key: "key", Value: String("b")},
				{Key: "value", Value: Int(2)},
			}},
			Map{m: listMap.ListMap[Value]{
				{Key: "key", Value: String("c")},
				{Key: "value", Value: Int(3)},
			}},
		)},
		{exp: "{a:1,b:2,c:3,d:-1}.size()", res: Int(4)},
		{exp: "{a:1,b:2}.replace(m->m.a+m.b)", res: Int(3)},
		{exp: "{a:1,b:2}.isAvail(\"a\")", res: Bool(true)},
		{exp: "{a:1,b:2}.isAvail(\"c\")", res: Bool(false)},
		{exp: "{a:1,b:2}.get(\"a\")", res: Int(1)},
		{exp: "\"\"+{a:1,b:2}.put(\"c\",3)", res: String("{c:3, a:1, b:2}")},
		{exp: "{a:1,b:2}.put(\"c\",3).c", res: Int(3)},
		{exp: "{a:1,b:2}.put(\"c\",3).b", res: Int(2)},
		{exp: "{a:1,b:2}.put(\"c\",3).size()", res: Int(3)},
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
			assert.Equalf(t, test.want, test.a.Equals(test.b), "Equals(%v)", test.b)
		})
	}
}
