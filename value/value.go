package value

import (
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"strconv"
)

type MapImplementation[V any] interface {
	Get(key string) (V, bool)
}

type Value interface {
	ToList() ([]Value, bool)
	ToMap() (MapImplementation[Value], bool)
	ToInt() (int, bool)
	ToFloat() (float64, bool)
	ToString() (string, bool)
	ToBool() (bool, bool)
	ToClosure() (funcGen.Function[Value], bool)
	GetMethod(name string) (funcGen.Function[Value], bool)
}

type Empty struct {
}

func (e Empty) ToList() ([]Value, bool) {
	return nil, false
}

func (e Empty) ToMap() (MapImplementation[Value], bool) {
	return nil, false
}

func (e Empty) ToInt() (int, bool) {
	return 0, false
}

func (e Empty) ToFloat() (float64, bool) {
	return 0, false
}

func (e Empty) ToString() (string, bool) {
	return "", false
}

func (e Empty) ToBool() (bool, bool) {
	return false, false
}

func (e Empty) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (e Empty) GetMethod(name string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

type List struct {
	Empty
	items []Value
}

func (l List) ToList() ([]Value, bool) {
	return l.items, true
}

func (l List) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := ListMethods[name]
	return m, ok
}

var ListMethods = map[string]funcGen.Function[Value]{
	"map":    {Func: ListMap, Args: 2, IsPure: true},
	"reduce": {Func: ListReduce, Args: 2, IsPure: true},
}

func ListMap(stack funcGen.Stack[Value], closureStore []Value) Value {
	list, ok := stack.Get(0).ToList()
	if !ok {
		panic("map call on not list!")
	}
	c, ok := stack.Get(1).ToClosure()
	if !ok {
		panic("argument of map needs to be a closure")
	}
	if c.Args != 1 {
		panic("map requires closure with one argument")
	}
	st := funcGen.NewEmptyStack[Value]()
	var m = make([]Value, len(list))
	for i, e := range list {
		st.Push(e)
		m[i] = c.Func(st.CreateFrame(1), nil)
	}
	return List{items: m}
}

func ListReduce(stack funcGen.Stack[Value], closureStore []Value) Value {
	list, ok := stack.Get(0).ToList()
	if !ok {
		panic("map call on not list!")
	}
	c, ok := stack.Get(1).ToClosure()
	if !ok {
		panic("argument of map needs to be a closure")
	}
	if c.Args != 2 {
		panic("reduce requires closure with two arguments")
	}
	var red Value
	st := funcGen.NewEmptyStack[Value]()
	for i, e := range list {
		if i == 0 {
			red = e
		} else {
			st.Push(red)
			st.Push(e)
			red = c.Func(st.CreateFrame(2), nil)
		}
	}
	return red
}

type Map struct {
	Empty
	M MapImplementation[Value]
}

func (v Map) ToMap() (MapImplementation[Value], bool) {
	return v.M, true
}

var MapMethods = make(map[string]funcGen.Function[Value])

func (v Map) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := MapMethods[name]
	return m, ok
}

type Closure struct {
	Empty
	closure funcGen.Function[Value]
}

func (c Closure) ToClosure() (funcGen.Function[Value], bool) {
	return c.closure, true
}

type Bool struct {
	Empty
	B bool
}

func (b Bool) ToBool() (bool, bool) {
	return b.B, true
}

type Float struct {
	Empty
	F float64
}

func (f Float) ToBool() (bool, bool) {
	if f.F != 0 {
		return true, true
	}
	return false, true
}

func (f Float) ToInt() (int, bool) {
	return int(f.F), true
}

func (f Float) ToFloat() (float64, bool) {
	return f.F, true
}

type Int struct {
	Empty
	I int
}

func (i Int) ToBool() (bool, bool) {
	if i.I != 0 {
		return true, true
	}
	return false, true
}

func (i Int) ToInt() (int, bool) {
	return i.I, true
}

func (i Int) ToFloat() (float64, bool) {
	return float64(i.I), true
}

type String struct {
	Empty
	S string
}

func (s String) ToString() (string, bool) {
	return s.S, true
}

type factory struct{}

func (f factory) ParseNumber(n string) (Value, error) {
	i, err := strconv.Atoi(n)
	if err == nil {
		return Int{I: i}, nil
	}
	fl, err := strconv.ParseFloat(n, 64)
	if err == nil {
		return Float{F: fl}, nil
	}
	return nil, err
}

func (f factory) FromString(s string) Value {
	return String{S: s}
}

func (f factory) GetMethod(value Value, methodName string) (funcGen.Function[Value], error) {
	if m, ok := value.GetMethod(methodName); ok {
		return m, nil
	} else {
		return funcGen.Function[Value]{}, fmt.Errorf("method not found: %s", methodName)
	}
}

func (f factory) FromClosure(c funcGen.Function[Value]) Value {
	return Closure{closure: c}
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
	return List{items: items}
}

func (f factory) AccessList(list Value, index Value) (Value, error) {
	if l, ok := list.ToList(); ok {
		if i, ok := index.ToInt(); ok {
			if i < 0 {
				return nil, fmt.Errorf("negative list index")
			} else if i >= len(l) {
				return nil, fmt.Errorf("index out of bounds")
			} else {
				return l[i], nil
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
						return Bool{B: false}
					} else {
						bVal := bFunc(st, cs)
						if b, ok := bVal.ToBool(); ok {
							return Bool{B: b}
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
						return Bool{B: true}
					} else {
						bVal := bFunc(st, cs)
						if b, ok := bVal.ToBool(); ok {
							return Bool{B: b}
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

func New() *funcGen.FunctionGenerator[Value] {
	theFactory := factory{}
	return funcGen.New[Value]().
		SetListHandler(theFactory).
		SetMapHandler(theFactory).
		SetClosureHandler(theFactory).
		SetMethodHandler(theFactory).
		SetCustomGenerator(theFactory).
		SetStringConverter(theFactory).
		SetIsEqual(Equal).
		SetToBool(func(c Value) (bool, bool) { return c.ToBool() }).
		AddOp("=", true, func(a Value, b Value) Value { return Bool{B: Equal(a, b)} }).
		AddOp("!=", true, func(a, b Value) Value { return Bool{B: !Equal(a, b)} }).
		AddOp("<", false, Less).
		AddOp(">", false, Swap(Less)).
		AddOp("<=", false, LessEqual).
		AddOp(">=", false, Swap(LessEqual)).
		AddOp("+", false, Add).
		AddOp("-", false, Sub).
		AddOp("*", true, Mul).
		AddOp("/", false, Div).
		AddUnary("-", func(a Value) Value { return Neg(a) }).
		AddUnary("!", func(a Value) Value { return Not(a) }).
		ModifyParser(func(p *parser2.Parser[Value]) {
			p.SetNumberParser(theFactory)
		}).
		AddStaticFunction("abs", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v.I < 0 {
						return Int{I: -v.I}
					}
					return v
				}
				if f, ok := v.ToFloat(); ok {
					return Float{F: math.Abs(f)}
				}
				panic(fmt.Errorf("abs not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("sqrt", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float{F: math.Sqrt(f)}
				}
				panic(fmt.Errorf("sqrt not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("ln", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float{F: math.Log(f)}
				}
				panic(fmt.Errorf("ln not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("exp", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float{F: math.Exp(f)}
				}
				panic(fmt.Errorf("exp not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		})

}
