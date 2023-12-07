package value

import (
	"bytes"
	"errors"
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
		return func(yield func(Value) bool) (bool, error) {
			for _, item := range items {
				if !yield(conv(item)) {
					return false, nil
				}
			}
			return true, nil
		}
	})
}

// NewList creates a new list containing the given elements
func NewList(items ...Value) *List {
	return &List{items: items, itemsPresent: true, iterable: createSliceIterable(items)}
}

func createSliceIterable(items []Value) iterator.Iterable[Value] {
	return func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) (bool, error) {
			for _, item := range items {
				if !yield(item) {
					return false, nil
				}
			}
			return true, nil
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

func (l *List) Eval() error {
	if !l.itemsPresent {
		var it []Value
		_, err := l.iterable()(func(value Value) bool {
			it = append(it, value)
			return true
		})
		if err != nil {
			return err
		}
		l.items = it
		l.itemsPresent = true
		l.iterable = createSliceIterable(it)
	}
	return nil
}

func (l *List) Equals(other *List) (bool, error) {
	a, aErr := l.ToSlice()
	if aErr != nil {
		return false, aErr
	}
	b, bErr := other.ToSlice()
	if bErr != nil {
		return false, bErr
	}
	if len(a) != len(b) {
		return false, nil
	}
	for i, aa := range a {
		equal, err := Equal(aa, b[i])
		if err != nil {
			return false, err
		}
		if !equal {
			return false, nil
		}
	}
	return true, nil
}

func (l *List) Iterator() iterator.Iterator[Value] {
	return l.iterable()
}

// ToSlice returns the list elements as a slice
func (l *List) ToSlice() ([]Value, error) {
	err := l.Eval()
	if err != nil {
		return nil, err
	}
	return l.items[0:len(l.items):len(l.items)], nil
}

// CopyToSlice creates a slice copy of all elements
func (l *List) CopyToSlice() ([]Value, error) {
	err := l.Eval()
	if err != nil {
		return nil, err
	}
	co := make([]Value, len(l.items))
	copy(co, l.items)
	return co, nil
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

func ToFunc(name string, st funcGen.Stack[Value], n int, args int) (funcGen.Function[Value], error) {
	if c, ok := st.Get(n).ToClosure(); ok {
		if c.Args == args {
			return c, nil
		} else {
			return funcGen.Function[Value]{}, fmt.Errorf("%d. argument of %s needs to be a function with %d arguments", n, name, args)
		}
	} else {
		return funcGen.Function[Value]{}, fmt.Errorf("%d. argument of %s needs to be a function", n, name)
	}
}

func ToFloat(name string, st funcGen.Stack[Value], n int) (float64, error) {
	if c, ok := st.Get(n).ToFloat(); ok {
		return c, nil
	} else {
		return 0, fmt.Errorf("%d. argument of %s needs to be a float", n, name)
	}
}

func (l *List) Accept(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("accept", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.FilterAuto[Value](l.iterable, func() func(v Value) (bool, error) {
		lst := funcGen.NewEmptyStack[Value]()
		return func(v Value) (bool, error) {
			eval, err := f.Eval(lst, v)
			if err != nil {
				return false, err
			}
			if accept, ok := eval.ToBool(); ok {
				return accept, nil
			}
			return false, fmt.Errorf("function in accept does not return a bool")
		}
	})), nil
}

func (l *List) Map(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("map", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.MapAuto[Value, Value](l.iterable, func() func(i int, v Value) (Value, error) {
		lst := funcGen.NewEmptyStack[Value]()
		return func(i int, v Value) (Value, error) {
			return f.Eval(lst, v)
		}
	})), nil
}

func (l *List) Compact(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("compact", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.Compact[Value](l.iterable, func(a, b Value) (bool, error) {
		aVal, err := f.Eval(st, a)
		if err != nil {
			return false, err
		}
		bVal, err := f.Eval(st, b)
		if err != nil {
			return false, err
		}
		return Equal(aVal, bVal)
	})), nil
}

func (l *List) Cross(st funcGen.Stack[Value]) (*List, error) {
	other := st.Get(1)
	f, err := ToFunc("cross", st, 2, 2)
	if err != nil {
		return nil, err
	}
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Cross[Value, Value](l.iterable, otherList.iterable, func(a, b Value) (Value, error) {
			st.Push(a)
			st.Push(b)
			return f.Func(st.CreateFrame(2), nil)
		})), nil
	} else {
		return nil, errors.New("first argument in cross needs to be a list")
	}
}

