package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"sort"
)

// NewListConvert creates a list containing the given elements if the elements
// do not implement the Value interface. The given function converts the type.
func NewListConvert[I any](conv func(I) Value, items ...I) *List {
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) bool {
			for _, item := range items {
				if !yield(conv(item)) {
					return false
				}
			}
			return true
		}
	})
}

// NewList creates a new list containing the given elements
func NewList(items ...Value) *List {
	return &List{items: items, itemsPresent: true, iterable: createSliceIterable(items)}
}

func createSliceIterable(items []Value) iterator.Iterable[Value] {
	return func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) bool {
			for _, item := range items {
				if !yield(item) {
					return false
				}
			}
			return true
		}
	}
}

// NewListFromIterable creates a list based on the given Iterable
func NewListFromIterable(li iterator.Iterable[Value]) *List {
	return &List{iterable: li, itemsPresent: false}
}

// List represents a list of values
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

func (l *List) String() string {
	var b bytes.Buffer
	b.WriteString("[")
	first := true
	l.iterable()(func(v Value) bool {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
		return true
	})
	b.WriteString("]")
	return b.String()
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
		l.iterable = createSliceIterable(it)
	}
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

func (l *List) Iterator() iterator.Iterator[Value] {
	return l.iterable()
}

// ToSlice returns the list elements as a slice
func (l *List) ToSlice() []Value {
	l.Eval()
	return l.items[0:len(l.items):len(l.items)]
}

// CopyToSlice creates a slice copy of all elements
func (l *List) CopyToSlice() []Value {
	l.Eval()
	co := make([]Value, len(l.items))
	copy(co, l.items)
	return co
}

// Append creates a new list with a single element appended
// The original list remains unchanged while appending element
// by element is still efficient.
func (l *List) Append(st funcGen.Stack[Value]) *List {
	l.Eval()
	newList := append(l.items, st.Get(1))
	// Guarantee a copy operation the next time append is called on this
	// list, which is only a rare special case, as the new list is usually
	// appended to.
	if len(l.items) != cap(l.items) {
		l.items = l.items[:len(l.items):len(l.items)]
	}
	return NewList(newList...)
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
			panic(fmt.Errorf("%d. argument of %s needs to be a closure with %d arguments", n, name, args))
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

func (l *List) Compact(st funcGen.Stack[Value]) *List {
	f := toFunc("compact", st, 1, 1)
	return NewListFromIterable(iterator.Compact[Value](l.iterable, func(a, b Value) bool {
		aVal := f.Eval(st, a)
		bVal := f.Eval(st, b)
		return Equal(aVal, bVal)
	}))
}

func (l *List) Cross(st funcGen.Stack[Value]) *List {
	other := st.Get(1)
	f := toFunc("cross", st, 2, 2)
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Cross[Value, Value](l.iterable, otherList.iterable, func(a, b Value) Value {
			st.Push(a)
			st.Push(b)
			return f.Func(st.CreateFrame(2), nil)
		}))
	} else {
		panic("first argument in cross needs to be a list")
	}
}

func (l *List) Merge(st funcGen.Stack[Value]) *List {
	other := st.Get(1)
	f := toFunc("merge", st, 2, 2)
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Merge[Value](l.iterable, otherList.iterable, func(a, b Value) bool {
			st.Push(a)
			st.Push(b)
			if less, ok := f.Func(st.CreateFrame(2), nil).ToBool(); ok {
				return less
			} else {
				panic("closure in merge needs to return a bool, (a<b)")
			}
		}))
	} else {
		panic("first argument in merge needs to be a list")
	}
}

