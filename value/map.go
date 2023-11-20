package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
)

type MapImplementation[V any] interface {
	Get(key string) (V, bool)
	Iter(yield func(key string, v Value) bool) bool
	Size() int
}

type SimpleMap map[string]Value

func (s SimpleMap) Get(key string) (Value, bool) {
	v, ok := s[key]
	return v, ok
}

func (s SimpleMap) Iter(yield func(key string, v Value) bool) bool {
	for k, v := range s {
		if !yield(k, v) {
			return false
		}
	}
	return true
}

func (s SimpleMap) Size() int {
	return len(s)
}

type Map struct {
	M MapImplementation[Value]
}

func (v Map) ToList() (*List, bool) {
	return nil, false
}

func (v Map) ToInt() (int, bool) {
	return 0, false
}

func (v Map) ToFloat() (float64, bool) {
	return 0, false
}

func (v Map) ToString() (string, bool) {
	var b bytes.Buffer
	b.WriteString("{")
	first := true
	v.M.Iter(func(key string, v Value) bool {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(key)
		b.WriteString(":")
		if s, ok := v.ToString(); ok {
			b.WriteString(s)
		} else {
			b.WriteString("?")
		}
		return true
	})
	b.WriteString("}")
	return b.String(), true
}

func (v Map) ToBool() (bool, bool) {
	return false, false
}

func (v Map) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (v Map) ToMap() (Map, bool) {
	return v, true
}
func (v Map) Size() int {
	return v.M.Size()
}

func (v Map) Accept(st funcGen.Stack[Value]) Map {
	f := toFunc("accept", st, 1, 2)
	newMap := listMap.New[Value](v.M.Size())
	v.M.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		if cond, ok := f.Func(st.CreateFrame(2), nil).ToBool(); ok {
			if cond {
				newMap.Put(key, v)
			}
		} else {
			panic(fmt.Errorf("closure in accept does not return a bool"))
		}
		return true
	})
	return Map{M: newMap}
}

func (v Map) Map(st funcGen.Stack[Value]) Map {
	f := toFunc("map", st, 1, 2)
	newMap := listMap.New[Value](v.M.Size())
	v.M.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		newMap.Put(key, f.Func(st.CreateFrame(2), nil))
		return true
	})
	return Map{M: newMap}
}

func (v Map) Replace(st funcGen.Stack[Value]) Value {
	f := toFunc("replace", st, 1, 1)
	return f.Eval(st, v)
}

func (v Map) List() *List {
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(f func(Value) bool) bool {
			v.M.Iter(func(key string, v Value) bool {
				m := listMap.New[Value](2)
				m.Put("key", String(key))
				m.Put("value", v)
				f(Map{m})
				return true
			})
			return true
		}
	})
}

func methodAtMap(args int, method func(m Map, stack funcGen.Stack[Value]) Value) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) Value {
		if m, ok := stack.Get(0).ToMap(); ok {
			return method(m, stack)
		}
		panic("call of map method on non map")
	}, Args: args, IsPure: true}
}

var MapMethods = map[string]funcGen.Function[Value]{
	"accept":  methodAtMap(2, func(m Map, stack funcGen.Stack[Value]) Value { return m.Accept(stack) }),
	"map":     methodAtMap(2, func(m Map, stack funcGen.Stack[Value]) Value { return m.Map(stack) }),
	"replace": methodAtMap(2, func(m Map, stack funcGen.Stack[Value]) Value { return m.Replace(stack) }),
	"list":    methodAtMap(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.List() }),
	"size":    methodAtMap(1, func(m Map, stack funcGen.Stack[Value]) Value { return Int(m.Size()) }),
}

func (v Map) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := MapMethods[name]
	return m, ok
}

func (v Map) Get(key string) (Value, bool) {
	return v.M.Get(key)
}