func (l *List) Merge(st funcGen.Stack[Value]) (*List, error) {
	other := st.Get(1)
	f, err := ToFunc("merge", st, 2, 2)
	if err != nil {
		return nil, err
	}
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Merge[Value](l.iterable, otherList.iterable, func(a, b Value) (bool, error) {
			st.Push(a)
			st.Push(b)
			value, err2 := f.Func(st.CreateFrame(2), nil)
			if err2 != nil {
				return false, err2
			}
			if less, ok := value.ToBool(); ok {
				return less, nil
			} else {
				return false, errors.New("function in merge needs to return a bool, (a<b)")
			}
		})), nil
	} else {
		return nil, errors.New("first argument in merge needs to be a list")
	}
}

func (l *List) First() (Value, error) {
	if l.itemsPresent {
		if len(l.items) > 0 {
			return l.items[0], nil
		}
	} else {
		var first Value
		found := false
		_, err := l.iterable()(func(value Value) bool {
			first = value
			found = true
			return false
		})
		if err != nil {
			return nil, err
		}
		if found {
			return first, nil
		}
	}
	return nil, errors.New("error in first, no items in list")
}

func (l *List) Last() (Value, error) {
	if l.itemsPresent {
		if len(l.items) > 0 {
			return l.items[len(l.items)-1], nil
		}
	} else {
		var last Value
		found := false
		_, err := l.iterable()(func(value Value) bool {
			last = value
			found = true
			return true
		})
		if err != nil {
			return nil, err
		}
		if found {
			return last, nil
		}
	}
	return nil, errors.New("error in first, no items in list")
}

func (l *List) IndexOf(st funcGen.Stack[Value]) (Int, error) {
	v := st.Get(1)
	index := -1
	i := 0
	var innerErr error
	_, err := l.iterable()(func(value Value) bool {
		equal, err := Equal(v, value)
		if err != nil {
			innerErr = err
			return false
		}
		if equal {
			index = i
			return false
		}
		i++
		return true
	})
	if innerErr != nil {
		return 0, innerErr
	}
	if err != nil {
		return 0, err
	}
	return Int(index), nil
}

type SortableLess struct {
	items []Value
	st    funcGen.Stack[Value]
	less  funcGen.Function[Value]
	err   error
}

func (s SortableLess) Len() int {
	return len(s.items)
}

func (s SortableLess) Less(i, j int) bool {
	s.st.Push(s.items[i])
	s.st.Push(s.items[j])
	value, err := s.less.Func(s.st.CreateFrame(2), nil)
	if err != nil {
		if s.err == nil {
			s.err = err
		}
	}
	if l, ok := value.ToBool(); ok {
		return l
	} else {
		if s.err == nil {
			s.err = errors.New("function in orderLess needs to return a bool")
		}
		return false
	}
}

func (s SortableLess) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (l *List) OrderLess(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("orderLess", st, 1, 2)
	if err != nil {
		return nil, err
	}
	items, err := l.CopyToSlice()
	if err != nil {
		return nil, err
	}
	s := SortableLess{items: items, st: st, less: f}
	sort.Sort(s)
	return NewList(items...), s.err
}

type Sortable struct {
	items    []Value
	rev      bool
	st       funcGen.Stack[Value]
	pickFunc funcGen.Function[Value]
	err      error
}

func (s *Sortable) Len() int {
	return len(s.items)
}

func (s *Sortable) registerError(err error) {
	if s.err == nil {
		s.err = err
	}
}

func (s *Sortable) pick(i int) (Value, bool) {
	s.st.Push(s.items[i])
	value, err := s.pickFunc.Func(s.st.CreateFrame(1), nil)
	s.registerError(err)
	return value, err != nil
}