func (l *List) First() Value {
	if l.itemsPresent {
		if len(l.items) > 0 {
			return l.items[0]
		}
	} else {
		var first Value
		found := false
		l.iterable()(func(value Value) bool {
			first = value
			found = true
			return false
		})
		if found {
			return first
		}
	}
	panic("error in first, no items in list")
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

type SortableLess struct {
	items []Value
	st    funcGen.Stack[Value]
	less  funcGen.Function[Value]
}

func (s SortableLess) Len() int {
	return len(s.items)
}

func (s SortableLess) Less(i, j int) bool {
	s.st.Push(s.items[i])
	s.st.Push(s.items[j])
	if l, ok := s.less.Func(s.st.CreateFrame(2), nil).ToBool(); ok {
		return l
	} else {
		panic("closure in order needs to return a bool")
	}
}

func (s SortableLess) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (l *List) OrderLess(st funcGen.Stack[Value]) *List {
	f := toFunc("orderLess", st, 1, 2)
	items := l.CopyToSlice()
	sort.Sort(SortableLess{items: items, st: st, less: f})
	return NewList(items...)
}

type Sortable struct {
	items    []Value
	rev      bool
	st       funcGen.Stack[Value]
	pickFunc funcGen.Function[Value]
}

func (s Sortable) Len() int {
	return len(s.items)
}

func (s Sortable) pick(i int) Value {
	s.st.Push(s.items[i])
	return s.pickFunc.Func(s.st.CreateFrame(1), nil)
}

func (s Sortable) Less(i, j int) bool {
	if s.rev {
		return Less(s.pick(j), s.pick(i))
	} else {
		return Less(s.pick(i), s.pick(j))
	}
}

func (s Sortable) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (l *List) Order(st funcGen.Stack[Value], rev bool) *List {
	f := toFunc("order", st, 1, 1)
	items := l.CopyToSlice()
	sort.Sort(Sortable{items: items, rev: rev, st: st, pickFunc: f})
	return NewList(items...)
}

func (l *List) Reverse() *List {
	items := l.CopyToSlice()
	//reverse items
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return NewList(items...)
}

func (l *List) Combine(st funcGen.Stack[Value]) *List {
	f := toFunc("combine", st, 1, 2)
	return NewListFromIterable(iterator.Combine[Value, Value](l.iterable, func(a, b Value) Value {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	}))
}

func (l *List) Combine3(st funcGen.Stack[Value]) *List {
	f := toFunc("combine3", st, 1, 3)
	return NewListFromIterable(iterator.Combine3[Value, Value](l.iterable, func(a, b, c Value) Value {
		st.Push(a)
		st.Push(b)
		st.Push(c)
		return f.Func(st.CreateFrame(3), nil)
	}))
}

func (l *List) CombineN(st funcGen.Stack[Value]) *List {
	if n, ok := st.Get(1).ToInt(); ok {
		f := toFunc("combineN", st, 2, 2)
		return NewListFromIterable(iterator.CombineN[Value, Value](l.iterable, n, func(i0 int, i []Value) Value {
			st.Push(Int(i0))
			st.Push(NewList(i...))
			return f.Func(st.CreateFrame(2), nil)
		}))
	}
	panic("first argument in combineN needs to be an int")
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

func (l *List) IIrCombine(st funcGen.Stack[Value]) *List {
	initial := toFunc("iirCombine", st, 1, 1)
	function := toFunc("iirCombine", st, 2, 3)
	return NewListFromIterable(iterator.IirMap[Value, Value](l.iterable,
		func(item Value) Value {
			return initial.Eval(st, item)
		},
		func(item Value, lastItem Value, last Value) Value {
			st.Push(lastItem)
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(3), nil)
		}))
}

func (l *List) Visit(st funcGen.Stack[Value]) Value {
	visitor := st.Get(1)
	function := toFunc("visit", st, 2, 2)
	l.iterable()(func(value Value) bool {
		st.Push(visitor)
		st.Push(value)
		visitor = function.Func(st.CreateFrame(2), nil)
		return true
	})
	return visitor
}

func (l *List) Present(st funcGen.Stack[Value]) Value {
	function := toFunc("present", st, 1, 1)
	isPresent := false
	l.iterable()(func(value Value) bool {
		st.Push(value)
		if pr, ok := function.Func(st.CreateFrame(1), nil).ToBool(); ok {
			if pr {
				isPresent = true
				return false
			}
		} else {
			panic("closure in present needs to return a bool")
		}
		return true
	})
	return Bool(isPresent)
}

func (l *List) Top(st funcGen.Stack[Value]) *List {
	if i, ok := st.Get(1).ToInt(); ok {
		return NewListFromIterable(iterator.FirstN[Value](l.iterable, i))
	}
	panic("error in top, no int given")
}

func (l *List) Skip(st funcGen.Stack[Value]) *List {
	if i, ok := st.Get(1).ToInt(); ok {
		return NewListFromIterable(iterator.Skip[Value](l.iterable, i))
	}
	panic("error in skip, no int given")
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

func (l *List) MapReduce(st funcGen.Stack[Value]) Value {
	initial := st.Get(1)
	f := toFunc("mapReduce", st, 2, 2)
	return iterator.MapReduce(l.iterable, initial, func(s Value, v Value) Value {
		st.Push(s)
		st.Push(v)
		return f.Func(st.CreateFrame(2), nil)
	})
}

func (l *List) MinMax(st funcGen.Stack[Value]) Value {
	f := toFunc("minMax", st, 1, 1)
	first := true
	var minVal Value = Int(0)
	var maxVal Value = Int(0)
	l.Iterator()(func(value Value) bool {
		st.Push(value)
		r := f.Func(st.CreateFrame(1), nil)
		if first {
			first = false
			minVal = r
			maxVal = r
		} else {
			if Less(r, minVal) {
				minVal = r
			}
			if Less(maxVal, r) {
				maxVal = r
			}
		}
		return true
	})
	return NewMap(listMap.New[Value](3).
		Append("min", minVal).
		Append("max", maxVal).
		Append("valid", Bool(!first)))
}

func (l *List) Replace(st funcGen.Stack[Value]) Value {
	f := toFunc("replace", st, 1, 1)
	return f.Eval(st, l)
}

func (l *List) Number(st funcGen.Stack[Value]) *List {
	f := toFunc("number", st, 1, 2)
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) bool {
			n := Int(0)
			return l.iterable()(func(value Value) bool {
				st.Push(n)
				st.Push(value)
				n++
				return yield(f.Func(st.CreateFrame(2), nil))
			})
		}
	})
}

