package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"math/rand"
	"sort"
	"strconv"
)

type Value interface {
	ToList() (*List, bool)
	ToMap() (Map, bool)
	ToInt() (int, bool)
	ToFloat() (float64, bool)
	ToString(st funcGen.Stack[Value]) (string, error)
	ToBool() (bool, bool)
	ToClosure() (funcGen.Function[Value], bool)
	GetMethod(name string) (funcGen.Function[Value], error)
}

func MethodAtType[V Value](args int, method func(obj V, stack funcGen.Stack[Value]) (Value, error)) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) (Value, error) {
		if obj, ok := stack.Get(0).(V); ok {
			return method(obj, stack)
		}
		return nil, fmt.Errorf("internal error: call of method on wrong type")
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

const NIL = nilType(0)

type nilType int

func (n nilType) ToList() (*List, bool) {
	return nil, false
}

func (n nilType) ToMap() (Map, bool) {
	return Map{}, false
}

func (n nilType) ToInt() (int, bool) {
	return 0, false
}

func (n nilType) ToFloat() (float64, bool) {
	return 0, false
}

func (n nilType) ToString(funcGen.Stack[Value]) (string, error) {
	return "nil", nil
}

func (n nilType) String() string {
	return "nil"
}

func (n nilType) ToBool() (bool, bool) {
	return false, false
}

func (n nilType) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (n nilType) GetMethod(string) (funcGen.Function[Value], error) {
	return funcGen.Function[Value]{}, fmt.Errorf("nil has no methods")
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

func (c Closure) ToString(funcGen.Stack[Value]) (string, error) {
	return "", errors.New("a function has no string representation")
}

func (c Closure) ToBool() (bool, bool) {
	return false, false
}

var ClosureMethods = MethodMap{
	"args": MethodAtType(0, func(c Closure, stack funcGen.Stack[Value]) (Value, error) { return Int(c.Args), nil }).
		SetMethodDescription("Returns the number of arguments the function takes."),
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

func (b Bool) ToString(funcGen.Stack[Value]) (string, error) {
	if b {
		return "true", nil
	}
	return "false", nil
}

func (b Bool) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var BoolMethods = MethodMap{
	"string": MethodAtType(0, func(b Bool, stack funcGen.Stack[Value]) (Value, error) {
		s, err := b.ToString(stack)
		return String(s), err
	}).
		SetMethodDescription("Returns the string 'true' or 'false'."),
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

func (f Float) ToString(funcGen.Stack[Value]) (string, error) {
	return strconv.FormatFloat(float64(f), 'g', -1, 64), nil
}

func (f Float) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var FloatMethods = MethodMap{
	"string": MethodAtType(0, func(f Float, stack funcGen.Stack[Value]) (Value, error) {
		s, err := f.ToString(stack)
		return String(s), err
	}).
		SetMethodDescription("Returns a string representation of the float."),
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

func (i Int) ToString(funcGen.Stack[Value]) (string, error) {
	return strconv.Itoa(int(i)), nil
}

func (i Int) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

var IntMethods = MethodMap{
	"string": MethodAtType(0, func(i Int, stack funcGen.Stack[Value]) (Value, error) {
		s, err := i.ToString(stack)
		return String(s), err
	}).
		SetMethodDescription("Returns a string representation of the int."),
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
			return nil, fmt.Errorf("key '%s' not found in map", key)
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
			} else {
				size, err := l.Size(funcGen.NewEmptyStack[Value]())
				if err != nil {
					return nil, err
				}
				if i >= size {
					return nil, fmt.Errorf("index out of bounds %d>=size(%d)", i, size)
				} else {
					return l.items[i], nil
				}
			}
		} else {
			return nil, fmt.Errorf("not an int: %v", index)
		}
	} else {
		return nil, fmt.Errorf("not a list: %v", list)
	}
}

func (f factory) Generate(ast parser2.AST, gc funcGen.GeneratorContext, g *funcGen.FunctionGenerator[Value]) (funcGen.ParserFunc[Value], error) {
	if tc, ok := ast.(*parser2.TryCatch); ok {
		if cl, ok := tc.Catch.(*parser2.ClosureLiteral); ok && len(cl.Names) == 1 {
			tryFunc, err := g.GenerateFunc(tc.Try, gc)
			if err != nil {
				return nil, tc.EnhanceErrorf(err, "error in try expression")
			}
			catchFunc, err := g.GenerateFunc(tc.Catch, gc)
			if err != nil {
				return nil, tc.EnhanceErrorf(err, "error in catch expression")
			}
			l := tc.GetLine()
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				tryVal, tryErr := tryFunc(st, cs)
				if tryErr == nil {
					return tryVal, nil
				}
				catchVal, err := catchFunc(st, cs)
				if err != nil {
					return nil, l.EnhanceErrorf(err, "error in getting catch function")
				}
				theFunc, ok := g.ExtractFunction(catchVal)
				if !ok || theFunc.Args != 1 {
					// impossible because condition is checked above
					return nil, l.Errorf("internal catch error")
				}
				st.Push(String(tryErr.Error()))
				return theFunc.Func(st.CreateFrame(1), cs)
			}, nil
		}
	}
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
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				aVal, err := aFunc(st, cs)
				if err != nil {
					return nil, err
				}
				if a, ok := aVal.ToBool(); ok {
					if !a {
						return Bool(false), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.ToBool(); ok {
							return Bool(b), nil
						} else {
							return nil, fmt.Errorf("not a bool: %v", bVal)
						}
					}
				} else {
					return nil, fmt.Errorf("not a bool: %v", aVal)
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
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				aVal, err := aFunc(st, cs)
				if err != nil {
					return nil, err
				}
				if a, ok := aVal.ToBool(); ok {
					if a {
						return Bool(true), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.ToBool(); ok {
							return Bool(b), nil
						} else {
							return nil, fmt.Errorf("not a bool: %v", bVal)
						}
					}
				} else {
					return nil, fmt.Errorf("not a bool: %v", aVal)
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
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			v := st.Get(0)
			if fl, ok := v.ToFloat(); ok {
				return Float(f(fl)), nil
			}
			return nil, fmt.Errorf("%s not alowed on %v", name, v)
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("float", "The mathematical "+name+" function.")
}

func New() *funcGen.FunctionGenerator[Value] {
	return funcGen.New[Value]().
		AddConstant("nil", NIL).
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
		AddOp("=", true, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
			equal, err := Equal(st, a, b)
			return Bool(equal), err
		}).
		AddOp("!=", true, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
			equal, err := Equal(st, a, b)
			return Bool(!equal), err
		}).
		AddOp("~", false, In).
		AddOp("<", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
			less, err := Less(st, a, b)
			if err != nil {
				return nil, err
			}
			return Bool(less), nil
		}).
		AddOp(">", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
			less, err := Less(st, b, a)
			if err != nil {
				return nil, err
			}
			return Bool(less), nil
		}).
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
		AddUnary("-", func(a Value) (Value, error) { return Neg(a) }).
		AddUnary("!", func(a Value) (Value, error) { return Not(a) }).
		AddStaticFunction("string", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				s, err := st.Get(0).ToString(st)
				return String(s), err
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the string representation of the value.")).
		AddStaticFunction("float", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float(f), nil
				}
				return nil, fmt.Errorf("float not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the float representation of the value.")).
		AddStaticFunction("int", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if i, ok := v.ToInt(); ok {
					return Int(i), nil
				}
				return nil, fmt.Errorf("int not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the int representation of the value.")).
		AddStaticFunction("abs", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v < 0 {
						return -v, nil
					}
					return v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Float(math.Abs(f)), nil
				}
				return nil, fmt.Errorf("abs not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "If value is negative, returns -value. Otherwise returns the value unchanged.")).
		AddStaticFunction("sqr", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v * v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Float(f * f), nil
				}
				return nil, fmt.Errorf("sqr not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the square of the value.")).
		AddStaticFunction("rnd", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if n, ok := v.ToInt(); ok {
					return Int(rand.Intn(n)), nil
				}
				return nil, errors.New("rnd only allowed on int")
			},
			Args:   1,
			IsPure: false,
		}.SetDescription("n", "Returns a random integer between 0 and n-1.")).
		AddStaticFunction("round", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Int(math.Round(f)), nil
				}
				return nil, fmt.Errorf("sqr not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the value rounded to the nearest integer.")).
		AddStaticFunction("min", funcGen.Function[Value]{
			Func:   minFunc,
			Args:   -1,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the smaller of a and b.")).
		AddStaticFunction("max", funcGen.Function[Value]{
			Func:   maxFunc,
			Args:   -1,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the larger of a and b.")).
		AddStaticFunction("createLowPass", funcGen.Function[Value]{
			Func:   createLowPass,
			Args:   4,
			IsPure: true,
		}.SetDescription("name", "func(p) float", "func(p) float", "tau", "Returns a low pass filter creating signal [name]")).
		AddStaticFunction("list", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if size, ok := v.ToInt(); ok {
					return NewListFromIterable(iterator.Generate[Value, funcGen.Stack[Value]](size, func(i int) (Value, error) { return Int(i), nil })), nil
				}
				return nil, fmt.Errorf("list not alowed on %v", v)
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a list with n integer values, starting with 0.")).
		AddStaticFunction("goto", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if state, ok := v.ToInt(); ok {
					return createState(state), nil
				}
				return nil, errors.New("goto requires an int")
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a map with the key 'state' set to the given value.")).
		AddStaticFunction("sprintf", funcGen.Function[Value]{Func: sprintf, Args: -1, IsPure: true}.
			SetDescription("format", "args", "the classic, well known sprintf function")).
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

func minFunc(st funcGen.Stack[Value], cs []Value) (Value, error) {
	var m Value
	for i := 0; i < st.Size(); i++ {
		v := st.Get(i)
		if i == 0 {
			m = v
		} else {
			less, err := Less(st, v, m)
			if err != nil {
				return nil, err
			}
			if less {
				m = v
			}
		}
	}
	return m, nil
}

func maxFunc(st funcGen.Stack[Value], cs []Value) (Value, error) {
	var m Value
	for i := 0; i < st.Size(); i++ {
		v := st.Get(i)
		if i == 0 {
			m = v
		} else {
			less, err := Less(st, m, v)
			if err != nil {
				return nil, err
			}
			if less {
				m = v
			}
		}
	}
	return m, nil
}

func sprintf(st funcGen.Stack[Value], cs []Value) (Value, error) {
	switch st.Size() {
	case 0:
		return String(""), nil
	case 1:
		return String(fmt.Sprint(st.Get(0))), nil
	default:
		if s, ok := st.Get(0).(String); ok {
			values := make([]any, st.Size()-1)
			for i := 1; i < st.Size(); i++ {
				values[i-1] = st.Get(i)
			}
			return String(fmt.Sprintf(string(s), values...)), nil
		} else {
			return nil, fmt.Errorf("sprintf requires string as first argument")
		}
	}
}

func createLowPass(st funcGen.Stack[Value], store []Value) (Value, error) {
	var name string
	if n, ok := st.Get(0).(String); ok {
		name = string(n)
	} else {
		return nil, fmt.Errorf("createLowPass requires a string as first argument")
	}
	t, err := ToFunc("createLowPass", st, 1, 1)
	if err != nil {
		return nil, err
	}
	xf, err := ToFunc("createLowPass", st, 2, 1)
	if err != nil {
		return nil, err
	}
	tau, err := ToFloat("createLowPass", st, 3)
	if err != nil {
		return nil, err
	}
	lp := Closure(funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			p0 := st.Get(0)
			p1 := st.Get(1)
			ol, _ := st.Get(2).ToMap()
			yv, _ := ol.Get(name)
			t0, err := MustFloat(t.Eval(st, p0))
			if err != nil {
				return nil, err
			}
			t1, err := MustFloat(t.Eval(st, p1))
			if err != nil {
				return nil, err
			}
			x, err := MustFloat(xf.Eval(st, p1))
			if err != nil {
				return nil, err
			}
			y, err := MustFloat(yv, nil)
			dt := t1 - t0
			a := math.Exp(-dt / tau)
			yn := y*a + x*(1-a)
			m, _ := p1.ToMap()
			return NewMap(AppendMap{key: name, value: Float(yn), parent: m}), nil
		},
		Args:   3,
		IsPure: true,
	})
	in := Closure(funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			p0 := st.Get(0)
			x, err := xf.Eval(st, p0)
			if err != nil {
				return nil, err
			}
			m, _ := p0.ToMap()
			return NewMap(AppendMap{key: name, value: x, parent: m}), nil
		},
		Args:   1,
		IsPure: true,
	})
	return NewMap(listMap.New[Value](2).Append("filter", lp).Append("initial", in)), nil
}

func MustFloat(v Value, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if f, ok := v.ToFloat(); ok {
		return f, nil
	}
	return 0, fmt.Errorf("not a float: %v", v)
}
