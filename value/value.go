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

type Value interface {
	ToList() (*List, bool)
	ToMap() (Map, bool)
	ToInt() (int, bool)
	ToFloat() (float64, bool)
	String() string
	ToBool() (bool, bool)
	ToClosure() (funcGen.Function[Value], bool)
	GetMethod(name string) (funcGen.Function[Value], error)
}

func MethodAtType[V Value](args int, method func(obj V, stack funcGen.Stack[Value]) Value) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) Value {
		if obj, ok := stack.Get(0).(V); ok {
			return method(obj, stack)
		}
		panic("internal error: call of method on wrong type")
	}, Args: args + 1, IsPure: true}
}

type MethodMap map[string]funcGen.Function[Value]

func (mm MethodMap) Get(name string) (funcGen.Function[Value], error) {
	if m, ok := mm[name]; ok {
		return m, nil
	}

	type fes struct {
		name string
		fu   funcGen.Function[Value]
	}
	var l []fes
	for k, f := range mm {
		l = append(l, fes{name: k, fu: f})
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].name < l[j].name
	})
	var b bytes.Buffer
	for _, fe := range l {
		b.WriteRune('\n')
		fe.fu.Description.WriteTo(&b, fe.name)
	}
	return funcGen.Function[Value]{}, fmt.Errorf("method '%s' not found; available are:\n%s", name, b.String())
}

type Closure funcGen.Function[Value]

func (c Closure) ToList() (*List, bool) {
	return nil, false
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

func (c Closure) String() string {
	return "<function>"
}

func (c Closure) ToBool() (bool, bool) {
	return false, false
}

var ClosureMethods = MethodMap{
	"string": MethodAtType(0, func(c Closure, stack funcGen.Stack[Value]) Value { return String(c.String()) }),
}

func (c Closure) GetMethod(name string) (funcGen.Function[Value], error) {
	return ClosureMethods.Get(name)
}

func (c Closure) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value](c), true
}

type Bool bool

func (b Bool) ToList() (*List, bool) {
	return nil, false
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

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b Bool) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var BoolMethods = MethodMap{
	"string": MethodAtType(0, func(b Bool, stack funcGen.Stack[Value]) Value { return String(b.String()) }),
}

func (b Bool) GetMethod(name string) (funcGen.Function[Value], error) {
	return BoolMethods.Get(name)
}

func (b Bool) ToBool() (bool, bool) {
	return bool(b), true
}

type Float float64

func (f Float) ToList() (*List, bool) {
	return nil, false
}

func (f Float) ToMap() (Map, bool) {
	return Map{}, false
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'g', -1, 64)
}

func (f Float) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var FloatMethods = MethodMap{
	"string": MethodAtType(0, func(f Float, stack funcGen.Stack[Value]) Value { return String(f.String()) }),
}

func (f Float) GetMethod(name string) (funcGen.Function[Value], error) {
	return FloatMethods.Get(name)
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

func (i Int) ToList() (*List, bool) {
	return nil, false
}

func (i Int) ToMap() (Map, bool) {
	return Map{}, false
}

func (i Int) String() string {
	return strconv.Itoa(int(i))
}

func (i Int) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var IntMethods = MethodMap{
	"string": MethodAtType(0, func(i Int, stack funcGen.Stack[Value]) Value { return String(i.String()) }),
}

func (i Int) GetMethod(name string) (funcGen.Function[Value], error) {
	return IntMethods.Get(name)
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
	m, err := value.GetMethod(methodName)
	if err != nil {
		return funcGen.Function[Value]{}, err
	} else {
		return m, nil
	}
}

func (f factory) FromClosure(c funcGen.Function[Value]) Value {
	return Closure(c)
}

func (f factory) ToClosure(value Value) (funcGen.Function[Value], bool) {
	return value.ToClosure()
}

func (f factory) FromMap(items listMap.ListMap[Value]) Value {
	return Map{m: items}
}

func (f factory) AccessMap(mapValue Value, key string) (Value, error) {
	if m, ok := mapValue.ToMap(); ok {
		if v, ok := m.Get(key); ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("key %s not found in map", key)
		}
	} else {
		return nil, fmt.Errorf("'.%s' not possible; not a map", key)
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
				return nil, fmt.Errorf("index out of bounds %d>=size(%d)", i, l.Size())
			} else {
				return l.items[i], nil
			}
		} else {
			return nil, fmt.Errorf("not an int: %v", index)
		}
	} else {
		return nil, fmt.Errorf("not a list: %v", list)
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
		AddOp("~", false, In).
		AddOp("<", false, func(a Value, b Value) Value { return Bool(Less(a, b)) }).
		AddOp(">", false, func(a Value, b Value) Value { return Bool(Less(b, a)) }).
		AddOp("<=", false, LessEqual).
		AddOp(">=", false, Swap(LessEqual)).
		AddOp("+", false, Add).
		AddOp("-", false, Sub).
		AddOp("<<", false, Left).
		AddOp(">>", false, Right).
		AddOp("*", true, Mul).
		AddOp("%", false, Mod).
		AddOp("/", false, Div).
		AddOp("^", false, Pow).
		AddUnary("-", func(a Value) Value { return Neg(a) }).
		AddUnary("!", func(a Value) Value { return Not(a) }).
		AddStaticFunction("string", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				return String(st.Get(0).String())
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("float", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float(f)
				}
				panic(fmt.Errorf("float not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
		AddStaticFunction("int", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) Value {
				v := st.Get(0)
				if i, ok := v.ToInt(); ok {
					return Int(i)
				}
				panic(fmt.Errorf("int not alowed on %v", v))
			},
			Args:   1,
			IsPure: true,
		}).
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
		AddStaticFunction("sprintf", funcGen.Function[Value]{Func: sprintf, Args: -1, IsPure: true}).
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

func sprintf(st funcGen.Stack[Value], cs []Value) Value {
	switch st.Size() {
	case 0:
		return String("")
	case 1:
		return String(fmt.Sprint(st.Get(0)))
	default:
		if s, ok := st.Get(0).(String); ok {
			values := make([]any, st.Size()-1)
			for i := 1; i < st.Size(); i++ {
				values[i-1] = st.Get(i)
			}
			return String(fmt.Sprintf(string(s), values...))
		} else {
			panic("sprintf requires string as first argument")
		}
	}
}
