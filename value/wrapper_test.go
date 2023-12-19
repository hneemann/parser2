package value

import (
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

	is := map[string]Value{}
	m.m.Iter(func(key string, v Value) bool {
		is[key] = v
		return true
	})
	assert.Equal(t, Map{RealMap{"a": Int(1), "b": Int(2), "c": Float(1.5)}}, Map{RealMap(is)})
}

type dataType struct {
	MyInt8   int8
	MyInt    int
	MyString string
	MyFloat  float64
}

func TestNewReflection(t *testing.T) {
	m := NewToMapReflection[dataType]()
	data := dataType{
		MyInt8:   8,
		MyInt:    7,
		MyString: "Hello",
		MyFloat:  1.1,
	}

	v := m.Create(data)
	assert.Equal(t, 4, v.Size())
	{
		val, ok := v.Get("MyInt8")
		assert.True(t, ok)
		assert.Equal(t, Int(8), val)
	}
	{
		val, ok := v.Get("MyInt")
		assert.True(t, ok)
		assert.Equal(t, Int(7), val)
	}
	{
		val, ok := v.Get("MyString")
		assert.True(t, ok)
		assert.Equal(t, String("Hello"), val)
	}
	{
		val, ok := v.Get("MyFloat")
		assert.True(t, ok)
		assert.Equal(t, Float(1.1), val)
	}
}

func Benchmark(b *testing.B) {

	data := dataType{
		MyInt8:   8,
		MyInt:    7,
		MyString: "Hello",
		MyFloat:  1.1,
	}

	type creator interface {
		Create(dataType) Map
	}

	benchmarks := []struct {
		name  string
		toMap creator
	}{
		{
			name: "manual",
			toMap: NewToMap[dataType]().
				Attr("MyInt", func(t dataType) Value { return Int(t.MyInt) }).
				Attr("MyString", func(t dataType) Value { return String(t.MyString) }).
				Attr("MyFloat", func(t dataType) Value { return Float(t.MyFloat) }),
		},
		{
			name:  "reflect",
			toMap: NewToMapReflection[dataType](),
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			m := bm.toMap.Create(data)
			for i := 0; i < b.N; i++ {
				for i := 0; i < 100; i++ {
					m.Get("MyInt")
					m.Get("MyString")
					m.Get("MyFloat")
				}
			}
		})
	}
}
