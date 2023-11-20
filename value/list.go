package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"sort"
)

func NewList(items ...Value) *List {
	return &List{items: items, itemsPresent: true, iterable: func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) bool {
			for _, item := range items {
				if !yield(item) {
					return false
				}
			}
			return true
		}
	}}
}

func NewListFromIterable(li iterator.Iterable[Value]) *List {
	return &List{iterable: li, itemsPresent: false}
}

type List struct {
	items        []Value
	itemsPresent bool
	iterable     iterator.Iterable[Value]
}

func (l *List) ToMap() (Map, bool) {
	return Map{}, false
}

func (l *List) ToInt() (int, bool) {
	return 0, false
}

func (l *List) ToFloat() (float64, bool) {
	return 0, false
}

func (l *List) ToString() (string, bool) {
	var b bytes.Buffer
	b.WriteString("[")
	first := true
	l.iterable()(func(v Value) bool {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		if s, ok := v.ToString(); ok {
			b.WriteString(s)
		} else {
			b.WriteString("?")
		}
		return true
	})
	b.WriteString("]")
	return b.String(), true
}

func (l *List) ToBool() (bool, bool) {
	return false, false
}

func (l *List) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (l *List) ToList() (*List, bool) {
	return l, true
}

func (l *List) Eval() {
	if !l.itemsPresent {
		var it []Value
		l.iterable()(func(value Value) bool {
			it = append(it, value)
			return true
		})
		l.items = it
		l.itemsPresent = true
	}
}

func (l *List) ToSlice() []Value {
	l.Eval()
	return l.items[0:len(l.items):len(l.items)]
}

func (l *List) Size() int {
	l.Eval()
	return len(l.items)
}

func toFunc(name string, st funcGen.Stack[Value], n int, args int) funcGen.Function[Value] {
	if c, ok := st.Get(n).ToClosure(); ok {
		if c.Args == args {
			return c
		} else {
			panic(fmt.Errorf("%d. argument of %s needs to be a closure with %d argoments", n, name, args))
		}
	} else {
		panic(fmt.Errorf("%d. argument of %s needs to be a closure", n, name))
	}
}

func (l *List) Accept(st funcGen.Stack[Value]) *List {
	f := toFunc("accept", st, 1, 1)
	return NewListFromIterable(iterator.FilterAuto[Value](l.iterable, func() func(v Value) bool {
		lst := funcGen.NewEmptyStack[Value]()
		return func(v Value) bool {
			if accept, ok := f.Eval(lst, v).ToBool(); ok {
				return accept
			}
			panic(fmt.Errorf("closure in accept does not return a bool"))
		}
	}))
}

func (l *List) Map(st funcGen.Stack[Value]) *List {
	f := toFunc("map", st, 1, 1)
	return NewListFromIterable(iterator.MapAuto[Value, Value](l.iterable, func() func(i int, v Value) Value {
		lst := funcGen.NewEmptyStack[Value]()
		return func(i int, v Value) Value {
			return f.Eval(lst, v)
		}
	}))
}

type Sortable struct {
	items []Value
	st    funcGen.Stack[Value]
	less  funcGen.Function[Value]
}

func (s Sortable) Len() int {
	return len(s.items)
}

func (s Sortable) Less(i, j int) bool {
	s.st.Push(s.items[i])
	s.st.Push(s.items[j])
	if l, ok := s.less.Func(s.st.CreateFrame(2), nil).ToBool(); ok {
		return l
	} else {
		panic("closure in order needs to return a bool")
	}
}

func (s Sortable) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (l *List) IndexOf(st funcGen.Stack[Value]) Int {
	v := st.Get(1)
	index := -1
	i := 0
	l.iterable()(func(value Value) bool {
		if Equal(v, value) {
			index = i
			return false
		}
		i++
		return true
	})
	return Int(index)
}

func (l *List) Order(st funcGen.Stack[Value]) *List {
	f := toFunc("order", st, 1, 2)

	items := l.ToSlice()
	itemsCopy := make([]Value, len(items))
	copy(itemsCopy, items)

	s := Sortable{
		items: itemsCopy,
		st:    st,
		less:  f,
	}

	sort.Sort(s)
	return NewList(itemsCopy...)
}

func (l *List) Combine(st funcGen.Stack[Value]) *List {
	f := toFunc("combine", st, 1, 2)
	return NewListFromIterable(iterator.Combine[Value, Value](l.iterable, func(a, b Value) Value {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	}))
}

func (l *List) IIr(st funcGen.Stack[Value]) *List {
	initial := toFunc("iir", st, 1, 1)
	function := toFunc("iir", st, 2, 2)
	return NewListFromIterable(iterator.IirMap[Value, Value](l.iterable,
		func(item Value) Value {
			return initial.Eval(st, item)
		},
		func(item Value, lastItem Value, last Value) Value {
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(2), nil)
		}))
}

func (l *List) Reduce(st funcGen.Stack[Value]) Value {
	f := toFunc("reduce", st, 1, 2)
	res, ok := iterator.Reduce[Value](l.iterable, func(a, b Value) Value {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	})
	if ok {
		return res
	}
	panic("error in reduce, no items in list")
}

func (l *List) Replace(st funcGen.Stack[Value]) Value {
	f := toFunc("replace", st, 1, 1)
	return f.Eval(st, l)
}

func (l *List) GroupBy(st funcGen.Stack[Value]) Map {
	keyFunc := toFunc("groupBy", st, 1, 1)
	valueFunc := toFunc("groupBy", st, 2, 1)
	m := make(map[string]*[]Value)
	l.iterable()(func(value Value) bool {
		k := keyFunc.Eval(st, value)
		v := valueFunc.Eval(st, value)
		if key, ok := k.ToString(); ok {
			if l, ok := m[key]; ok {
				*l = append(*l, v)
			} else {
				ll := []Value{v}
				m[key] = &ll
			}
		} else {
			panic("groupBy requires a string as key type")
		}
		return true
	})
	ma := make(SimpleMap)
	for k, l := range m {
		ma[k] = NewList(*l...)
	}
	return Map{ma}
}

func methodAtList(args int, method func(list *List, stack funcGen.Stack[Value]) Value) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) Value {
		if obj, ok := stack.Get(0).ToList(); ok {
			return method(obj, stack)
		}
		panic("call of list method on non list")
	}, Args: args, IsPure: true}
}

var ListMethods = map[string]funcGen.Function[Value]{
	"accept":  methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Accept(stack) }),
	"map":     methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Map(stack) }),
	"reduce":  methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Reduce(stack) }),
	"replace": methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Replace(stack) }),
	"combine": methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Combine(stack) }),
	"indexOf": methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.IndexOf(stack) }),
	"group":   methodAtList(3, func(list *List, stack funcGen.Stack[Value]) Value { return list.GroupBy(stack) }),
	"order":   methodAtList(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Order(stack) }),
	"iir":     methodAtList(3, func(list *List, stack funcGen.Stack[Value]) Value { return list.IIr(stack) }),
	"size":    methodAtList(1, func(list *List, stack funcGen.Stack[Value]) Value { return Int(list.Size()) }),
}

func (l *List) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := ListMethods[name]
	return m, ok
}

func (l *List) Equals(other *List) bool {
	a := l.ToSlice()
	b := other.ToSlice()
	if len(a) != len(b) {
		return false
	}
	for i, aa := range a {
		if !Equal(aa, b[i]) {
			return false
		}
	}
	return true
}
