package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"sort"
)

// MapStorage is the abstraction of a map
type MapStorage interface {
	//Get returns the value for the given key and true if the key is present
	Get(key string) (Value, bool)
	//Iter iterates over the map
	Iter(yield func(key string, v Value) bool) bool
	//Size returns the number of entries in the map
	Size() int
}

type emptyMapStorage struct {
}

func (e emptyMapStorage) Get(string) (Value, bool) {
	return Int(0), false
}

func (e emptyMapStorage) Iter(func(key string, v Value) bool) bool {
	return true
}

func (e emptyMapStorage) Size() int {
	return 0
}

var EmptyMap = Map{emptyMapStorage{}}

// RealMap is a MapStorage implementation that uses a real map
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

// Map is a map of strings to values
// This is the type that is used in the parser
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

func (v Map) ToString(st funcGen.Stack[Value]) (string, error) {
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
		s, err := v.ToString(st)
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

func (v Map) String() string {
	s, err := v.ToString(funcGen.NewEmptyStack[Value]())
	if err != nil {
		return fmt.Sprintf("Map Error: %v", err)
	}
	return s
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

func (v Map) Equals(st funcGen.Stack[Value], other Map, equal funcGen.BoolFunc[Value]) (bool, error) {
	if v.Size() != other.Size() {
		return false, nil
	}
	eq := true
	var innerErr error
	v.m.Iter(func(key string, v Value) bool {
		if o, ok := other.Get(key); ok {
			b, err := equal(st, o, v)
			if err != nil {
				innerErr = err
				return false
			}
			if !b {
				eq = false
				return false
			}
		} else {
			eq = false
			return false
		}
		return true
	})
	return eq, innerErr
}

func (v Map) Accept(st funcGen.Stack[Value]) (Map, error) {
	f, err := ToFunc("accept", st, 1, 2)
	if err != nil {
		return EmptyMap, err
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
		return EmptyMap, innerErr
	}
	return Map{m: newMap}, nil
}

func (v Map) Map(st funcGen.Stack[Value]) (Map, error) {
	f, err := ToFunc("map", st, 1, 2)
	if err != nil {
		return EmptyMap, err
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
		return EmptyMap, innerErr
	}
	return Map{m: newMap}, nil
}

func (v Map) ReplaceMap(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("replace", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return f.Eval(st, v)
}

func (v Map) List() *List {
	return NewListFromIterable(func(st funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		var err error
		v.m.Iter(func(key string, v Value) bool {
			err = yield(NewMap(listMap.New[Value](2).
				Append("key", String(key)).
				Append("value", v)))
			if err != nil {
				return false
			}
			return true
		})
		if err != nil && err != iterator.SBC {
			return err
		}
		return nil
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
			return nil, errors.New("isAvail requires a string as argument")
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
		k := string(key)
		if v, ok := v.m.Get(k); ok {
			return v, nil
		} else {
			return nil, parser2.NewNotFoundError(k, fmt.Errorf("key '%v' not found in map", k))
		}
	}
	return nil, errors.New("get requires a string as argument")
}

// AppendMap is a MapStorage implementation that adds a key/value pair to an
// existing MapStorage. It is used in the Put method. This approach is much more
// efficient than creating a new real map as long as the number of keys is small.
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
		k := string(key)
		if _, ok := v.Get(k); ok {
			return EmptyMap, fmt.Errorf("key '%s' already present in map", k)
		}
		val := stack.Get(2)
		return Map{AppendMap{key: k, value: val, parent: v.m}}, nil
	}
	return EmptyMap, errors.New("put requires a string as first argument")
}

// MergeMap is a MapStorage implementation that merges two MapStorages. It is
// used in the Merge method. It is more efficient than creating a new real map as
// long as the number of keys is small.
type MergeMap struct {
	a, b MapStorage
}

func (m MergeMap) Get(key string) (Value, bool) {
	if e, ok := m.a.Get(key); ok {
		return e, true
	}
	return m.b.Get(key)
}

func (m MergeMap) Iter(yield func(key string, v Value) bool) bool {
	if m.a.Iter(yield) {
		return m.b.Iter(yield)
	}
	return true
}

func (m MergeMap) Size() int {
	return m.a.Size() + m.b.Size()
}

type ReplaceMap struct {
	orig, rep MapStorage
	depth     int
}

func (m ReplaceMap) Get(key string) (Value, bool) {
	if e, ok := m.rep.Get(key); ok {
		return e, true
	}
	return m.orig.Get(key)
}

func (m ReplaceMap) Iter(yield func(key string, v Value) bool) bool {
	m.orig.Iter(func(key string, v Value) bool {
		if rep, ok := m.rep.Get(key); ok {
			return yield(key, rep)
		}
		return yield(key, v)
	})
	return true
}

func (m ReplaceMap) Size() int {
	return m.orig.Size()
}

func (m ReplaceMap) createFlat() MapStorage {
	size := m.Size()
	if size > 20 {
		rm := make(RealMap, size)
		m.Iter(func(key string, v Value) bool {
			rm[key] = v
			return true
		})
		return rm
	} else {
		lm := listMap.New[Value](size)
		m.Iter(func(key string, v Value) bool {
			lm = lm.Append(key, v)
			return true
		})
		return lm
	}
}

func (v Map) Merge(other Map) (Map, error) {
	var exists string
	other.Iter(func(key string, val Value) bool {
		if _, ok := v.Get(key); ok {
			exists = key
			return false
		}
		return true
	})
	if exists != "" {
		return EmptyMap, fmt.Errorf("first map already contains key '%s'", exists)
	}
	return Map{MergeMap{a: v.m, b: other.m}}, nil
}

func (v Map) Combine(st funcGen.Stack[Value]) (Map, error) {
	fun, err := ToFunc("combine", st, 2, 2)
	if err != nil {
		return EmptyMap, err
	}
	if other, ok := st.Get(1).ToMap(); ok {
		result := listMap.New[Value](v.Size())
		var innerErr error
		v.Iter(func(key string, val Value) bool {
			if o, ok := other.Get(key); ok {
				st.Push(val)
				st.Push(o)
				r, err := fun.Func(st.CreateFrame(2), nil)
				if err != nil {
					innerErr = err
					return false
				}
				result = result.Append(key, r)
			} else {
				innerErr = fmt.Errorf("key '%s' not present in second map", key)
				return false
			}
			return true
		})
		if innerErr != nil {
			return EmptyMap, innerErr
		}
		return Map{result}, nil
	} else {
		return EmptyMap, errors.New("combine requires a map as first argument")
	}
}

func (v Map) Replace(stack funcGen.Stack[Value]) (Map, error) {
	f, err := ToFunc("replace", stack, 1, 1)
	if err != nil {
		return EmptyMap, err
	}

	repMap, err := f.Eval(stack, v)
	if err != nil {
		return EmptyMap, err
	}

	if rep, ok := repMap.ToMap(); ok {
		depth := 0
		if r, ok := v.m.(ReplaceMap); ok {
			depth = r.depth
		}
		if r, ok := rep.m.(ReplaceMap); ok {
			if r.depth > depth {
				depth = r.depth
			}
		}
		rm := ReplaceMap{
			orig:  v.m,
			rep:   rep,
			depth: depth + 1,
		}
		if depth >= 10 {
			return NewMap(rm.createFlat()), nil
		}
		return NewMap(rm), nil
	}
	return EmptyMap, errors.New("the result of the function passed to replace must be a map")
}
func createMapMethods() MethodMap {
	return MethodMap{
		"eval": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Eval() }).
			SetMethodDescription("Evaluates the map to a real hash map. This is more efficient if the map has many " +
				"keys and the associated values are requested often."),
		"accept": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Accept(stack) }).
			SetMethodDescription("func(key, value) bool",
				"Accept takes a function as argument and returns a new map with all entries for which the function returns true."),
		"map": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Map(stack) }).
			SetMethodDescription("func(key, value) value",
				"Map takes a function as argument and returns a new map with the same keys and all values replaced by the function."),
		"replaceMap": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.ReplaceMap(stack) }).
			SetMethodDescription("func(map) value",
				"Takes a function as argument and returns the result of the function. "+
					"The function is called with the map as argument."),
		"list": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.List(), nil }).
			SetMethodDescription("Returns a list of maps with the key and value of each entry in the map."),
		"size": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return Int(m.Size()), nil }).
			SetMethodDescription("Returns the number of entries in the map."),
		"string": MethodAtType(0, func(m Map, stack funcGen.Stack[Value]) (Value, error) {
			s, err := m.ToString(stack)
			return String(s), err
		}).
			SetMethodDescription("Returns a string representation of the map."),
		"isAvail": MethodAtType(-1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.IsAvail(stack) }).
			SetMethodDescription("key", "Returns true if the key is available in the map."),
		"get": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.GetM(stack) }).
			SetMethodDescription("key", "Returns the value for the given key."),
		"put": MethodAtType(2, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.PutM(stack) }).
			SetMethodDescription("key", "value",
				"Returns a new map with the given key and value added."),
		"replace": MethodAtType(1, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Replace(stack) }).
			SetMethodDescription("func(map) rep_map",
				"Calls the given function with the original map as argument and returns a 'replacement' map. "+
					"The key/values from the 'replacement' map are used to replace the key/values in the original map."),
		"combine": MethodAtType(2, func(m Map, stack funcGen.Stack[Value]) (Value, error) { return m.Combine(stack) }).
			SetMethodDescription("other_map", "func(a,b) r",
				"Combines the two maps with the given funktion to a new map. The function is called for each key that is in both maps. "+
					"The first argument is the value of the first map and the second argument is the value of the second map. "+
					"The function must return a value that is used as value in the new map."),
	}
}

func (v Map) GetType() Type {
	return MapTypeId
}

func (v Map) keyListDescription() string {

	type desc interface {
		KeyListDescription() string
	}

	if d, ok := v.m.(desc); ok {
		return d.KeyListDescription()
	}

	var keys []string
	v.Iter(func(key string, v Value) bool {
		keys = append(keys, key)
		return true
	})
	sort.Strings(keys)

	var b bytes.Buffer
	for _, key := range keys {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
	}
	return b.String()
}

func (v Map) Eval() (Value, error) {
	rm := make(RealMap)
	v.Iter(func(key string, v Value) bool {
		rm[key] = v
		return true
	})
	return NewMap(rm), nil
}