func (l *List) GroupByString(st funcGen.Stack[Value]) *List {
	keyFunc := toFunc("groupByString", st, 1, 1)
	return groupBy(l, func(value Value) Value {
		st.Push(value)
		key := keyFunc.Func(st.CreateFrame(1), nil)
		return String(key.String())
	})
}

func (l *List) GroupByInt(st funcGen.Stack[Value]) *List {
	keyFunc := toFunc("groupByInt", st, 1, 1)
	return groupBy(l, func(value Value) Value {
		st.Push(value)
		key := keyFunc.Func(st.CreateFrame(1), nil)
		if i, ok := key.ToInt(); ok {
			return Int(i)
		} else {
			panic("groupByInt requires an int as key")
		}
	})
}

func groupBy(list *List, keyFunc func(Value) Value) *List {
	m := make(map[Value]*[]Value)
	list.iterable()(func(value Value) bool {
		key := keyFunc(value)
		if l, ok := m[key]; ok {
			*l = append(*l, value)
		} else {
			ll := []Value{value}
			m[key] = &ll
		}
		return true
	})
	result := make([]Value, 0, len(m))
	for k, v := range m {
		result = append(result, Map{listMap.New[Value](2).
			Append("key", k).
			Append("value", NewList(*v...))})
	}
	return NewList(result...)
}

func (l *List) UniqueString(st funcGen.Stack[Value]) *List {
	keyFunc := toFunc("uniqueString", st, 1, 1)
	return unique(l, func(value Value) Value {
		st.Push(value)
		key := keyFunc.Func(st.CreateFrame(1), nil)
		return String(key.String())
	})
}

