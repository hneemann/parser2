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

func (v Map) Storage() MapStorage {
	return v.m
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

func (v Map) ToString() (string, error) {
	var b bytes.Buffer
	b.WriteString("{")
	first := true
	var innerErr error
	v.m.Iter(func(key string, v Value) bool {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(key)
		b.WriteString(":")
		s, err := v.ToString()
		if err != nil {
			innerErr = err
			return false
		}
		b.WriteString(s)
		return true
	})
	if innerErr != nil {
		return "", innerErr
	}
	b.WriteString("}")
	return b.String(), nil
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

func (v Map) Equals(other Map) (bool, error) {
	if v.Size() != other.Size() {
		return false, nil
	}
	equal := true
	var innerErr error
	v.m.Iter(func(key string, v Value) bool {
		if o, ok := other.Get(key); ok {
			b, err := Equal(o, v)
			if err != nil {
				innerErr = err
				return false
			}
			if !b {
				equal = false
				return false
			}
		} else {
			equal = false
			return false
		}
		return true
	})
	return equal, innerErr
}

func (v Map) Accept(st funcGen.Stack[Value]) (Map, error) {
	f, err := ToFunc("accept", st, 1, 2)
	if err != nil {
		return Map{}, err
	}
	newMap := listMap.New[Value](v.m.Size())
	var innerErr error
	v.m.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		var value Value
		value, innerErr = f.Func(st.CreateFrame(2), nil)
		if innerErr != nil {
			return false
		}
		if cond, ok := value.ToBool(); ok {
			if cond {
				newMap = newMap.Append(key, v)
			}
		} else {
			innerErr = fmt.Errorf("function in accept does not return a bool")
			return false
		}
		return true
	})
	if innerErr != nil {
		return Map{}, innerErr
	}
	return Map{m: newMap}, nil
}

func (v Map) Map(st funcGen.Stack[Value]) (Map, error) {
	f, err := ToFunc("map", st, 1, 2)
	if err != nil {
		return Map{}, err
	}
	newMap := listMap.New[Value](v.m.Size())
	var innerErr error
	v.m.Iter(func(key string, v Value) bool {
		st.Push(String(key))
		st.Push(v)
		var value Value
		value, innerErr = f.Func(st.CreateFrame(2), nil)
		if innerErr != nil {
			return false
		}
		newMap = newMap.Append(key, value)
		return true
	})
	if innerErr != nil {
		return Map{}, innerErr
	}
	return Map{m: newMap}, nil
}

func (v Map) Replace(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("replace", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return f.Eval(st, v)
}

func (v Map) List() *List {
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) (bool, error) {
			v.m.Iter(func(key string, v Value) bool {
				return yield(NewMap(listMap.New[Value](2).
					Append("key", String(key)).
					Append("value", v)))
			})
			return true, nil
		}
	})
}

func (v Map) Get(key string) (Value, bool) {
	return v.m.Get(key)
}

func (v Map) IsAvail(stack funcGen.Stack[Value]) (Value, error) {
	for i := 1; i < stack.Size(); i++ {
		if key, ok := stack.Get(i).(String); ok {
			_, ok := v.m.Get(string(key))
			if !ok {
				return Bool(false), nil
			}
		} else {
			return nil, fmt.Errorf("isAvail requires a string as argument")
		}
	}
	return Bool(true), nil
}

func (v Map) ContainsKey(key String) Value {
	_, ok := v.m.Get(string(key))
	return Bool(ok)
}

func (v Map) GetM(stack funcGen.Stack[Value]) (Value, error) {
	if key, ok := stack.Get(1).(String); ok {
		if v, ok := v.m.Get(string(key)); ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("key %v not found in map", key)
		}
	}
	return nil, fmt.Errorf("get requires a string as argument")
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

func (v Map) PutM(stack funcGen.Stack[Value]) (Map, error) {
	if key, ok := stack.Get(1).(String); ok {
		val := stack.Get(2)
		return Map{AppendMap{key: string(key), value: val, parent: v.m}}, nil
	}
	return Map{}, fmt.Errorf("get requires a string as argument")
}

var MapMethods = MethodMap{
	"accept": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Accept(stack) }).
		SetMethodDescription("func(key, value) bool",
			"Accept takes a function as argument and returns a new map with all entries for which the function returns true."),
	"map": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Map(stack) }).
		SetMethodDescription("func(key, value) value",
			"Map takes a function as argument and returns a new map with the same keys and all values replaced by the function."),
	"replace": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Replace(stack) }).
		SetMethodDescription("func(map) value",
			"Replace takes a function as argument and returns the result of the function. "+
				"The function is called with the map as argument."),
	"list": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.List(), nil }).
		SetMethodDescription("Returns a list of maps with the key and value of each entry in the map."),
	"size": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return Int(m.Size()), nil }).
		SetMethodDescription("Returns the number of entries in the map."),
	"string": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) {
		s, err := m.ToString()
		return String(s), err
	}).
		SetMethodDescription("Returns a string representation of the map."),
	"isAvail": MethodAtType(-1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.IsAvail(stack) }).
		SetMethodDescription("key", "Returns true if the key is available in the map."),
	"get": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.GetM(stack) }).
		SetMethodDescription("key", "Returns the value for the given key."),
	"put": MethodAtType(2, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.PutM(stack) }).
		SetMethodDescription("key", "value",
			"Returns a new map with the given key and value added. The original map is not changed."),
}

func (v Map) GetMethod(name string) (funcGen.Function[Value], error) {
	return MapMethods.Get(name)
}