func (s *Sortable) Less(i, j int) bool {
	pi, oki := s.pick(i)
	pj, okj := s.pick(j)
	if oki && okj {
		if s.rev {
			less, err := Less(pj, pi)
			s.registerError(err)
			return less
		} else {
			less, err := Less(pi, pj)
			s.registerError(err)
			return less
		}
	} else {
		return false
	}
}

func (s *Sortable) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (l *List) Order(st funcGen.Stack[Value], rev bool) (*List, error) {
	f, err := ToFunc("order", st, 1, 1)
	if err != nil {
		return nil, err
	}
	items, err := l.CopyToSlice()
	if err != nil {
		return nil, err
	}
	s := Sortable{items: items, rev: rev, st: st, pickFunc: f}
	sort.Sort(&s)
	return NewList(items...), s.err
}

func (l *List) Reverse() (*List, error) {
	items, err := l.CopyToSlice()
	if err != nil {
		return nil, err
	}
	//reverse items
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return NewList(items...), nil
}

func (l *List) Combine(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("combine", st, 1, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.Combine[Value, Value](l.iterable, func(a, b Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	})), nil
}

func (l *List) Combine3(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("combine3", st, 1, 3)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.Combine3[Value, Value](l.iterable, func(a, b, c Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		st.Push(c)
		return f.Func(st.CreateFrame(3), nil)
	})), nil
}

func (l *List) CombineN(st funcGen.Stack[Value]) (*List, error) {
	if n, ok := st.Get(1).ToInt(); ok {
		f, err := ToFunc("combineN", st, 2, 1)
		if err != nil {
			return nil, err
		}
		return NewListFromIterable(iterator.CombineN[Value, Value](l.iterable, n, func(i0 int, i []Value) (Value, error) {
			st.Push(NewList(i...))
			return f.Func(st.CreateFrame(1), nil)
		})), nil
	}
	return nil, errors.New("first argument in combineN needs to be an int")
}

func (l *List) IIr(st funcGen.Stack[Value]) (*List, error) {
	initial, err := ToFunc("iir", st, 1, 1)
	if err != nil {
		return nil, err
	}
	function, err := ToFunc("iir", st, 2, 2)
	return NewListFromIterable(iterator.IirMap[Value, Value](l.iterable,
		func(item Value) (Value, error) {
			return initial.Eval(st, item)
		},
		func(item Value, lastItem Value, last Value) (Value, error) {
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(2), nil)
		})), nil
}

func (l *List) IIrCombine(st funcGen.Stack[Value]) (*List, error) {
	initial, err := ToFunc("iirCombine", st, 1, 1)
	if err != nil {
		return nil, err
	}
	function, err := ToFunc("iirCombine", st, 2, 3)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.IirMap[Value, Value](l.iterable,
		func(item Value) (Value, error) {
			return initial.Eval(st, item)
		},
		func(item Value, lastItem Value, last Value) (Value, error) {
			st.Push(lastItem)
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(3), nil)
		})), nil
}

func (l *List) IIrApply(st funcGen.Stack[Value]) (*List, error) {
	if m, ok := st.Get(1).ToMap(); ok {
		initial, err := funcFromMap(m, "initial", 1)
		if err != nil {
			return nil, err
		}
		function, err := funcFromMap(m, "filter", 3)
		return NewListFromIterable(iterator.IirMap[Value, Value](l.iterable,
			func(item Value) (Value, error) {
				return initial.Eval(st, item)
			},
			func(item Value, lastItem Value, last Value) (Value, error) {
				st.Push(lastItem)
				st.Push(item)
				st.Push(last)
				return function.Func(st.CreateFrame(3), nil)
			})), nil
	} else {
		return nil, errors.New("first argument in iirApply needs to be a map")
	}
}

func funcFromMap(m Map, key string, args int) (funcGen.Function[Value], error) {
	if f, ok := m.Get(key); ok {
		if ff, ok := f.ToClosure(); ok {
			if ff.Args == args {
				return ff, nil
			} else {
				return funcGen.Function[Value]{}, fmt.Errorf("function in %s needs to have %d arguments", key, args)
			}
		} else {
			return funcGen.Function[Value]{}, fmt.Errorf("value in %s needs to be a function", key)
		}
	} else {
		return funcGen.Function[Value]{}, fmt.Errorf("function %s is missing", key)
	}
}