func (l *List) UniqueInt(st funcGen.Stack[Value]) *List {
	keyFunc := toFunc("uniqueInt", st, 1, 1)
	return unique(l, func(value Value) Value {
		st.Push(value)
		key := keyFunc.Func(st.CreateFrame(1), nil)
		if i, ok := key.ToInt(); ok {
			return Int(i)
		} else {
			panic("uniqueInt requires an int as key")
		}
	})
}

func unique(list *List, keyFunc func(Value) Value) *List {
	m := make(map[Value]struct{})
	list.iterable()(func(value Value) bool {
		key := keyFunc(value)
		m[key] = struct{}{}
		return true
	})
	result := make([]Value, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return NewList(result...)
}

func (l *List) MovingWindow(st funcGen.Stack[Value]) *List {
	f := toFunc("movingWindow", st, 1, 1)
	items := l.ToSlice()
	values := make([]float64, len(items))
	for i, elem := range items {
		if float, ok := f.Eval(st, elem).ToFloat(); ok {
			values[i] = float
		} else {
			panic("closure in movingWindow needs to return a float")
		}
	}

	var mainList []Value
	startIndex := 0
	for i, val := range values {
		for math.Abs(val-values[startIndex]) > 1 {
			startIndex++
		}
		mainList = append(mainList, NewList(items[startIndex:i+1:i+1]...))
	}
	return NewList(mainList...)
}

func (l *List) containsItem(item Value) bool {
	found := false
	l.iterable()(func(value Value) bool {
		if Equal(item, value) {
			found = true
			return false
		}
		return true
	})
	return found
}

func (l *List) containsAllItems(lookForList *List) bool {
	lookFor := lookForList.CopyToSlice()

	if l.itemsPresent && len(l.items) < len(lookFor) {
		return false
	}

	l.iterable()(func(value Value) bool {
		for i, lf := range lookFor {
			if Equal(lf, value) {
				lookFor = append(lookFor[0:i], lookFor[i+1:]...)
				break
			}
		}
		return len(lookFor) > 0
	})
	return len(lookFor) == 0
}

var ListMethods = MethodMap{
	"accept":        MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Accept(stack) }),
	"map":           MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Map(stack) }),
	"reduce":        MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Reduce(stack) }),
	"mapReduce":     MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.MapReduce(stack) }),
	"minMax":        MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MinMax(stack) }),
	"replace":       MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Replace(stack) }),
	"combine":       MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Combine(stack) }),
	"combine3":      MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Combine3(stack) }),
	"combineN":      MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.CombineN(stack) }),
	"multiUse":      MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MultiUse(stack) }),
	"indexOf":       MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.IndexOf(stack) }),
	"groupByString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.GroupByString(stack) }),
	"groupByInt":    MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.GroupByInt(stack) }),
	"uniqueString":  MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.UniqueString(stack) }),
	"uniqueInt":     MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.UniqueInt(stack) }),
	"compact":       MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Compact(stack) }),
	"cross":         MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Cross(stack) }),
	"merge":         MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Merge(stack) }),
	"order":         MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Order(stack, false) }),
	"orderRev":      MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Order(stack, true) }),
	"orderLess":     MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.OrderLess(stack) }),
	"reverse":       MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return list.Reverse() }),
	"append":        MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Append(stack) }),
	"iir":           MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.IIr(stack) }),
	"iirCombine":    MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.IIrCombine(stack) }),
	"visit":         MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Visit(stack) }),
	"top":           MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Top(stack) }),
	"skip":          MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Skip(stack) }),
	"number":        MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Number(stack) }),
	"present":       MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Present(stack) }),
	"size":          MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return Int(list.Size()) }),
	"first":         MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return list.First() }),
	"string":        MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return String(list.String()) }),
	"movingWindow":  MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MovingWindow(stack) }),
}

func (l *List) GetMethod(name string) (funcGen.Function[Value], error) {
	return ListMethods.Get(name)
}
