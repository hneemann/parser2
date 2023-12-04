package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
)

type MapStorage interface {
	Get(key string) (Value, bool)
	Iter(yield func(key string, v Value) bool) bool
	Size() int
}

type RealMap map[string]Value

func (s RealMap) Get(key string) (Value, bool) {
	v, ok := s[key]
	return v, ok
}

func (s RealMap) Iter(yield func(key string, v Value) bool) bool {
	for k, v := range s {
		if !yield(k, v) {
			return false
		}
	}
	return true
}

func (s RealMap) Size() int {
	return len(s)
}

type Map struct {
	m MapStorage
}

func NewMap(m MapStorage) Map {
	return Map{m: m}
}

func (v Map) Iter(yield func(key string, v Value) bool) bool {
	return v.m.Iter(yield)
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

func (v Map) String() string {
	var b bytes.Buffer
	b.WriteString("{")
	first := true
	v.m.Iter(func(key string, v Value) bool {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(v.String())
		return true
	})
	b.WriteString("}")
	return b.String()
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
	return v.m.Size()
}

func (v Map) Equals(other Map) bool {
	if v.Size() != other.Size() {
		return false
	}
	equal := true
	v.m.Iter(func(key string, v Value) bool {
		if o, ok := other.Get(key); ok {
			if !Equal(o, v) {
				equal = false
				return false
			}
		} else {
			equal = false
			return false
		}
		return true
	})
	return equal
}

func (v Map) Accept(st funcGen.Stack[Value]) Map {
	f := toFunc("accept", st, 1, 2)
	newMap := listMap.New[Value](v.m.Size())
	v.m.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		if cond, ok := f.Func(st.CreateFrame(2), nil).ToBool(); ok {
			if cond {
				newMap = newMap.Append(key, v)
			}
		} else {
			panic(fmt.Errorf("function in accept does not return a bool"))
		}
		return true
	})
	return Map{m: newMap}
}

func (v Map) Map(st funcGen.Stack[Value]) Map {
	f := toFunc("map", st, 1, 2)
	newMap := listMap.New[Value](v.m.Size())
	v.m.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		newMap = newMap.Append(key, f.Func(st.CreateFrame(2), nil))
		return true
	})
	return Map{m: newMap}
}

func (v Map) Replace(st funcGen.Stack[Value]) Value {
	f := toFunc("replace", st, 1, 1)
	return f.Eval(st, v)
}

func (v Map) List() *List {
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) bool {
			v.m.Iter(func(key string, v Value) bool {
				return yield(NewMap(listMap.New[Value](2).
					Append("key", String(key)).
					Append("value", v)))
			})
			return true
		}
	})
}

func (v Map) Get(key string) (Value, bool) {
	return v.m.Get(key)
}

func (v Map) IsAvail(stack funcGen.Stack[Value]) Value {
	if key, ok := stack.Get(1).(String); ok {
		_, ok := v.m.Get(string(key))
		return Bool(ok)
	}
	panic("isAvail requires a string as argument")
}

func (v Map) ContainsKey(key String) Value {
	_, ok := v.m.Get(string(key))
	return Bool(ok)
}

func (v Map) GetM(stack funcGen.Stack[Value]) Value {
	if key, ok := stack.Get(1).(String); ok {
		if v, ok := v.m.Get(string(key)); ok {
			return v
		} else {
			panic(fmt.Errorf("key %v not found in map", key))
		}
	}
	panic("get requires a string as argument")
}

type AppendMap struct {
	key    string
	value  Value
	parent MapStorage
}

func (a AppendMap) Get(key string) (Value, bool) {
	if key == a.key {
		return a.value, true
	} else {
		return a.parent.Get(key)
	}
}

func (a AppendMap) Iter(yield func(key string, v Value) bool) bool {
	if !yield(a.key, a.value) {
		return false
	} else {
		return a.parent.Iter(yield)
	}
}

func (a AppendMap) Size() int {
	return a.parent.Size() + 1
}

func (v Map) PutM(stack funcGen.Stack[Value]) Map {
	if key, ok := stack.Get(1).(String); ok {
		val := stack.Get(2)
		return Map{AppendMap{key: string(key), value: val, parent: v.m}}
	}
	panic("get requires a string as argument")
}

var MapMethods = MethodMap{
	"accept":  MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.Accept(stack) }),
	"map":     MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.Map(stack) }),
	"replace": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.Replace(stack) }),
	"list":    MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) Value { return m.List() }),
	"size":    MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) Value { return Int(m.Size()) }),
	"string":  MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) Value { return String(m.String()) }),
	"isAvail": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.IsAvail(stack) }),
	"get":     MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) Value { return m.GetM(stack) }),
	"put":     MethodAtType(2, func(m Map, stack funcGen.Stack[Value]) Value { return m.PutM(stack) }),
}

func (v Map) GetMethod(name string) (funcGen.Function[Value], error) {
	return MapMethods.Get(name)
}
