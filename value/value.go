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

func NewList(items ...Value) Value {
	return List{items: items}
}

type List struct {
	items []Value
}

func (l List) ToMap() (MapImplementation[Value], bool) {
	return nil, false
}

func (l List) ToInt() (int, bool) {
	return 0, false
}

func (l List) ToFloat() (float64, bool) {
	return 0, false
}

func (l List) ToString() (string, bool) {
	return "", false
}

func (l List) ToBool() (bool, bool) {
	return false, false
}

func (l List) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
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
	"size":   {Func: ListSize, Args: 1, IsPure: true},
}

func ListSize(stack funcGen.Stack[Value], closureStore []Value) Value {
	list, ok := stack.Get(0).ToList()
	if !ok {
		panic("size call on not list!")
	}
	return Int(len(list))
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
	M MapImplementation[Value]
}

func (v Map) ToList() ([]Value, bool) {
	return nil, false
}

func (v Map) ToInt() (int, bool) {
	return 0, false
}

func (v Map) ToFloat() (float64, bool) {
	return 0, false
}

func (v Map) ToString() (string, bool) {
	return "", false
}

func (v Map) ToBool() (bool, bool) {
	return false, false
}

func (v Map) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (v Map) ToMap() (MapImplementation[Value], bool) {
	return v.M, true
}

var MapMethods = make(map[string]funcGen.Function[Value])

func (v Map) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := MapMethods[name]
	return m, ok
}

type Closure funcGen.Function[Value]

func (c Closure) ToList() ([]Value, bool) {
	return nil, false
}

func (c Closure) ToMap() (MapImplementation[Value], bool) {
	return nil, false
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

func (c Closure) GetMethod(name string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (c Closure) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value](c), true
}

type Bool bool

func (b Bool) ToList() ([]Value, bool) {
	return nil, false
}

func (b Bool) ToMap() (MapImplementation[Value], bool) {
	return nil, false
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

func (b Bool) GetMethod(name string) (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (b Bool) ToBool() (bool, bool) {
	return bool(b), true
}

type Float float64

func (f Float) ToList() ([]Value, bool) {
	return nil, false
}

func (f Float) ToMap() (MapImplementation[Value], bool) {
	return nil, false
}

func (f Float) ToString() (string, bool) {
	return "", false
}

func (f Float) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (f Float) GetMethod(name string) (funcGen.Function[Value], bool) {
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

func (i Int) ToList() ([]Value, bool) {
	return nil, false
}

func (i Int) ToMap() (MapImplementation[Value], bool) {
	return nil, false
}

func (i Int) ToString() (string, bool) {
	return "", false
}

func (i Int) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (i Int) GetMethod(name string) (funcGen.Function[Value], bool) {
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

func (s String) ToList() ([]Value, bool) {
	return nil, false
}

func (s String) ToMap() (MapImplementation[Value], bool) {
	return nil, false
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

func (s String) GetMethod(name string) (funcGen.Function[Value], bool) {
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
		AddUnary("-", func(a Value) Value { return Neg(a) }).
		AddUnary("!", func(a Value) Value { return Not(a) }).
		ModifyParser(func(p *parser2.Parser[Value]) {
			p.SetNumberParser(theFactory)
		}).
		AddStaticFunction("abs", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v < 0 {
						return Int(-v)
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
		AddStaticFunction("sqrt", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float(math.Sqrt(f))
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
					return Float(math.Log(f))
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
					return Float(math.Exp(f))
				}
				panic(fmt.Errorf("exp not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		})

}
