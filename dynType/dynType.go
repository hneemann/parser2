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

type vClosure parser2.Closure[Value]

func (v vClosure) Float() float64 {
	return 0
}

func (v vClosure) Bool() bool {
	return false
}

func (v vClosure) CreateFunction() parser2.Function[Value] {
	c := parser2.Closure[Value](v)
	return c.Impl
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
	f := c.CreateFunction()
	if f.Args != 1 {
		panic("map requires closure with one argument")
	}
	var m = make([]Value, len(v))
	for i, e := range v {
		m[i] = f.Func([]Value{e})
	}
	return vList(m)
}

type vMap map[string]Value

func (v vMap) Float() float64 {
	return 0
}

func (v vMap) Bool() bool {
	return len(v) > 0
}

func vNeg(a Value) Value {
	return vFloat(-a.Float())
}

func vOr(a, b Value) Value {
	return vBool(a.Bool() || b.Bool())
}

func vAnd(a, b Value) Value {
	return vBool(a.Bool() && b.Bool())
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

func vNotEqual(a, b Value) Value {
	return !vEqual(a, b).(vBool)
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

func vSub(a, b Value) Value {
	return vFloat(a.Float() - b.Float())
}

func vMul(a, b Value) Value {
	return vFloat(a.Float() * b.Float())
}

func vDiv(a, b Value) Value {
	return vFloat(a.Float() / b.Float())
}

type typeHandler struct{}

func (th typeHandler) ParseNumber(s string) (Value, error) {
	f, err := strconv.ParseFloat(s, 64)
	return vFloat(f), err
}

func (th typeHandler) FromClosure(closure parser2.Closure[Value]) Value {
	return vClosure(closure)
}

func (th typeHandler) ToClosure(fu Value) (parser2.Closure[Value], bool) {
	cl, ok := fu.(vClosure)
	return parser2.Closure[Value](cl), ok
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

func (th typeHandler) Generate(ast parser2.AST, g *parser2.FunctionGenerator[Value]) parser2.Func[Value] {
	// ite without evaluation of not required expression
	if fc, ok := ast.(*parser2.FunctionCall); ok && len(fc.Args) == 3 {
		if id, ok := fc.Func.(parser2.Ident); ok {
			if id == "ite" {
				condFunc := g.GenerateFunc(fc.Args[0])
				thenFunc := g.GenerateFunc(fc.Args[1])
				elseFunc := g.GenerateFunc(fc.Args[2])
				return func(v parser2.Variables[Value]) Value {
					if condFunc(v).Bool() {
						return thenFunc(v)
					} else {
						return elseFunc(v)
					}
				}
			}
		}
	}
	if op, ok := ast.(*parser2.Operate); ok {
		// AND and OR with short evaluation
		switch op.Operator {
		case "&":
			aFunc := g.GenerateFunc(op.A)
			bFunc := g.GenerateFunc(op.B)
			return func(v parser2.Variables[Value]) Value {
				if !aFunc(v).Bool() {
					return vBool(false)
				} else {
					return vBool(bFunc(v).Bool())
				}
			}
		case "|":
			aFunc := g.GenerateFunc(op.A)
			bFunc := g.GenerateFunc(op.B)
			return func(v parser2.Variables[Value]) Value {
				if aFunc(v).Bool() {
					return vBool(true)
				} else {
					return vBool(bFunc(v).Bool())
				}
			}
		}
	}
	return nil
}

var th typeHandler

var DynType = parser2.New[Value]().
	AddOp("|", vOr).
	AddOp("&", vAnd).
	AddOp("=", vEqual).
	AddOp("!=", vNotEqual).
	AddOp("<", vLess).
	AddOp(">", swap(vLess)).
	AddOp("<=", vLessEqual).
	AddOp(">=", swap(vLessEqual)).
	AddOp("+", vAdd).
	AddOp("-", vSub).
	AddOp("*", vMul).
	AddOp("/", vDiv).
	AddUnary("-", vNeg).
	SetListHandler(th).
	SetMapHandler(th).
	SetStringHandler(th).
	SetClosureHandler(th).
	SetNumberParser(th).
	SetCustomGenerator(th).
	AddConstant("pi", vFloat(math.Pi)).
	AddConstant("true", vBool(true)).
	AddConstant("false", vBool(false)).
	AddStaticFunction("sqrt", parser2.Function[Value]{
		Func:   func(v []Value) Value { return vFloat(math.Sqrt(v[0].Float())) },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("ln", parser2.Function[Value]{
		Func:   func(v []Value) Value { return vFloat(math.Log(v[0].Float())) },
		Args:   1,
		IsPure: true,
	}).
	AddStaticFunction("sprintf", parser2.Function[Value]{
		Func:   sprintf,
		Args:   -1,
		IsPure: true,
	}).
	AddStaticFunction("lowPass", parser2.Function[Value]{
		Func:   lowPass,
		Args:   1,
		IsPure: false,
	})

func sprintf(a []Value) Value {
	switch len(a) {
	case 0:
		return vString("")
	case 1:
		return vString(fmt.Sprint(a[0]))
	default:
		if s, ok := a[0].(vString); ok {
			values := make([]any, len(a)-1)
			for i, v := range a[1:] {
				values[i] = v
			}
			return vString(fmt.Sprintf(string(s), values...))
		} else {
			panic("sprintf requires string as first argument")
		}
	}
}

func lowPass(a []Value) Value {
	tau := a[0].Float()
	init := false
	lt := 0.0
	lx := 0.0
	return vClosure{
		Impl: parser2.Function[Value]{
			Func: func(args []Value) Value {
				t := args[0].Float()
				x := args[1].Float()
				if !init {
					lt = t
					lx = x
					init = true
				} else {
					dt := t - lt
					a := math.Exp(-dt / tau)
					lx = lx*a + x*(1-a)
					lt = t
				}
				return vFloat(lx)
			},
			Args:   2,
			IsPure: false,
		},
	}
}
