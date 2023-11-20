package value

import (
	"bytes"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"sort"
	"strconv"
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

type Value interface {
	ToList() (List, bool)
	ToMap() (Map, bool)
	ToInt() (int, bool)
	ToFloat() (float64, bool)
	ToString() (string, bool)
	ToBool() (bool, bool)
	ToClosure() (funcGen.Function[Value], bool)
	GetMethod(name string) (funcGen.Function[Value], bool)
}

func NewList(items ...Value) List {
	return List{items: items, itemsPresent: true, iterable: func() iterator.Iterator[Value] {
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

func NewListFromIterable(li iterator.Iterable[Value]) List {
	return List{iterable: li, itemsPresent: false}
}

type List struct {
	items        []Value
	itemsPresent bool
	iterable     iterator.Iterable[Value]
}

func (l List) ToMap() (Map, bool) {
	return Map{}, false
}

func (l List) ToInt() (int, bool) {
	return 0, false
}

func (l List) ToFloat() (float64, bool) {
	return 0, false
}

func (l List) ToString() (string, bool) {
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

func (l List) ToBool() (bool, bool) {
	return false, false
}

func (l List) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (l List) ToList() (List, bool) {
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

func (l List) Accept(st funcGen.Stack[Value]) List {
	f := toFunc("accept", st, 1, 1)
	return NewListFromIterable(iterator.Filter[Value](l.iterable, func(v Value) bool {
		if accept, ok := f.Eval(st, v).ToBool(); ok {
			return accept
		}
		panic(fmt.Errorf("closure in accept does not return a bool"))
	}))
}

func (l List) Map(st funcGen.Stack[Value]) List {
	f := toFunc("map", st, 1, 1)
	return NewListFromIterable(iterator.MapAuto[Value, Value](l.iterable, func() func(i int, v Value) Value {
		return func(i int, v Value) Value {
			return f.Eval(st, v)
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

func (l List) Order(st funcGen.Stack[Value]) List {
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

func (l List) Combine(st funcGen.Stack[Value]) List {
	f := toFunc("combine", st, 1, 2)
	return NewListFromIterable(iterator.Combine[Value, Value](l.iterable, func(a, b Value) Value {
		st.Push(a)
		st.Push(b)
		return f.Func(st.CreateFrame(2), nil)
	}))
}

func (l List) IIr(st funcGen.Stack[Value]) List {
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

func (l List) Reduce(st funcGen.Stack[Value]) Value {
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

func (l List) Replace(st funcGen.Stack[Value]) Value {
	f := toFunc("replace", st, 1, 1)
	return f.Eval(st, l)
}

func (l List) GroupBy(st funcGen.Stack[Value]) Map {
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

func methodAtList(args int, method func(list List, stack funcGen.Stack[Value]) Value) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) Value {
		if obj, ok := stack.Get(0).ToList(); ok {
			return method(obj, stack)
		}
		panic("call of list method on non list")
	}, Args: args, IsPure: true}
}

var ListMethods = map[string]funcGen.Function[Value]{
	"accept":  methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Accept(stack) }),
	"map":     methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Map(stack) }),
	"reduce":  methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Reduce(stack) }),
	"replace": methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Replace(stack) }),
	"combine": methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Combine(stack) }),
	"group":   methodAtList(3, func(list List, stack funcGen.Stack[Value]) Value { return list.GroupBy(stack) }),
	"order":   methodAtList(2, func(list List, stack funcGen.Stack[Value]) Value { return list.Order(stack) }),
	"iir":     methodAtList(3, func(list List, stack funcGen.Stack[Value]) Value { return list.IIr(stack) }),
	"size":    methodAtList(1, func(list List, stack funcGen.Stack[Value]) Value { return Int(list.Size()) }),
}

func (l List) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := ListMethods[name]
	return m, ok
}

type Map struct {
	M MapImplementation[Value]
}

func (v Map) ToList() (List, bool) {
	return List{}, false
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

func (v Map) List() List {
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

type Closure funcGen.Function[Value]

func (c Closure) ToList() (List, bool) {
	return List{}, false
}

func (c Closure) ToMap() (Map, bool) {
	return Map{}, false
}

func (c Closure) ToInt() (int, bool) {
	return 0, false
}

func (c Closure) ToFloat() (float64, bool) {
	return 0, false
}

func (c Closure) ToString() (string, bool) {
	return "", false
}

func (c Closure) ToBool() (bool, bool) {
	return false, false
}

func (c Closure) GetMethod(string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (c Closure) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value](c), true
}

type Bool bool

func (b Bool) ToList() (List, bool) {
	return List{}, false
}

func (b Bool) ToMap() (Map, bool) {
	return Map{}, false
}

func (b Bool) ToInt() (int, bool) {
	return 0, false
}

func (b Bool) ToFloat() (float64, bool) {
	return 0, false
}

func (b Bool) ToString() (string, bool) {
	if b {
		return "true", true
	}
	return "false", true
}

func (b Bool) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (b Bool) GetMethod(string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (b Bool) ToBool() (bool, bool) {
	return bool(b), true
}

type Float float64

func (f Float) ToList() (List, bool) {
	return List{}, false
}

func (f Float) ToMap() (Map, bool) {
	return Map{}, false
}

func (f Float) ToString() (string, bool) {
	return "", false
}

func (f Float) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (f Float) GetMethod(string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (f Float) ToBool() (bool, bool) {
	if f != 0 {
		return true, true
	}
	return false, true
}

func (f Float) ToInt() (int, bool) {
	return int(f), true
}

func (f Float) ToFloat() (float64, bool) {
	return float64(f), true
}

type Int int

func (i Int) ToList() (List, bool) {
	return List{}, false
}

func (i Int) ToMap() (Map, bool) {
	return Map{}, false
}

func (i Int) ToString() (string, bool) {
	return strconv.Itoa(int(i)), true
}

func (i Int) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (i Int) GetMethod(string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (i Int) ToBool() (bool, bool) {
	if i != 0 {
		return true, true
	}
	return false, true
}

func (i Int) ToInt() (int, bool) {
	return int(i), true
}

func (i Int) ToFloat() (float64, bool) {
	return float64(i), true
}

type String string

func (s String) ToList() (List, bool) {
	return List{}, false
}

func (s String) ToMap() (Map, bool) {
	return Map{}, false
}

func (s String) ToInt() (int, bool) {
	return 0, false
}

func (s String) ToFloat() (float64, bool) {
	return 0, false
}

func (s String) ToBool() (bool, bool) {
	return false, false
}

func (s String) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (s String) GetMethod(string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (s String) ToString() (string, bool) {
	return string(s), true
}

type factory struct{}

func (f factory) ParseNumber(n string) (Value, error) {
	i, err := strconv.Atoi(n)
	if err == nil {
		return Int(i), nil
	}
	fl, err := strconv.ParseFloat(n, 64)
	if err == nil {
		return Float(fl), nil
	}
	return nil, err
}

func (f factory) FromString(s string) Value {
	return String(s)
}

func (f factory) GetMethod(value Value, methodName string) (funcGen.Function[Value], error) {
	if m, ok := value.GetMethod(methodName); ok {
		return m, nil
	} else {
		return funcGen.Function[Value]{}, fmt.Errorf("method not found: %s", methodName)
	}
}

func (f factory) FromClosure(c funcGen.Function[Value]) Value {
	return Closure(c)
}

func (f factory) ToClosure(value Value) (funcGen.Function[Value], bool) {
	return value.ToClosure()
}

func (f factory) FromMap(items listMap.ListMap[Value]) Value {
	return Map{M: items}
}

func (f factory) AccessMap(mapValue Value, key string) (Value, error) {
	if m, ok := mapValue.ToMap(); ok {
		if v, ok := m.Get(key); ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("key %s not found in map", key)
		}
	} else {
		return nil, fmt.Errorf("not a map")
	}
}

func (f factory) IsMap(mapValue Value) bool {
	_, ok := mapValue.ToMap()
	return ok
}

func (f factory) FromList(items []Value) Value {
	return NewList(items...)
}

func (f factory) AccessList(list Value, index Value) (Value, error) {
	if l, ok := list.ToList(); ok {
		if i, ok := index.ToInt(); ok {
			if i < 0 {
				return nil, fmt.Errorf("negative list index")
			} else if i >= l.Size() {
				return nil, fmt.Errorf("index out of bounds")
			} else {
				return l.items[i], nil
			}
		} else {
			return nil, fmt.Errorf("not an index")
		}
	} else {
		return nil, fmt.Errorf("not a list")
	}
}

func (f factory) Generate(ast parser2.AST, gc funcGen.GeneratorContext, g *funcGen.FunctionGenerator[Value]) (funcGen.Func[Value], error) {
	if op, ok := ast.(*parser2.Operate); ok {
		// AND and OR with short evaluation
		switch op.Operator {
		case "&":
			aFunc, err := g.GenerateFunc(op.A, gc)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, gc)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) Value {
				aVal := aFunc(st, cs)
				if a, ok := aVal.ToBool(); ok {
					if !a {
						return Bool(false)
					} else {
						bVal := bFunc(st, cs)
						if b, ok := bVal.ToBool(); ok {
							return Bool(b)
						} else {
							panic(fmt.Errorf("not a bool: %v", bVal))
						}
					}
				} else {
					panic(fmt.Errorf("not a bool: %v", aVal))
				}
			}, nil
		case "|":
			aFunc, err := g.GenerateFunc(op.A, gc)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, gc)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) Value {
				aVal := aFunc(st, cs)
				if a, ok := aVal.ToBool(); ok {
					if a {
						return Bool(true)
					} else {
						bVal := bFunc(st, cs)
						if b, ok := bVal.ToBool(); ok {
							return Bool(b)
						} else {
							panic(fmt.Errorf("not a bool: %v", bVal))
						}
					}
				} else {
					panic(fmt.Errorf("not a bool: %v", aVal))
				}
			}, nil
		}
	}
	return nil, nil
}

var theFactory = factory{}

func SetUpParser(fc *funcGen.FunctionGenerator[Value]) *funcGen.FunctionGenerator[Value] {
	fc.ModifyParser(func(p *parser2.Parser[Value]) {
		p.SetNumberParser(theFactory)
	})
	return fc
}

func simpleOnlyFloatFunc(name string, f func(float64) float64) funcGen.Function[Value] {
	return funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) Value {
			v := st.Get(0)
			if fl, ok := v.ToFloat(); ok {
				return Float(f(fl))
			}
			panic(fmt.Errorf("%s not alowed on %v", name, v))
		},
		Args:   1,
		IsPure: true,
	}
}

