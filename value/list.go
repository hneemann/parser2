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
func NewListConvert[I any](conv func(I) Value, items []I) *List {
	return NewListFromSizedIterable(func(s funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		for _, item := range items {
			if e := yield(conv(item)); e != nil {
				return e
			}
		}
		return nil
	}, len(items))
}

// NewListOfMaps creates a list containing the given elements converted to a map using
// the given ToMapInterface.
func NewListOfMaps[I any](toMap ToMapInterface[I], items []I) *List {
	return NewListFromSizedIterable(func(s funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		for _, item := range items {
			if e := yield(toMap.Create(item)); e != nil {
				return e
			}
		}
		return nil
	}, len(items))
}

// NewList creates a new list containing the given elements
func NewList(items ...Value) *List {
	return &List{items: items, itemsPresent: true, iterable: createSliceIterable(items), size: len(items)}
}

func createSliceIterable(items []Value) iterator.Producer[Value, funcGen.Stack[Value]] {
	return func(s funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		for _, item := range items {
			if e := yield(item); e != nil {
				return e
			}
		}
		return nil
	}
}

// NewListFromIterable creates a list based on the given Iterable
func NewListFromIterable(li iterator.Producer[Value, funcGen.Stack[Value]]) *List {
	return &List{iterable: li, itemsPresent: false, size: -1}
}

// NewListFromSizedIterable creates a list based on the given Iterable.
// In contrast to NewListFromIterable, this function is to be used if the
// size of the iterable is known.
func NewListFromSizedIterable(li iterator.Producer[Value, funcGen.Stack[Value]], size int) *List {
	return &List{iterable: li, itemsPresent: false, size: size}
}

// List represents a list of values
type List struct {
	items        []Value
	itemsPresent bool
	iterable     iterator.Producer[Value, funcGen.Stack[Value]]
	size         int
}

func (l *List) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (l *List) ToInt() (int, bool) {
	return 0, false
}

func (l *List) ToFloat() (float64, bool) {
	return 0, false
}

func (l *List) ToString(st funcGen.Stack[Value]) (string, error) {
	var b bytes.Buffer
	b.WriteString("[")
	first := true
	err := l.iterable(st, func(v Value) error {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		s, err := v.ToString(st)
		if err != nil {
			return err
		}
		b.WriteString(s)
		return nil
	})
	if err != nil {
		return "", err
	}
	b.WriteString("]")
	return b.String(), nil
}

