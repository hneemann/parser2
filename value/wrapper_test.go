package value

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Example to show, how a simple struct can be converted
// to a map usable by the parser

// example struct
type testStruct struct {
	a int
	b int
	c float64
}

// creating a ToMap instance for the struct
var testStructToMap = NewToMap[testStruct]().
	Attr("a", func(t testStruct) Value { return Int(t.a) }).
	Attr("b", func(t testStruct) Value { return Int(t.b) }).
	Attr("c", func(t testStruct) Value { return Float(t.c) })

func TestToMap(t *testing.T) {
	ts := testStruct{
		a: 1,
		b: 2,
		c: 1.5,
	}

	// create a map from the struct
	m := testStructToMap.Create(ts)

	// test the map
	assert.Equal(t, 3, m.Size())

	a, aOk := m.Get("a")
	assert.True(t, aOk)
	assert.Equal(t, Int(1), a)
	b, bOk := m.Get("b")
	assert.True(t, bOk)
	assert.Equal(t, Int(2), b)
	c, cOk := m.Get("c")
	assert.True(t, cOk)
	assert.Equal(t, Float(1.5), c)

	var s string
	m.M.Iter(func(key string, v Value) bool {
		s += fmt.Sprintf("%s:%v ", key, v)
		return true
	})
	assert.Equal(t, "a:1 b:2 c:1.5 ", s)
}
