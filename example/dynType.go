package example

import (
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"math"
	"strconv"
)

type Value interface {
	Float() float64
	Bool() bool
}

type vFloat float64

func (v vFloat) Float() float64 {
	return float64(v)
}

func (v vFloat) Bool() bool {
	return v != 0
}

type vString string

func (v vString) Float() float64 {
	return 0
}

func (v vString) Bool() bool {
	return len(v) > 0
}

type vBool bool

func (v vBool) Float() float64 {
	if v {
		return 1
	}
	return 0
}

func (v vBool) Bool() bool {
	return bool(v)
}

type vClosure funcGen.Function[Value]

func (v vClosure) Float() float64 {
	return 0
}

func (v vClosure) Bool() bool {
	return false
}

type vList []Value

func (v vList) Float() float64 {
	return 0
}

func (v vList) Bool() bool {
	return len(v) > 0
}

func (v vList) Size() Value {
	return vFloat(len(v))
}

func (v vList) Map(val Value) Value {
	c, ok := val.(vClosure)
	if !ok {
		panic("argument of map needs to be a closure")
	}
	if c.Args != 1 {
		panic("map requires closure with one argument")
	}
	st := funcGen.NewStack[Value](make([]Value, 0, 10))
	var m = make([]Value, len(v))
	for i, e := range v {
		st.Push(e)
		m[i] = c.Func(st, nil)
		st.Remove(1)
	}
	return vList(m)
}

func (v vList) Reduce(val Value) Value {
	c, ok := val.(vClosure)
	if !ok {
		panic("argument of map needs to be a closure")
	}
	if c.Args != 2 {
		panic("reduce requires closure with two arguments")
	}
	var red Value
	st := funcGen.NewStack[Value](make([]Value, 0, 10))
	for i, e := range v {
		if i == 0 {
			red = e
		} else {
			st.Push(red)
			st.Push(e)
			red = c.Func(st, nil)
			st.Remove(2)
		}
	}
	return red
}

type vMap map[string]Value

func (v vMap) Float() float64 {
	return 0
}

func (v vMap) Bool() bool {
	return len(v) > 0
}

func vEqual(a, b Value) Value {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as == bs)
		}
	}
	if ab, oka := a.(vBool); oka {
		if bb, okb := b.(vBool); okb {
			return vBool(ab == bb)
		}
	}
	return vBool(a.Float() == b.Float())
}

func vLess(a, b Value) Value {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as < bs)
		}
	}
	return vBool(a.Float() < b.Float())
}

func vLessEqual(a, b Value) Value {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as <= bs)
		}
	}
	return vBool(a.Float() <= b.Float())
}

func swap(f func(a, b Value) Value) func(a, b Value) Value {
	return func(a, b Value) Value {
		return f(b, a)
	}
}

func vAdd(a, b Value) Value {
	if af, oka := a.(vString); oka {
		if bf, okb := b.(vString); okb {
			return af + bf
		}
	}
	return vFloat(a.Float() + b.Float())
}

type typeHandler struct{}

func (th typeHandler) ParseNumber(s string) (Value, error) {
	f, err := strconv.ParseFloat(s, 64)
	return vFloat(f), err
}

func (th typeHandler) FromClosure(closure funcGen.Function[Value]) Value {
	return vClosure(closure)
}

func (th typeHandler) ToClosure(fu Value) (funcGen.Function[Value], bool) {
	cl, ok := fu.(vClosure)
	return funcGen.Function[Value](cl), ok
}

func (th typeHandler) FromList(items []Value) Value {
	return vList(items)
}

func (th typeHandler) AccessList(list Value, index Value) (Value, error) {
	li, ok := list.(vList)
	if ok {
		i := int(index.Float())
		if i < 0 || i >= len(li) {
			return nil, fmt.Errorf("list index out of bounds: %v", i)
		}
		return li[i], nil
	} else {
		return nil, fmt.Errorf("not a list: %v", list)
	}
}

func (th typeHandler) FromMap(items map[string]Value) Value {
	return vMap(items)
}

func (th typeHandler) IsMap(m Value) bool {
	_, ok := m.(vMap)
	return ok
}