func (l *List) String() string {
	var b bytes.Buffer
	b.WriteString("[")
	first := true
	count := 10
	st := funcGen.NewEmptyStack[Value]()
	err := l.iterable(st,
		func(v Value) error {
			if count == 0 {
				return iterator.SBC
			}

			if first {
				first = false
			} else {
				b.WriteString(", ")
			}
			s, err := v.ToString(st)
			if err != nil {
				return err
			}
			b.WriteString(s)
			count--
			return nil
		})
	if err != nil {
		if err == iterator.SBC {
			b.WriteString(", ...]")
		} else {
			return fmt.Sprintf("error in list: %v", err)
		}
	} else {
		b.WriteString("]")
	}
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

func (l *List) Eval(st funcGen.Stack[Value]) error {
	if !l.itemsPresent {
		var it []Value
		err := l.iterable(st, func(value Value) error {
			it = append(it, value)
			return nil
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

func deepEvalLists(st funcGen.Stack[Value], v Value) error {
	switch v := v.(type) {
	case *List:
		sl, err := v.ToSlice(st)
		if err != nil {
			return err
		}
		for _, vv := range sl {
			err := deepEvalLists(st, vv)
			if err != nil {
				return err
			}
		}
		return nil
	case Map:
		var innerErr error
		v.Iter(func(key string, value Value) bool {
			err := deepEvalLists(st, value)
			if err != nil {
				innerErr = err
				return false
			}
			return true
		})
		return innerErr
	}
	return nil
}

func (l *List) Equals(st funcGen.Stack[Value], other *List, equal funcGen.BoolFunc[Value]) (bool, error) {
	a, aErr := l.ToSlice(st)
	if aErr != nil {
		return false, aErr
	}
	b, bErr := other.ToSlice(st)
	if bErr != nil {
		return false, bErr
	}
	if len(a) != len(b) {
		return false, nil
	}
	for i, aa := range a {
		eq, err := equal(st, aa, b[i])
		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}

func (l *List) Iterate(st funcGen.Stack[Value], consumer iterator.Consumer[Value]) error {
	return l.iterable(st, consumer)
}

// ToSlice returns the list elements as a slice
func (l *List) ToSlice(st funcGen.Stack[Value]) ([]Value, error) {
	err := l.Eval(st)
	if err != nil {
		return nil, err
	}
	return l.items[0:len(l.items):len(l.items)], nil
}

// CopyToSlice creates a slice copy of all elements
func (l *List) CopyToSlice(st funcGen.Stack[Value]) ([]Value, error) {
	err := l.Eval(st)
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
func (l *List) Append(st funcGen.Stack[Value]) (*List, error) {
	err := l.Eval(st)
	if err != nil {
		return nil, err
	}
	newList := append(l.items, st.Get(1))
	// Guarantee a copy operation the next time append is called on this
	// list, which is only a rare special case, as the new list is usually
	// appended to.
	if len(l.items) != cap(l.items) {
		l.items = l.items[:len(l.items):len(l.items)]
	}
	return NewList(newList...), nil
}

func (l *List) SizeIfKnown() (int, bool) {
	if l.itemsPresent {
		return len(l.items), true
	} else if l.size >= 0 {
		return l.size, true
	} else {
		return 0, false
	}
}

func (l *List) Size(st funcGen.Stack[Value]) (int, error) {
	err := l.Eval(st)
	if err != nil {
		return 0, err
	}
	return len(l.items), nil
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

func (l *List) Accept(sta funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("accept", sta, 1, 1)
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

func (l *List) Map(sta funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("map", sta, 1, 1)
	if err != nil {
		return nil, err
	}

	return NewListFromSizedIterable(iterator.MapAuto[Value, Value](l.iterable, func() func(i int, v Value) (Value, error) {
		lst := funcGen.NewEmptyStack[Value]()
		return func(i int, v Value) (Value, error) {
			return f.Eval(lst, v)
		}
	}), l.size), nil

}

func (l *List) Compact(fg *FunctionGenerator) (*List, error) {
	return NewListFromIterable(func(st funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		var last Value
		fmt.Println("start")
		err := l.iterable(st, func(v Value) error {
			if last == nil {
				last = v
				return yield(v)
			}

			eq, err := fg.equal(st, last, v)
			if err != nil {
				return err
			}
			last = v
			if !eq {
				return yield(v)
			} else {
				return nil
			}
		})
		return err
	}), nil
}

func (l *List) Cross(sta funcGen.Stack[Value]) (*List, error) {
	other := sta.Get(1)
	f, err := ToFunc("cross", sta, 2, 2)
	if err != nil {
		return nil, err
	}
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Cross[Value, Value](l.iterable, otherList.iterable, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
			st.Push(a)
			st.Push(b)
			return f.Func(st.CreateFrame(2), nil)
		})), nil
	} else {
		return nil, errors.New("first argument in cross needs to be a list")
	}
}

func (l *List) Merge(sta funcGen.Stack[Value]) (*List, error) {
	other := sta.Get(1)
	f, err := ToFunc("merge", sta, 2, 2)
	if err != nil {
		return nil, err
	}
	if otherList, ok := other.ToList(); ok {
		return NewListFromIterable(iterator.Merge[Value](l.iterable, otherList.iterable, func(st funcGen.Stack[Value], a, b Value) (bool, error) {
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
		}, func() funcGen.Stack[Value] {
			return funcGen.NewEmptyStack[Value]()
		})), nil
	} else {
		return nil, errors.New("first argument in merge needs to be a list")
	}
}

func (l *List) First(st funcGen.Stack[Value]) (Value, error) {
	if l.itemsPresent {
		if len(l.items) > 0 {
			return l.items[0], nil
		}
	} else {
		var first Value
		found := false
		err := l.iterable(st, func(value Value) error {
			first = value
			found = true
			return iterator.SBC
		})
		if err != nil && err != iterator.SBC {
			return nil, err
		}
		if found {
			return first, nil
		}
	}
	return nil, errors.New("error in first, no items in list")
}

func (l *List) Single(st funcGen.Stack[Value]) (Value, error) {
	if l.itemsPresent {
		if len(l.items) == 1 {
			return l.items[0], nil
		}
	} else {
		var first Value
		found := false
		err := l.iterable(st, func(value Value) error {
			if found {
				return errors.New("error in single, more than one item in list")
			}
			first = value
			found = true
			return nil
		})
		if err != nil {
			return nil, err
		}
		if found {
			return first, nil
		}
	}
	return nil, errors.New("error in single not a single item in list")
}

func (l *List) Last(st funcGen.Stack[Value]) (Value, error) {
	if l.itemsPresent {
		if len(l.items) > 0 {
			return l.items[len(l.items)-1], nil
		}
	} else {
		var last Value
		found := false
		err := l.iterable(st, func(value Value) error {
			last = value
			found = true
			return nil
		})
		if err != nil {
			return nil, err
		}
		if found {
			return last, nil
		}
	}
	return nil, errors.New("error in last, no items in list")
}

func (l *List) IndexWhere(st funcGen.Stack[Value]) (Int, error) {
	f, err := ToFunc("indexOf", st, 1, 1)
	if err != nil {
		return 0, err
	}
	index := -1
	i := 0
	err = l.iterable(st, func(value Value) error {
		st.Push(value)
		found, err := f.Func(st.CreateFrame(1), nil)
		if err != nil {
			return err
		}
		if f, ok := found.ToBool(); ok {
			if f {
				index = i
				return iterator.SBC
			}
		} else {
			return errors.New("function in indexOf needs to return a bool")
		}
		i++
		return nil
	})
	if err != nil && err != iterator.SBC {
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
	items, err := l.CopyToSlice(st)
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
	less     funcGen.BoolFunc[Value]
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
	return value, err == nil
}

func (s *Sortable) Less(i, j int) bool {
	pi, oki := s.pick(i)
	pj, okj := s.pick(j)
	if oki && okj {
		if s.rev {
			less, err := s.less(s.st, pj, pi)
			s.registerError(err)
			return less
		} else {
			less, err := s.less(s.st, pi, pj)
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

func (l *List) Order(st funcGen.Stack[Value], rev bool, fg *FunctionGenerator) (*List, error) {
	f, err := ToFunc("order", st, 1, 1)
	if err != nil {
		return nil, err
	}
	items, err := l.CopyToSlice(st)
	if err != nil {
		return nil, err
	}
	s := Sortable{items: items, rev: rev, st: st, pickFunc: f, less: fg.less}
	sort.Sort(&s)
	return NewList(items...), s.err
}

func (l *List) Reverse(st funcGen.Stack[Value]) (*List, error) {
	items, err := l.CopyToSlice(st)
	if err != nil {
		return nil, err
	}
	//reverse items
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return NewList(items...), nil
}

func (l *List) Combine(sta funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("combine", sta, 1, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.Combine[Value, Value](l.iterable, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	})), nil
}

func (l *List) Combine3(sta funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("combine3", sta, 1, 3)
	if err != nil {
		return nil, err
	}
	return NewListFromIterable(iterator.Combine3[Value, Value](l.iterable, func(st funcGen.Stack[Value], a, b, c Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		st.Push(c)
		return f.Func(st.CreateFrame(3), nil)
	})), nil
}

func (l *List) CombineN(sta funcGen.Stack[Value]) (*List, error) {
	if n, ok := sta.Get(1).ToInt(); ok {
		f, err := ToFunc("combineN", sta, 2, 1)
		if err != nil {
			return nil, err
		}
		return NewListFromIterable(iterator.CombineN[Value, Value](l.iterable, n, func(st funcGen.Stack[Value], i0 int, i []Value) (Value, error) {
			st.Push(NewList(i...))
			return f.Func(st.CreateFrame(1), nil)
		})), nil
	}
	return nil, errors.New("first argument in combineN needs to be an int")
}

func (l *List) IIr(sta funcGen.Stack[Value]) (*List, error) {
	initial, err := ToFunc("iir", sta, 1, 1)
	if err != nil {
		return nil, err
	}
	function, err := ToFunc("iir", sta, 2, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromSizedIterable(iterator.IirMap[Value, Value](l.iterable,
		func(st funcGen.Stack[Value], item Value) (Value, error) {
			return initial.Eval(st, item)
		},
		func(st funcGen.Stack[Value], item Value, lastItem Value, last Value) (Value, error) {
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(2), nil)
		}), l.size), nil
}

func (l *List) IIrCombine(sta funcGen.Stack[Value]) (*List, error) {
	initial, err := ToFunc("iirCombine", sta, 1, 1)
	if err != nil {
		return nil, err
	}
	function, err := ToFunc("iirCombine", sta, 2, 3)
	if err != nil {
		return nil, err
	}
	return NewListFromSizedIterable(iterator.IirMap[Value, Value](l.iterable,
		func(st funcGen.Stack[Value], item Value) (Value, error) {
			return initial.Eval(st, item)
		},
		func(st funcGen.Stack[Value], item Value, lastItem Value, last Value) (Value, error) {
			st.Push(lastItem)
			st.Push(item)
			st.Push(last)
			return function.Func(st.CreateFrame(3), nil)
		}), l.size), nil
}

func (l *List) IIrApply(sta funcGen.Stack[Value]) (*List, error) {
	if m, ok := sta.Get(1).ToMap(); ok {
		initial, err := funcFromMap(m, "initial", 1)
		if err != nil {
			return nil, err
		}
		function, err := funcFromMap(m, "filter", 3)
		return NewListFromSizedIterable(iterator.IirMap[Value, Value](l.iterable,
			func(st funcGen.Stack[Value], item Value) (Value, error) {
				return initial.Eval(st, item)
			},
			func(st funcGen.Stack[Value], item Value, lastItem Value, last Value) (Value, error) {
				st.Push(lastItem)
				st.Push(item)
				st.Push(last)
				return function.Func(st.CreateFrame(3), nil)
			}), l.size), nil
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
	err = l.iterable(st, func(value Value) error {
		st.Push(visitor)
		st.Push(value)
		visitor, err = function.Func(st.CreateFrame(2), nil)
		if err != nil {
			return err
		}
		return nil
	})
	return visitor, err
}

func createState(s int) Value {
	return NewMap(listMap.New[Value](1).Append("state", Int(s)))
}

func (l *List) FSM(sta funcGen.Stack[Value]) (Value, error) {
	function, err := ToFunc("fsm", sta, 1, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromSizedIterable(iterator.IirMap[Value, Value](l.iterable,
		func(st funcGen.Stack[Value], item Value) (Value, error) {
			st.Push(createState(0))
			st.Push(item)
			return function.Func(st.CreateFrame(2), nil)
		},
		func(st funcGen.Stack[Value], item Value, lastItem Value, last Value) (Value, error) {
			st.Push(last)
			st.Push(item)
			return function.Func(st.CreateFrame(2), nil)
		}), l.size), nil
}

func (l *List) Present(st funcGen.Stack[Value]) (Value, error) {
	function, err := ToFunc("present", st, 1, 1)
	if err != nil {
		return nil, err
	}
	isPresent := false
	err = l.iterable(st, func(value Value) error {
		st.Push(value)
		v, err2 := function.Func(st.CreateFrame(1), nil)
		if err2 != nil {
			return err2
		}
		if pr, ok := v.ToBool(); ok {
			if pr {
				isPresent = true
				return iterator.SBC
			}
		} else {
			return errors.New("function in present needs to return a bool")
		}
		return nil
	})
	if err != nil && err != iterator.SBC {
		return nil, err
	}
	return Bool(isPresent), nil
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
	return iterator.Reduce[Value](st, l.iterable, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	})
}

func (l *List) Sum(st funcGen.Stack[Value], fg *FunctionGenerator) (Value, error) {
	var sum Value
	add := fg.GetOpImpl("+")
	err := l.iterable(st, func(value Value) error {
		if sum == nil {
			sum = value
		} else {
			var err error
			sum, err = add.Calc(st, sum, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if sum == nil {
		return nil, errors.New("sum on empty list")
	}
	return sum, nil
}

func (l *List) MapReduce(st funcGen.Stack[Value]) (Value, error) {
	initial := st.Get(1)
	f, err := ToFunc("mapReduce", st, 2, 2)
	if err != nil {
		return nil, err
	}
	return iterator.MapReduce(st, l.iterable, initial, func(st funcGen.Stack[Value], s Value, v Value) (Value, error) {
		st.Push(s)
		st.Push(v)
		return f.Func(st.CreateFrame(2), nil)
	})
}

func (l *List) MinMax(st funcGen.Stack[Value], fg *FunctionGenerator) (Value, error) {
	f, err := ToFunc("minMax", st, 1, 1)
	if err != nil {
		return nil, err
	}
	first := true
	var minVal Value = Int(0)
	var minItem Value = Int(0)
	var maxVal Value = Int(0)
	var maxItem Value = Int(0)
	err = l.iterable(st, func(value Value) error {
		st.Push(value)
		r, err := f.Func(st.CreateFrame(1), nil)
		if err != nil {
			return err
		}
		if first {
			first = false
			minVal = r
			maxVal = r
			minItem = value
			maxItem = value
		} else {
			le, err := fg.less(st, r, minVal)
			if err != nil {
				return err
			}
			if le {
				minVal = r
				minItem = value
			}
			gr, err := fg.less(st, maxVal, r)
			if err != nil {
				return err
			}
			if gr {
				maxVal = r
				maxItem = value
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewMap(listMap.New[Value](3).
		Append("min", minVal).
		Append("max", maxVal).
		Append("minItem", minItem).
		Append("maxItem", maxItem).
		Append("valid", Bool(!first))), nil
}

func (l *List) Min(st funcGen.Stack[Value], fg *FunctionGenerator) (Value, error) {
	first := true
	var minVal Value = Int(0)
	err := l.iterable(st, func(value Value) error {
		if first {
			first = false
			minVal = value
		} else {
			le, err := fg.less(st, value, minVal)
			if err != nil {
				return err
			}
			if le {
				minVal = value
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if first {
		return nil, errors.New("min of empty list")
	}
	return minVal, nil
}

func (l *List) Max(st funcGen.Stack[Value], fg *FunctionGenerator) (Value, error) {
	first := true
	var maxVal Value = Int(0)
	err := l.iterable(st, func(value Value) error {
		if first {
			first = false
			maxVal = value
		} else {
			le, err := fg.less(st, maxVal, value)
			if err != nil {
				return err
			}
			if le {
				maxVal = value
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if first {
		return nil, errors.New("max of empty list")
	}
	return maxVal, nil
}

func (l *List) Mean(st funcGen.Stack[Value], fg *FunctionGenerator) (Value, error) {
	add := fg.GetOpImpl("+")
	div := fg.GetOpImpl("/")
	var sum Value
	n := 0
	err := l.iterable(st, func(value Value) error {
		if sum == nil {
			sum = value
			n = 1
		} else {
			var err error
			sum, err = add.Calc(st, sum, value)
			if err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if n > 0 {
		return div.Calc(st, sum, Int(n))
	} else {
		return nil, errors.New("mean of empty list")
	}
}

func (l *List) ReplaceList(st funcGen.Stack[Value]) (Value, error) {
	f, err := ToFunc("replaceList", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return f.Eval(st, l)
}

func (l *List) Number(sta funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("number", sta, 1, 2)
	if err != nil {
		return nil, err
	}
	return NewListFromSizedIterable(func(st funcGen.Stack[Value], yield iterator.Consumer[Value]) error {
		n := Int(0)
		return l.iterable(st, func(value Value) error {
			st.Push(n)
			st.Push(value)
			n++
			v, err2 := f.Func(st.CreateFrame(2), nil)
			if err2 != nil {
				return err2
			}
			return yield(v)
		})
	}, l.size), nil
}

func (l *List) GroupByEqual(st funcGen.Stack[Value], fg *FunctionGenerator) (*List, error) {
	keyFunc, err := ToFunc("groupByEqual", st, 1, 1)
	if err != nil {
		return nil, err
	}

	type item struct {
		key    Value
		values []Value
	}

	var items []item

	err = l.iterable(st, func(value Value) error {
		key, err := keyFunc.Eval(st, value)
		if err != nil {
			return err
		}

		for ind, item := range items {
			if eq, err := fg.equal(st, item.key, key); err != nil {
				return err
			} else {
				if eq {
					items[ind].values = append(items[ind].values, value)
					return nil
				}
			}
		}
		items = append(items, item{key: key, values: []Value{value}})
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := make([]Value, 0, len(items))
	for _, v := range items {
		result = append(result, Map{listMap.New[Value](2).
			Append("key", v.key).
			Append("values", NewList(v.values...))})
	}
	return NewList(result...), nil
}

func (l *List) GroupByString(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("groupByString", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return groupBy(st, l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		s, err := key.ToString(st)
		return String(s), err
	})
}

func (l *List) GroupByInt(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("groupByInt", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return groupBy(st, l, func(value Value) (Value, error) {
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

func groupBy(st funcGen.Stack[Value], list *List, keyFunc func(Value) (Value, error)) (*List, error) {
	m := make(map[Value]*[]Value)
	err := list.iterable(st, func(value Value) error {
		key, err := keyFunc(value)
		if err != nil {
			return err
		}
		if l, ok := m[key]; ok {
			*l = append(*l, value)
		} else {
			ll := []Value{value}
			m[key] = &ll
		}
		return nil
	})
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
	return unique(st, l, func(value Value) (Value, error) {
		st.Push(value)
		key, err := keyFunc.Func(st.CreateFrame(1), nil)
		if err != nil {
			return nil, err
		}
		s, err := key.ToString(st)
		return String(s), err
	})
}

func (l *List) UniqueInt(st funcGen.Stack[Value]) (*List, error) {
	keyFunc, err := ToFunc("uniqueInt", st, 1, 1)
	if err != nil {
		return nil, err
	}
	return unique(st, l, func(value Value) (Value, error) {
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

func unique(st funcGen.Stack[Value], list *List, keyFunc func(Value) (Value, error)) (*List, error) {
	m := make(map[Value]struct{})
	err := list.iterable(st, func(value Value) error {
		key, err := keyFunc(value)
		if err != nil {
			return err
		}
		m[key] = struct{}{}
		return nil
	})
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
	items, err := l.ToSlice(st)
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

func (l *List) MovingWindowRemove(st funcGen.Stack[Value]) (*List, error) {
	f, err := ToFunc("movingWindowRemove", st, 1, 1)
	if err != nil {
		return nil, err
	}
	items, err := l.ToSlice(st)
	if err != nil {
		return nil, err
	}

	var mainList []Value
	startIndex := 0
	for i := range items {
		for {
			li := NewList(items[startIndex : i+1 : i+1]...)
			if startIndex == i {
				mainList = append(mainList, li)
				break
			} else {
				removeValue, err := f.Eval(st, li)
				if err != nil {
					return nil, err
				}
				remove, ok := removeValue.ToBool()
				if !ok {
					return nil, errors.New("function in movingWindowList needs to return a bool")
				}
				if remove {
					startIndex++
				} else {
					mainList = append(mainList, li)
					break
				}
			}
		}
	}
	return NewList(mainList...), nil
}

func (l *List) containsItem(st funcGen.Stack[Value], item Value, fg *FunctionGenerator) (bool, error) {
	found := false
	err := l.iterable(st, func(value Value) error {
		eq, err := fg.equal(st, item, value)
		if err != nil {
			return err
		}
		if eq {
			found = true
			return iterator.SBC
		}
		return nil
	})
	if err != nil && err != iterator.SBC {
		return false, err
	}
	return found, nil
}

func (l *List) containsAllItems(st funcGen.Stack[Value], lookForList *List, fg *FunctionGenerator) (bool, error) {
	lookFor, err := lookForList.CopyToSlice(st)
	if err != nil {
		return false, err
	}

	if l.itemsPresent && len(l.items) < len(lookFor) {
		return false, nil
	}

	err = l.iterable(st, func(value Value) error {
		for i, lf := range lookFor {
			eq, err2 := fg.equal(st, lf, value)
			if err2 != nil {
				return err2
			}
			if eq {
				lookFor = append(lookFor[0:i], lookFor[i+1:]...)
				break
			}
		}
		if len(lookFor) == 0 {
			return iterator.SBC
		}
		return nil
	})
	if err != nil && err != iterator.SBC {
		return false, err
	}
	return len(lookFor) == 0, nil
}

type point struct {
	x, y float64
}

func interpolatePoints(points []point, x float64) float64 {
	n0 := 0
	n1 := len(points) - 1

	if x <= points[n0].x {
		return points[n0].y
	} else if x >= points[n1].x {
		return points[n1].y
	} else {
		for n1-n0 > 1 {
			n := (n0 + n1) / 2
			if x < points[n].x {
				n1 = n
			} else {
				n0 = n
			}
		}

		xr := (x - points[n0].x) / (points[n1].x - points[n0].x)
		y := points[n0].y + (points[n1].y-points[n0].y)*xr
		return y
	}
}

func (l *List) CreateInterpolation(st funcGen.Stack[Value]) (Value, error) {
	getXFunc, err := ToFunc("createInterpolation", st, 1, 1)
	if err != nil {
		return nil, err
	}
	getYFunc, err := ToFunc("createInterpolation", st, 2, 1)
	if err != nil {
		return nil, err
	}

	var points []point
	err = l.iterable(st, func(v Value) error {
		x, err := MustFloat(getXFunc.Eval(st, v))
		if err != nil {
			return err
		}
		if len(points) > 0 && x <= points[len(points)-1].x {
			return errors.New("x values in interpolation need to be increasing")
		}

		y, err := MustFloat(getYFunc.Eval(st, v))
		if err != nil {
			return err
		}
		points = append(points, point{x: x, y: y})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return Closure{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			fl, ok := st.Get(0).ToFloat()
			if !ok {
				return nil, errors.New("argument in interpolation needs to be a float")
			}
			return Float(interpolatePoints(points, fl)), nil
		},
		Args:   1,
		IsPure: true,
	}, nil
}

func (l *List) Linear(st funcGen.Stack[Value]) (Value, error) {
	getXFunc, err := ToFunc("linear", st, 1, 1)
	if err != nil {
		return nil, err
	}
	getYFunc, err := ToFunc("linear", st, 2, 1)
	if err != nil {
		return nil, err
	}

	sxi := 0.0
	syi := 0.0
	sxi2 := 0.0
	sxiyi := 0.0
	n := 0
	err = l.iterable(st, func(v Value) error {
		x, err := MustFloat(getXFunc.Eval(st, v))
		if err != nil {
			return err
		}
		y, err := MustFloat(getYFunc.Eval(st, v))
		if err != nil {
			return err
		}

		sxi += x
		syi += y
		sxi2 += x * x
		sxiyi += x * y
		n++

		return nil
	})
	if err != nil {
		return nil, err
	}

	a := (sxiyi - sxi*syi/float64(n)) / (sxi2 - sxi*sxi/float64(n))
	b := (syi - a*sxi) / float64(n)

	f := Closure{
		Func: func(stack funcGen.Stack[Value], closureStore []Value) (Value, error) {
			if x, ok := stack.Get(0).ToFloat(); ok {
				return Float(a*x + b), nil
			} else {
				return nil, errors.New("argument in linear needs to be a float")
			}
		},
		Args:   1,
		IsPure: true,
	}

	return NewMap(listMap.New[Value](2).
		Append("a", Float(a)).
		Append("b", Float(b)).
		Append("lineFunc", f)), nil
}

func (l *List) Set(st funcGen.Stack[Value]) (Value, error) {
	index, err := MustInt(st.Get(1), nil)
	if err != nil {
		return nil, err
	}
	sl, err := l.CopyToSlice(st)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(sl) {
		return nil, fmt.Errorf("index %d out of range", index)
	}
	sl[index] = st.Get(2)
	return NewList(sl...), nil
}

func createListMethods(fg *FunctionGenerator) MethodMap {
	return MethodMap{
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
		"sum": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Sum(stack, fg) }).
			SetMethodDescription("Returns the sum of all items in the list. Shorthand for reduce((a,b)->a+b)."),
		"mapReduce": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MapReduce(stack) }).
			SetMethodDescription("initialSum", "func(sum, item) sum",
				"MapReduce reduces the list to a single value. The initial value is given as the first argument. The function "+
					"is called with the initial value and the first item, and the result is used as the first argument for the "+
					"second item and so on."),
		"mean": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Mean(stack, fg) }).
			SetMethodDescription(
				"Returns the mean value of the list."),
		"min": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Min(stack, fg) }).
			SetMethodDescription("Returns the minimum value of the list."),
		"max": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Max(stack, fg) }).
			SetMethodDescription("Returns the maximum value of the list."),
		"minMax": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MinMax(stack, fg) }).
			SetMethodDescription("func(item) value",
				"Returns the minimum and maximum value of the list. The function is called for each item in the list and the "+
					"result is compared to the previous minimum and maximum."),
		"replaceList": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.ReplaceList(stack) }).
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
		"indexWhere": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.IndexWhere(stack) }).
			SetMethodDescription("func(item) condition",
				"Returns the index of the first occurrence of the given function returning true. If this never happens, -1 is returned."),
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
		"groupByEqual": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.GroupByEqual(stack, fg) }).
			SetMethodDescription("func(item) key", "Returns a list of lists grouped by the given function. "+
				"The function is called for each item in the list and the returned value is used as the key for the group. "+
				"The result is a list of maps with the keys 'key' and 'values'. The 'key' contains the value returned by the function "+
				"and 'values' contains a list of items that have the same key. "+
				"This method relies only on the Equal operator to determine if two keys are equal. This way no hash can be computed, "+
				"which makes this method much slower than the other groupBy methods, if the list is large (O(n²))."),
		"uniqueString": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.UniqueString(stack) }).
			SetMethodDescription("func(item) string", "Returns a list of unique strings returned by the given function."),
		"uniqueInt": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.UniqueInt(stack) }).
			SetMethodDescription("func(item) int", "Returns a list of unique integers returned by the given function."),
		"compact": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Compact(fg) }).
			SetMethodDescription("Returns a new list with the items compacted. " +
				"The given function is called for each successive pair of items in the list." +
				"If the items are equal, one is removed."),
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
		"order": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Order(stack, false, fg) }).
			SetMethodDescription("func(item) value",
				"Returns a new list with the items sorted in the order of the values returned by the given function. "+
					"The function is called for each item in the list and the returned values determine the order."),
		"orderRev": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Order(stack, true, fg) }).
			SetMethodDescription("func(item) value",
				"Returns a new list with the items sorted in the reverse order of the values returned by the given function. "+
					"The function is called for each item in the list and the returned values determine the order."),
		"orderLess": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.OrderLess(stack) }).
			SetMethodDescription("func(a, a) bool",
				"Returns a new list with the items sorted by the given function. "+
					"The function is called for pairs of items in the list and the returned bool needs to be true if a<b holds."),
		"reverse": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Reverse(stack) }).
			SetMethodDescription("Returns the list in reverse order."),
		"append": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Append(stack) }).
			SetMethodDescription("item", "Returns a new list with the given item appended. "+
				"If a list is to be created by adding element by element, this method is more efficient than using the '+' operator."),
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
					"An initial visitor is given as the first argument. The return value of the function is used as the new visitor."),
		"fsm": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.FSM(stack) }).
			SetMethodDescription("func(state, item) state",
				"Returns a new list with the given function applied to the items in the list. "+
					"The state is initialized with '{state:0}' and the function is called with the state and the item and returns the new state. "+
					"See also the function 'goto', which helps to create new state maps."),
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
		"set": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Set(stack) }).
			SetMethodDescription("index", "item", "Replaces the item at the given index with the given item. Returns the new list."),
		"size": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) {
			size, err := list.Size(stack)
			return Int(size), err
		}).
			SetMethodDescription("Returns the number of items in the list."),
		"first": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.First(stack) }).
			SetMethodDescription("Returns the first item in the list."),
		"single": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Single(stack) }).
			SetMethodDescription("Returns the first item in the list."),
		"last": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Last(stack) }).
			SetMethodDescription("Returns the last item in the list."),
		"eval": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list, list.Eval(stack) }).
			SetMethodDescription("Evaluates the list and stores all items in memory."),
		"string": MethodAtType(0, func(list *List, stack funcGen.Stack[Value]) (Value, error) {
			s, err := list.ToString(stack)
			return String(s), err
		}).
			SetMethodDescription("Returns the list as a string."),
		"movingWindow": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MovingWindow(stack) }).
			SetMethodDescription("func(item) float", "Returns a list of lists. "+
				"The inner lists contain all items that are close to each other. "+
				"Two items are close to each other if the given function returns a similar value for both items. "+
				"Similarity is defined as the absolute difference being smaller than 1."),
		"movingWindowRemove": MethodAtType(1, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.MovingWindowRemove(stack) }).
			SetMethodDescription("func([list of items]) bool", "Returns a list of lists. "+
				"The given remove-function is called with a sublist of items. At every call a new item from the original list is added to the sublist. "+
				"If the function returns true the first item of the sublist is removed and the function is called again until it returns false. "+
				"If the function returns false or if the sublist contains only one item, the sublist is added to the result."),
		"createInterpolation": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.CreateInterpolation(stack) }).
			SetMethodDescription("func(item) x", "func(item) y",
				"Returns a function that interpolates between the given points."),
		"linearReg": MethodAtType(2, func(list *List, stack funcGen.Stack[Value]) (Value, error) { return list.Linear(stack) }).
			SetMethodDescription("func(item) x", "func(item) y",
				"Returns a map containing the values a and b of the linear regression function y=a*x+b that fits the data points."),
		"binning": MethodAtType(5, Binning).
			SetMethodDescription("start", "size", "count", "indexFunc", "valueFunc",
				"Returns a map with the binning results. The index function must return the index of the bin for a specific element "+
					"of the list, and the value function must return the value to be added to the bin. "+
					"If only the number of times a value was in a bin is to be counted, the value function must return the constant one."),
		"binning2d": MethodAtType(9, Binning2d).
			SetMethodDescription("startX", "sizeX", "countX", "startY", "sizeYX", "countY", "indexFuncX", "indexFuncY", "valueFunc",
				"Returns a map with the binning results."),
		"collectBinning": MethodAtType(0, CollectBinning).
			SetMethodDescription("Sums up a list of binning results to create a total result."),
	}
}
func (l *List) GetType() Type {
	return ListTypeId
}