func New() *funcGen.FunctionGenerator[Value] {
	return funcGen.New[Value]().
		AddConstant("pi", Float(math.Pi)).
		AddConstant("true", Bool(true)).
		AddConstant("false", Bool(false)).
		SetListHandler(theFactory).
		SetMapHandler(theFactory).
		SetClosureHandler(theFactory).
		SetMethodHandler(theFactory).
		SetCustomGenerator(theFactory).
		SetStringConverter(theFactory).
		SetIsEqual(Equal).
		SetToBool(func(c Value) (bool, bool) { return c.ToBool() }).
		AddOp("|", true, Or).
		AddOp("&", true, And).
		AddOp("=", true, func(a Value, b Value) Value { return Bool(Equal(a, b)) }).
		AddOp("!=", true, func(a, b Value) Value { return Bool(!Equal(a, b)) }).
		AddOp("<", false, Less).
		AddOp(">", false, Swap(Less)).
		AddOp("<=", false, LessEqual).
		AddOp(">=", false, Swap(LessEqual)).
		AddOp("+", false, Add).
		AddOp("-", false, Sub).
		AddOp("*", true, Mul).
		AddOp("/", false, Div).
		AddOp("^", false, Pow).
		AddUnary("-", func(a Value) Value { return Neg(a) }).
		AddUnary("!", func(a Value) Value { return Not(a) }).
		AddStaticFunction("abs", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v < 0 {
						return -v
					}
					return v
				}
				if f, ok := v.ToFloat(); ok {
					return Float(math.Abs(f))
				}
				panic(fmt.Errorf("abs not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("sqr", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v * v
				}
				if f, ok := v.ToFloat(); ok {
					return Float(f * f)
				}
				panic(fmt.Errorf("sqr not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("round", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v
				}
				if f, ok := v.ToFloat(); ok {
					return Int(math.Round(f))
				}
				panic(fmt.Errorf("sqr not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("list", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if size, ok := v.ToInt(); ok {
					return NewListFromIterable(iterator.Generate(size, func(i int) Value { return Int(i) }))
				}
				panic(fmt.Errorf("list not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("sqrt", simpleOnlyFloatFunc("sqrt", func(x float64) float64 { return math.Sqrt(x) })).
		AddStaticFunction("ln", simpleOnlyFloatFunc("ln", func(x float64) float64 { return math.Log(x) })).
		AddStaticFunction("exp", simpleOnlyFloatFunc("exp", func(x float64) float64 { return math.Exp(x) })).
		AddStaticFunction("sin", simpleOnlyFloatFunc("sin", func(x float64) float64 { return math.Sin(x) })).
		AddStaticFunction("cos", simpleOnlyFloatFunc("cos", func(x float64) float64 { return math.Cos(x) })).
		AddStaticFunction("tan", simpleOnlyFloatFunc("tan", func(x float64) float64 { return math.Tan(x) })).
		AddStaticFunction("asin", simpleOnlyFloatFunc("asin", func(x float64) float64 { return math.Asin(x) })).
		AddStaticFunction("acos", simpleOnlyFloatFunc("acos", func(x float64) float64 { return math.Acos(x) })).
		AddStaticFunction("atan", simpleOnlyFloatFunc("atan", func(x float64) float64 { return math.Atan(x) }))

}