func (th typeHandler) AccessMap(m Value, key string) (Value, error) {
	ma, ok := m.(vMap)
	if ok {
		if v, ok := ma[key]; ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("map does not contain %v", key)
		}
	} else {
		return nil, fmt.Errorf("not a map: %v", m)
	}
}

func (th typeHandler) FromString(s string) Value {
	return vString(s)
}

func (th typeHandler) Generate(ast parser2.AST, am, cm funcGen.ArgsMap, g *funcGen.FunctionGenerator[Value]) (funcGen.Func[Value], error) {
	if op, ok := ast.(*parser2.Operate); ok {
		// AND and OR with short evaluation
		switch op.Operator {
		case "&":
			aFunc, err := g.GenerateFunc(op.A, am, cm)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) Value {
				if !aFunc(st, cs).Bool() {
					return vBool(false)
				} else {
					return vBool(bFunc(st, cs).Bool())
				}
			}, nil
		case "|":
			aFunc, err := g.GenerateFunc(op.A, am, cm)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) Value {
				if aFunc(st, cs).Bool() {
					return vBool(true)
				} else {
					return vBool(bFunc(st, cs).Bool())
				}
			}, nil
		}
	}
	return nil, nil
}

var th typeHandler

var DynType = funcGen.New[Value]().
	AddOp("|", true, func(a, b Value) Value { return vBool(a.Bool() || b.Bool()) }).
	AddOp("&", true, func(a, b Value) Value { return vBool(a.Bool() && b.Bool()) }).
	AddOp("=", true, vEqual).
	AddOp("!=", true, func(a, b Value) Value { return !vEqual(a, b).(vBool) }).
	AddOp("<", false, vLess).
	AddOp(">", false, swap(vLess)).
	AddOp("<=", false, vLessEqual).
	AddOp(">=", false, swap(vLessEqual)).
	AddOp("+", false, vAdd). // vAdd is not commutative, since strings can be added
	AddOp("-", false, func(a, b Value) Value { return vFloat(a.Float() - b.Float()) }).
	AddOp("*", true, func(a, b Value) Value { return vFloat(a.Float() * b.Float()) }).
	AddOp("/", false, func(a, b Value) Value { return vFloat(a.Float() / b.Float()) }).
	AddUnary("-", func(a Value) Value { return vFloat(-a.Float()) }).
	SetListHandler(th).
	SetMapHandler(th).
	SetStringConverter(th).
	SetClosureHandler(th).
	SetNumberParser(th).
	SetCustomGenerator(th).
	SetToBool(func(c Value) bool { return c.Bool() }).
	SetIsEqual(func(a, b Value) bool { return vEqual(a, b).Bool() }).
	AddConstant("pi", vFloat(math.Pi)).
	AddConstant("true", vBool(true)).
	AddConstant("false", vBool(false)).
	AddStaticFunction("abs", funcGen.Function[Value]{
		Func:   func(st funcGen.Stack[Value], cs []Value) Value { return vFloat(math.Abs(st.Get(0).Float())) },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("sqrt", funcGen.Function[Value]{
		Func:   func(st funcGen.Stack[Value], cs []Value) Value { return vFloat(math.Sqrt(st.Get(0).Float())) },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("ln", funcGen.Function[Value]{
		Func:   func(st funcGen.Stack[Value], cs []Value) Value { return vFloat(math.Log(st.Get(0).Float())) },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("sprintf", funcGen.Function[Value]{
		Func:   sprintf,
		Args:   -1,
		IsPure: true,
	})

func sprintf(st funcGen.Stack[Value], cs []Value) Value {
	switch st.Size() {
	case 0:
		return vString("")
	case 1:
		return vString(fmt.Sprint(st.Get(0)))
	default:
		if s, ok := st.Get(0).(vString); ok {
			values := make([]any, st.Size()-1)
			for i := 1; i < st.Size(); i++ {
				values[i-1] = st.Get(i)
			}
			return vString(fmt.Sprintf(string(s), values...))
		} else {
			panic("sprintf requires string as first argument")
		}
	}
}
