package dynType

import (
	"fmt"
	"github.com/hneemann/parser2"
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

type vClosure parser2.Function[Value]

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

func (v vList) Map(c vClosure) Value {
	if c.Args != 1 {
		panic("map requires closure with one argument")
	}
	var m = make([]Value, len(v))
	for i, e := range v {
		var err error
		m[i], err = c.Func([]Value{e})
		if err != nil {
			panic(err)
		}
	}
	return vList(m)
}

func (v vList) Reduce(c vClosure) Value {
	if c.Args != 2 {
		panic("reduce requires closure with two arguments")
	}
	var red Value
	for i, e := range v {
		if i == 0 {
			red = e
		} else {
			var err error
			red, err = c.Func([]Value{red, e})
			if err != nil {
				panic(err)
			}
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

func vEqual(a, b Value) (Value, error) {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as == bs), nil
		}
	}
	if ab, oka := a.(vBool); oka {
		if bb, okb := b.(vBool); okb {
			return vBool(ab == bb), nil
		}
	}
	return vBool(a.Float() == b.Float()), nil
}

func vLess(a, b Value) (Value, error) {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as < bs), nil
		}
	}
	return vBool(a.Float() < b.Float()), nil
}

func vLessEqual(a, b Value) (Value, error) {
	if as, oka := a.(vString); oka {
		if bs, okb := b.(vString); okb {
			return vBool(as <= bs), nil
		}
	}
	return vBool(a.Float() <= b.Float()), nil
}

func swap(f func(a, b Value) (Value, error)) func(a, b Value) (Value, error) {
	return func(a, b Value) (Value, error) {
		return f(b, a)
	}
}

func vAdd(a, b Value) (Value, error) {
	if af, oka := a.(vString); oka {
		if bf, okb := b.(vString); okb {
			return af + bf, nil
		}
	}
	return vFloat(a.Float() + b.Float()), nil
}

type typeHandler struct{}

func (th typeHandler) ParseNumber(s string) (Value, error) {
	f, err := strconv.ParseFloat(s, 64)
	return vFloat(f), err
}

func (th typeHandler) FromClosure(closure parser2.Function[Value]) Value {
	return vClosure(closure)
}

func (th typeHandler) ToClosure(fu Value) (parser2.Function[Value], bool) {
	cl, ok := fu.(vClosure)
	return parser2.Function[Value](cl), ok
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

func (th typeHandler) Generate(ast parser2.AST, g *parser2.FunctionGenerator[Value]) (parser2.Func[Value], error) {
	var zero Value
	if op, ok := ast.(*parser2.Operate); ok {
		// AND and OR with short evaluation
		switch op.Operator {
		case "&":
			aFunc, err := g.GenerateFunc(op.A)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B)
			if err != nil {
				return nil, err
			}
			return func(v parser2.Variables[Value]) (Value, error) {
				value, err := aFunc(v)
				if err != nil {
					return zero, err
				}
				if !value.Bool() {
					return vBool(false), nil
				} else {
					v2, err := bFunc(v)
					if err != nil {
						return zero, err
					}
					return vBool(v2.Bool()), nil
				}
			}, nil
		case "|":
			aFunc, err := g.GenerateFunc(op.A)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B)
			if err != nil {
				return nil, err
			}
			return func(v parser2.Variables[Value]) (Value, error) {
				value, err2 := aFunc(v)
				if err2 != nil {
					return zero, err
				}
				if value.Bool() {
					return vBool(true), nil
				} else {
					v2, err2 := bFunc(v)
					if err2 != nil {
						return zero, err
					}
					return vBool(v2.Bool()), nil
				}
			}, nil
		}
	}
	return nil, nil
}

var th typeHandler

var DynType = parser2.New[Value]().
	AddOp("|", true, func(a, b Value) (Value, error) { return vBool(a.Bool() || b.Bool()), nil }).
	AddOp("&", true, func(a, b Value) (Value, error) { return vBool(a.Bool() && b.Bool()), nil }).
	AddOp("=", true, vEqual).
	AddOp("!=", true, func(a, b Value) (Value, error) {
		equal, err := vEqual(a, b)
		if err != nil {
			return nil, err
		}
		return !equal.(vBool), nil
	}).
	AddOp("<", false, vLess).
	AddOp(">", false, swap(vLess)).
	AddOp("<=", false, vLessEqual).
	AddOp(">=", false, swap(vLessEqual)).
	AddOp("+", false, vAdd). // vAdd is not commutative, since strings can be added
	AddOp("-", false, func(a, b Value) (Value, error) { return vFloat(a.Float() - b.Float()), nil }).
	AddOp("*", true, func(a, b Value) (Value, error) { return vFloat(a.Float() * b.Float()), nil }).
	AddOp("/", false, func(a, b Value) (Value, error) { return vFloat(a.Float() / b.Float()), nil }).
	AddUnary("-", func(a Value) (Value, error) { return vFloat(-a.Float()), nil }).
	SetListHandler(th).
	SetMapHandler(th).
	SetStringConverter(th).
	SetClosureHandler(th).
	SetNumberParser(th).
	SetCustomGenerator(th).
	SetToBool(func(c Value) bool { return c.Bool() }).
	AddConstant("pi", vFloat(math.Pi)).
	AddConstant("true", vBool(true)).
	AddConstant("false", vBool(false)).
	AddStaticFunction("abs", parser2.Function[Value]{
		Func:   func(v []Value) (Value, error) { return vFloat(math.Abs(v[0].Float())), nil },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("sqrt", parser2.Function[Value]{
		Func:   func(v []Value) (Value, error) { return vFloat(math.Sqrt(v[0].Float())), nil },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("ln", parser2.Function[Value]{
		Func:   func(v []Value) (Value, error) { return vFloat(math.Log(v[0].Float())), nil },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("sprintf", parser2.Function[Value]{
		Func:   sprintf,
		Args:   -1,
		IsPure: true,
	})

func sprintf(a []Value) (Value, error) {
	switch len(a) {
	case 0:
		return vString(""), nil
	case 1:
		return vString(fmt.Sprint(a[0])), nil
	default:
		if s, ok := a[0].(vString); ok {
			values := make([]any, len(a)-1)
			for i, v := range a[1:] {
				values[i] = v
			}
			return vString(fmt.Sprintf(string(s), values...)), nil
		} else {
			return nil, fmt.Errorf("sprintf requires string as first argument")
		}
	}
}