func (l *List) Visit(st funcGen.Stack[Value]) (Value, error) {
	visitor := st.Get(1)
	function, err := ToFunc("visit", st, 2, 2)
	if err != nil {
		return nil, err
	}
	var innerErr error
	_, err = l.iterable()(func(value Value) bool {
		st.Push(visitor)
		st.Push(value)
		visitor, err = function.Func(st.CreateFrame(2), nil)
		if err != nil {
			innerErr = err
			return false
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	return visitor, err
}

func (l *List) Present(st funcGen.Stack[Value]) (Value, error) {
	function, err := ToFunc("present", st, 1, 1)
	if err != nil {
		return nil, err
	}
	isPresent := false
	var innerErr error
	_, err = l.iterable()(func(value Value) bool {
		st.Push(value)
		v, err2 := function.Func(st.CreateFrame(1), nil)
		if err2 != nil {
			innerErr = err2
			return false
		}
		if pr, ok := v.ToBool(); ok {
			if pr {
				isPresent = true
				return false
			}
		} else {
			innerErr = errors.New("function in present needs to return a bool")
			return false
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	return Bool(isPresent), err
}

func (l *List) Top(st funcGen.Stack[Value]) (*List, error) {
	if i, ok := st.Get(1).ToInt(); ok {
		return NewListFromIterable(iterator.FirstN[Value](l.iterable, i)), nil
	}
	return nil, errors.New("error in top, no int given")
}

func (l *List) Skip(st funcGen.Stack[Value]) (*List, error) {
	if i, ok := st.Get(1).ToInt(); ok {
		return NewListFromIterable(iterator.Skip[Value](l.iterable, i)), nil
	}
	return nil, errors.New("error in skip, no int given")
}

func (l *List) Reduce(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("reduce", st, 1, 2)
	if err != nil {
		return nil, err
	}
	return iterator.Reduce[Value](l.iterable, func(a, b Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	})
}

func (l *List) Sum() (Value, error) {
	var sum Value
	var innerErr error
	_, err := l.Iterator()(func(value Value) bool {
		if sum == nil {
			sum = value
		} else {
			var err error
			sum, err = Add(sum, value)
			if err != nil {
				innerErr = err
				return false
			}
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}
	return sum, nil
}

func (l *List) MapReduce(st funcGen.Stack[Value]) (Value, error) {
	initial := st.Get(1)
	f, err := ToFunc("mapReduce", st, 2, 2)
	if err != nil {
		return nil, err
	}
	return iterator.MapReduce(l.iterable, initial, func(s Value, v Value) (Value, error) {
		st.Push(s)
		st.Push(v)
		return f.Func(st.CreateFrame(2), nil)
	})
}

func (l *List) MinMax(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("minMax", st, 1, 1)
	if err != nil {
		return nil, err
	}
	first := true
	var minVal Value = Int(0)
	var maxVal Value = Int(0)
	var innerErr error
	_, err = l.Iterator()(func(value Value) bool {
		st.Push(value)
		r, err := f.Func(st.CreateFrame(1), nil)
		if err != nil {
			innerErr = err
			return false
		}
		if first {
			first = false
			minVal = r
			maxVal = r
		} else {
			less, err := Less(r, minVal)
			if err != nil {
				innerErr = err
				return false
			}
			if less {
				minVal = r
			}
			b, err := Less(maxVal, r)
			if err != nil {
				innerErr = err
				return false
			}
			if b {
				maxVal = r
			}
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}
	return NewMap(listMap.New[Value](3).
		Append("min", minVal).
		Append("max", maxVal).
		Append("valid", Bool(!first))), nil
}

func (l *List) Replace(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("replace", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return f.Eval(st, l)
}

func (l *List) Number(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("number", st, 1, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(func() iterator.Iterator[Value] {
		return func(yield func(Value) bool) (bool, error) {
			n := Int(0)
			var innerErr error
			_, err := l.iterable()(func(value Value) bool {
				st.Push(n)
				st.Push(value)
				n++
				v, err2 := f.Func(st.CreateFrame(2), nil)
				if err2 != nil {
					innerErr = err2
					return false
				}
				return yield(v)
			})
			if innerErr != nil {
				return false, innerErr
			}
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}), nil
}

func (l *List) GroupByString(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("groupByString", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return groupBy(l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		return String(key.String()), nil
	})
}

func (l *List) GroupByInt(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("groupByInt", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return groupBy(l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		if i, ok := key.ToInt(); ok {
			return Int(i), nil
		} else {
			return nil, errors.New("groupByInt requires an int as key")
		}
	})
}

func groupBy(list *List, keyFunc func(Value) (Value, error)) (*List, error) {
	m := make(map[Value]*[]Value)
	var innerErr error
	_, err := list.iterable()(func(value Value) bool {
		key, err := keyFunc(value)
		if err != nil {
			innerErr = err
			return false
		}
		if l, ok := m[key]; ok {
			*l = append(*l, value)
		} else {
			ll := []Value{value}
			m[key] = &ll
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}
	result := make([]Value, 0, len(m))
	for k, v := range m {
		result = append(result, Map{listMap.New[Value](2).
			Append("key", k).
			Append("values", NewList(*v...))})
	}
	return NewList(result...), nil
}

func (l *List) UniqueString(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("uniqueString", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return unique(l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		return String(key.String()), nil
	})
}

func (l *List) UniqueInt(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("uniqueInt", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return unique(l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		if i, ok := key.ToInt(); ok {
			return Int(i), nil
		} else {
			return nil, errors.New("uniqueInt requires an int as key")
		}
	})
}

func unique(list *List, keyFunc func(Value) (Value, error)) (*List, error) {
	m := make(map[Value]struct{})
	var innerErr error
	_, err := list.iterable()(func(value Value) bool {
		key, err := keyFunc(value)
		if err != nil {
			innerErr = err
			return false
		}
		m[key] = struct{}{}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}
	result := make([]Value, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return NewList(result...), nil
}

func (l *List) MovingWindow(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("movingWindow", st, 1, 1)
	if err != nil {
		return nil, err
	}
	items, err := l.ToSlice()
	if err != nil {
		return nil, err
	}
	values := make([]float64, len(items))
	for i, elem := range items {
		eval, err := f.Eval(st, elem)
		if err != nil {
			return nil, err
		}
		if float, ok := eval.ToFloat(); ok {
			values[i] = float
		} else {
			return nil, errors.New("function in movingWindow needs to return a float")
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
	return NewList(mainList...), nil
}

func (l *List) containsItem(item Value) (bool, error) {
	found := false
	var innerErr error
	_, err := l.iterable()(func(value Value) bool {
		equal, err := Equal(item, value)
		if err != nil {
			innerErr = err
			return false
		}
		if equal {
			found = true
			return false
		}
		return true
	})
	if innerErr != nil {
		return false, innerErr
	}
	if err != nil {
		return false, err

	}
	return found, nil
}

func (l *List) containsAllItems(lookForList *List) (bool, error) {
	lookFor, err := lookForList.CopyToSlice()
	if err != nil {
		return false, err
	}

	if l.itemsPresent && len(l.items) < len(lookFor) {
		return false, nil
	}

	var innerErr error
	_, err = l.iterable()(func(value Value) bool {
		for i, lf := range lookFor {
			equal, err2 := Equal(lf, value)
			if err2 != nil {
				innerErr = err2
				return false
			}
			if equal {
				lookFor = append(lookFor[0:i], lookFor[i+1:]...)
				break
			}
		}
		return len(lookFor) > 0
	})
	if innerErr != nil {
		return false, innerErr
	}
	if err != nil {
		return false, err
	}
	return len(lookFor) == 0, nil
}

var ListMethods = MethodMap{
	"accept": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Accept(stack) }).
		SetMethodDescription("func(item) bool",
			"Filters the list by the given function. If the function returns true, the item is accepted, otherwise it is skipped."),
	"map": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Map(stack) }).
		SetMethodDescription("func(item) newItem",
			"Maps the list by the given function. The function is called for each item in the list and the result is "+
				"added to the new list."),
	"reduce": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Reduce(stack) }).
		SetMethodDescription("func(item, item) item",
			"Reduces the list by the given function. The function is called with the first two list items, and the result "+
				"is used as the first argument for the third item and so on."),
	"sum": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Sum() }).
		SetMethodDescription("Returns the sum of all items in the list."),
	"mapReduce": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MapReduce(stack) }).
		SetMethodDescription("initialSum", "func(sum, item) sum",
			"MapReduce reduces the list to a single value. The initial value is given as the first argument. The function "+
				"is called with the initial value and the first item, and the result is used as the first argument for the "+
				"second item and so on."),
	"minMax": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MinMax(stack) }).
		SetMethodDescription("func(item) value",
			"Returns the minimum and maximum value of the list. The function is called for each item in the list and the "+
				"result is compared to the previous minimum and maximum."),
	"replace": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Replace(stack) }).
		SetMethodDescription("func(list) newItem",
			"Replaces the list by the result of the given function. The function is called with the list as argument."),
	"combine": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Combine(stack) }).
		SetMethodDescription("func(item, item) newItem",
			"Combines the list by the given function. The function is called for each pair of items in the list and the "+
				"result is added to the new list. "+
				"The resulting list is one item shorter than the original list."),
	"combine3": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Combine3(stack) }).
		SetMethodDescription("func(item, item, item) newItem",
			"Combines the list by the given function. The function is called for each triplet of items in the list and "+
				"the result is added to the new list. "+
				"The resulting list is two items shorter than the original list."),
	"combineN": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.CombineN(stack) }).
		SetMethodDescription("n", "func([item...]) newItem",
			"Combines the list by the given function. The function is called for each group of n items in the list and "+
				"the result is added to the new list. "+
				"The resulting list is n-1 items shorter than the original list."),
	"multiUse": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MultiUse(stack) }).
		SetMethodDescription("{name: func(item) newItem...}",
			"MultiUse allows to use the list multiple times without storing or recomputing its elements. The first argument "+
				"is a map of functions. "+
				"All the functions are called with the list as argument and the result is returned in a map. "+
				"The keys in the result map are the same keys used to pass the functions."),
	"indexOf": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.IndexOf(stack) }).
		SetMethodDescription("item",
			"Returns the index of the first occurrence of the given item in the list. If the item is not found, -1 is returned."),
	"groupByString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.GroupByString(stack) }).
		SetMethodDescription("func(item) string", "Returns a list of lists grouped by the given function. "+
			"The function is called for each item in the list and the returned string is used as the key for the group. "+
			"The result is a list of maps with the keys 'key' and 'values'. The 'key' contains the string returned by the function "+
			"and 'values' contains a list of items that have the same key."),
	"groupByInt": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.GroupByInt(stack) }).
		SetMethodDescription("func(item) int", "Returns a list of lists grouped by the given function. "+
			"The function is called for each item in the list and the returned integer is used as the key for the group. "+
			"The result is a list of maps with the keys 'key' and 'values'. The 'key' contains the integer returned by the function "+
			"and 'values' contains a list of items that have the same key."),
	"uniqueString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.UniqueString(stack) }).
		SetMethodDescription("func(item) string", "Returns a list of unique strings returned by the given function."),
	"uniqueInt": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.UniqueInt(stack) }).
		SetMethodDescription("func(item) int", "Returns a list of unique integers returned by the given function."),
	"compact": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Compact(stack) }).
		SetMethodDescription("func(a,b) bool", "Returns a new list with the items compacted. "+
			"The function is called for each pair of items in the list and needs to return true if a=b holds."+
			"Compacting means that an item is removed if the function returns true for the item and the previous item."),
	"cross": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Cross(stack) }).
		SetMethodDescription("other_list", "func(a,b) newItem",
			"Returns a new list with the given function applied to each pair of items in the list and the given list. "+
				"The function is called with an item from the first list and an item from the second list. "+
				"The length of the resulting list is the product of the lengths of the two lists."),
	"merge": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Merge(stack) }).
		SetMethodDescription("other_list", "func(a,b) bool",
			"Returns a new list with the items of both lists combined. "+
				"The given function is called for the pair of the first, non processed items in both lists. If the "+
				"return value is true the value of the original list is taken, otherwise the item from the other list. "+
				"The is repeated until all items of both lists are processed. "+
				"If the function returns true if a<b holds and both lists are ordered, also the new list is ordered."),
	"order": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Order(stack, false) }).
		SetMethodDescription("func(item) value",
			"Returns a new list with the items sorted in the order of the values returned by the given function. "+
				"The function is called for each item in the list and the returned values determine the order."),
	"orderRev": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Order(stack, true) }).
		SetMethodDescription("func(item) value",
			"Returns a new list with the items sorted in the reverse order of the values returned by the given function. "+
				"The function is called for each item in the list and the returned values determine the order."),
	"orderLess": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.OrderLess(stack) }).
		SetMethodDescription("func(a, a) bool",
			"Returns a new list with the items sorted by the given function. "+
				"The function is called for pairs of items in the list and the returned bool needs to be true if a<b holds."),
	"reverse": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Reverse() }).
		SetMethodDescription("Returns the list in reverse order."),
	"append": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Append(stack), nil }).
		SetMethodDescription("item", "Returns a new list with the given item appended."),
	"iir": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.IIr(stack) }).
		SetMethodDescription("func(first_item) first_new_item", "func(item, last_new_item) new_item",
			"Returns a new list with the given functions applied to the items in the list. "+
				"The first function is called with the first item in the list and returns the first item in the new list. "+
				"The second function is called with the remaining items in the list as the first argument, and the last new item. "+
				"For each subsequent item, the function is called with the item and the result of the previous call."),
	"iirCombine": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.IIrCombine(stack) }).
		SetMethodDescription("func(first_item) first_new_item", "func(i0, i1, last_new_item) new_item",
			"Returns a new list with the given functions applied to the items in the list. "+
				"The first function is called with the first item in the list and returns the first item in the new list. "+
				"The second function is called with the remaining pairs of items in the list as the first two arguments, and the last new item. "+
				"For each subsequent item, the function is called with the the pair of items and the result of the previous call. "+
				"The item i0 is the item in front of i1."),
	"iirApply": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.IIrApply(stack) }).
		SetMethodDescription("map",
			"Returns a new list with the given filter applied to the items in the list. "+
				"Works the same as 'iirCombine' except the required functions are taken from the map, stored in the keys 'initial' and 'filter'."),
	"visit": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Visit(stack) }).
		SetMethodDescription("initial_visitor", "func(visitor, item) visitor",
			"Visits each item in the list with the given function. The function is called with the visitor and the item. "+
				"An initial visitor is given as the first argument. The return value of the function is used as the new visitor "),
	"top": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Top(stack) }).
		SetMethodDescription("n", "Returns the first n items of the list."),
	"skip": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Skip(stack) }).
		SetMethodDescription("n", "Returns a list without the first n items."),
	"number": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Number(stack) }).
		SetMethodDescription("func(n,item) item",
			"Returns a list with the given function applied to each item in the list. "+
				"The function is called with the index of the item and the item itself."),
	"present": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Present(stack) }).
		SetMethodDescription("func(item) bool", "Returns true if the given function returns true for any item in the list."),
	"size": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return Int(list.Size()), nil }).
		SetMethodDescription("Returns the number of items in the list."),
	"first": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.First() }).
		SetMethodDescription("Returns the first item in the list."),
	"last": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Last() }).
		SetMethodDescription("Returns the last item in the list."),
	"eval": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { list.Eval(); return list, nil }).
		SetMethodDescription("Evaluates the list and stores all items in memory."),
	"string": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return String(list.String()), nil }).
		SetMethodDescription("Returns the list as a string."),
	"movingWindow": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MovingWindow(stack) }).
		SetMethodDescription("func(item) float", "Returns a list of lists. "+
			"The inner lists contain all items that are close to each other. "+
			"Two items are close to each other if the given function returns a similar value for both items. "+
			"Similarity is defined as the absolute difference being smaller than 1."),
}

func (l *List) GetMethod(name string) (funcGen.Function[Value], error) {
	return ListMethods.Get(name)
}
