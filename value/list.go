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
			panic(fmt.Errorf("%d. argument of %s needs to be a function with %d arguments", n, name, args))
		}
	} else {
		panic(fmt.Errorf("%d. argument of %s needs to be a function", n, name))
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
			panic(fmt.Errorf("function in accept does not return a bool"))
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
				panic("function in merge needs to return a bool, (a<b)")
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
		panic("function in order needs to return a bool")
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
			st.Push(NewList(i...))
			return f.Func(st.CreateFrame(1), nil)
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
			panic("function in present needs to return a bool")
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
			Append("values", NewList(*v...))})
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
			panic("function in movingWindow needs to return a float")
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
	"accept": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Accept(stack) }).
		SetDescription("func(item) bool",
			"Filters the list by the given function. If the function returns true, the item is accepted, otherwise it is skipped."),
	"map": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Map(stack) }).
		SetDescription("func(item) newItem",
			"Maps the list by the given function. The function is called for each item in the list and the result is "+
				"added to the new list."),
	"reduce": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Reduce(stack) }).
		SetDescription("func(item, item) item",
			"Reduces the list by the given function. The function is called with the first two list items, and the result "+
				"is used as the first argument for the third item and so on."),
	"mapReduce": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.MapReduce(stack) }).
		SetDescription("initialSum", "func(sum, item) sum",
			"MapReduce reduces the list to a single value. The initial value is given as the first argument. The function "+
				"is called with the initial value and the first item, and the result is used as the first argument for the "+
				"second item and so on."),
	"minMax": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MinMax(stack) }).
		SetDescription("func(item) value",
			"Returns the minimum and maximum value of the list. The function is called for each item in the list and the "+
				"result is compared to the previous minimum and maximum."),
	"replace": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Replace(stack) }).
		SetDescription("func(list) newItem",
			"Replaces the list by the result of the given function. The function is called with the list as argument."),
	"combine": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Combine(stack) }).
		SetDescription("func(item, item) newItem",
			"Combines the list by the given function. The function is called for each pair of items in the list and the "+
				"result is added to the new list. "+
				"The resulting list is one item shorter than the original list."),
	"combine3": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Combine3(stack) }).
		SetDescription("func(item, item, item) newItem",
			"Combines the list by the given function. The function is called for each triplet of items in the list and "+
				"the result is added to the new list. "+
				"The resulting list is two items shorter than the original list."),
	"combineN": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.CombineN(stack) }).
		SetDescription("n", "func([item...]) newItem",
			"Combines the list by the given function. The function is called for each group of n items in the list and "+
				"the result is added to the new list. "+
				"The resulting list is n-1 items shorter than the original list."),
	"multiUse": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MultiUse(stack) }).
		SetDescription("{name: func(item) newItem...}",
			"MultiUse allows to use the list multiple times without storing or recomputing its elements. The first argument "+
				"is a map of functions. "+
				"All the functions are called with the list as argument and the result is returned in a map. "+
				"The keys in the result map are the same keys used to pass the functions."),
	"indexOf": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.IndexOf(stack) }).
		SetDescription("item",
			"Returns the index of the first occurrence of the given item in the list. If the item is not found, -1 is returned."),
	"groupByString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.GroupByString(stack) }).
		SetDescription("func(item) string", "Returns a list of lists grouped by the given function. "+
			"The function is called for each item in the list and the returned string is used as the key for the group. "+
			"The result is a list of maps with the keys 'key' and 'values'. The 'key' contains the string returned by the function "+
			"and 'values' contains a list of items that have the same key."),
	"groupByInt": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.GroupByInt(stack) }).
		SetDescription("func(item) int", "Returns a list of lists grouped by the given function. "+
			"The function is called for each item in the list and the returned integer is used as the key for the group. "+
			"The result is a list of maps with the keys 'key' and 'values'. The 'key' contains the integer returned by the function "+
			"and 'values' contains a list of items that have the same key."),
	"uniqueString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.UniqueString(stack) }).
		SetDescription("func(item) string", "Returns a list of unique strings returned by the given function."),
	"uniqueInt": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.UniqueInt(stack) }).
		SetDescription("func(item) int", "Returns a list of unique integers returned by the given function."),
	"compact": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Compact(stack) }).
		SetDescription("func(a,b) bool", "Returns a new list with the items compacted. "+
			"The function is called for each pair of items in the list and needs to return true if a=b holds."+
			"Compacting means that an item is removed if the function returns true for the item and the previous item."),
	"cross": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Cross(stack) }).
		SetDescription("other_list", "func(a,b) newItem",
			"Returns a new list with the given function applied to each pair of items in the list and the given list. "+
				"The function is called with an item from the first list and an item from the second list. "+
				"The length of the resulting list is the product of the lengths of the two lists."),
	"merge": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Merge(stack) }).
		SetDescription("other_list", "func(a,b) bool",
			"Returns a new list with the items of both lists combined. "+
				"The given function is called for the pair of the first, non processed items in both lists. If the "+
				"return value is true the value of the original list is taken, otherwise the item from the other list. "+
				"The is repeated until all items of both lists are processed. "+
				"If the function returns true if a<b holds and both lists are ordered, also the new list is ordered."),
	"order": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Order(stack, false) }).
		SetDescription("func(item) value",
			"Returns a new list with the items sorted in the order of the values returned by the given function. "+
				"The function is called for each item in the list and the returned values determine the order."),
	"orderRev": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Order(stack, true) }).
		SetDescription("func(item) value",
			"Returns a new list with the items sorted in the reverse order of the values returned by the given function. "+
				"The function is called for each item in the list and the returned values determine the order."),
	"orderLess": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.OrderLess(stack) }).
		SetDescription("func(a, a) bool",
			"Returns a new list with the items sorted by the given function. "+
				"The function is called for pairs of items in the list and the returned bool needs to be true if a<b holds."),
	"reverse": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return list.Reverse() }).
		SetDescription("Returns the list in reverse order."),
	"append": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Append(stack) }).
		SetDescription("item", "Returns a new list with the given item appended."),
	"iir": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.IIr(stack) }).
		SetDescription("func(first_item) first_new_item", "func(item, last_new_item) new_item",
			"Returns a new list with the given functions applied to the items in the list. "+
				"The first function is called with the first item in the list and returns the first item in the new list. "+
				"The second function is called with the remaining items in the list as the first argument, and the last new item. "+
				"For each subsequent item, the function is called with the item and the result of the previous call."),
	"iirCombine": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.IIrCombine(stack) }).
		SetDescription("func(first_item) first_new_item", "func(i0, i1, last_new_item) new_item",
			"Returns a new list with the given functions applied to the items in the list. "+
				"The first function is called with the first item in the list and returns the first item in the new list. "+
				"The second function is called with the remaining pairs of items in the list as the first two arguments, and the last new item. "+
				"For each subsequent item, the function is called with the the pair of items and the result of the previous call. "+
				"The item i0 is the item in front of i1."),
	"visit": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) Value { return list.Visit(stack) }).
		SetDescription("initial_visitor", "func(visitor, item) visitor",
			"Visits each item in the list with the given function. The function is called with the visitor and the item. "+
				"An initial visitor is given as the first argument. The return value of the function is used as the new visitor "),
	"top": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Top(stack) }).
		SetDescription("n", "Returns the first n items of the list."),
	"skip": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Skip(stack) }).
		SetDescription("n", "Returns a list without the first n items."),
	"number": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Number(stack) }).
		SetDescription("func(n,item) item",
			"Returns a list with the given function applied to each item in the list. "+
				"The function is called with the index of the item and the item itself."),
	"present": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.Present(stack) }).
		SetDescription("func(item) bool", "Returns true if the given function returns true for any item in the list."),
	"size": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return Int(list.Size()) }).
		SetDescription("Returns the number of items in the list."),
	"first": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return list.First() }).
		SetDescription("Returns the first item in the list."),
	"string": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) Value { return String(list.String()) }).
		SetDescription("Returns the list as a string."),
	"movingWindow": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) Value { return list.MovingWindow(stack) }).
		SetDescription("func(item) float", "Returns a list of lists. "+
			"The inner lists contain all items that are close to each other. "+
			"Two items are close to each other if the given function returns a similar value for both items. "+
			"Similarity is defined as the absolute difference being smaller than 1."),
}

func (l *List) GetMethod(name string) (funcGen.Function[Value], error) {
	return ListMethods.Get(name)
}
