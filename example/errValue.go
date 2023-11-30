package example

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"math"
)

type ErrValue struct {
	val float64
	err float64
}

func (e ErrValue) ToList() (*value.List, bool) {
	return nil, false
}

func (e ErrValue) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (e ErrValue) ToInt() (int, bool) {
	return int(e.val), true
}

func (e ErrValue) ToFloat() (float64, bool) {
	return e.val, true
}

func (e ErrValue) String() string {
	return value.Float(e.val).String() + "±" + value.Float(e.err).String()
}

func (e ErrValue) ToBool() (bool, bool) {
	return false, false
}

func (e ErrValue) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

var ErrValueMethods = value.MethodMap{
	"val":    value.MethodAtType(1, func(ev ErrValue, stack funcGen.Stack[value.Value]) value.Value { return value.Float(ev.val) }),
	"err":    value.MethodAtType(1, func(ev ErrValue, stack funcGen.Stack[value.Value]) value.Value { return value.Float(ev.err) }),
	"string": value.MethodAtType(1, func(ev ErrValue, stack funcGen.Stack[value.Value]) value.Value { return value.String(ev.String()) }),
}

func (e ErrValue) GetMethod(name string) (funcGen.Function[value.Value], error) {
	return ErrValueMethods.Get(name)
}

func errOperation(name string,
	def func(a value.Value, b value.Value) value.Value,
	f func(a, b ErrValue) ErrValue) func(a value.Value, b value.Value) value.Value {

	return func(a value.Value, b value.Value) value.Value {
		if ae, ok := a.(ErrValue); ok {
			if be, ok := b.(ErrValue); ok {
				// both are error values
				return f(ae, be)
			} else {
				// a is error value, b is'nt
				if bf, ok := b.ToFloat(); ok {
					return f(ae, ErrValue{val: bf})
				} else {
					panic(fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b))
				}
			}
		} else {
			if be, ok := b.(ErrValue); ok {
				// b is error value, a is'nt
				if af, ok := a.ToFloat(); ok {
					return f(ErrValue{val: af}, be)
				} else {
					panic(fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b))
				}
			} else {
				// no error value at all
				return def(a, b)
			}
		}
	}
}

func toErr(stack funcGen.Stack[value.Value], store []value.Value) value.Value {
	if err, ok := stack.Get(0).ToFloat(); ok {
		return ErrValue{err: math.Abs(err)}
	}
	panic("err requires a float value")
}

var ErrValueParser = value.SetUpParser(value.New().
	AddOp("+", false, errOperation("+", value.Add,
		func(a, b ErrValue) ErrValue {
			return ErrValue{a.val + b.val, a.err + b.err}
		}),
	).
	AddOp("-", false, errOperation("-", value.Sub,
		func(a, b ErrValue) ErrValue {
			return ErrValue{a.val - b.val, a.err + b.err}
		}),
	).
	AddOp("*", true, errOperation("*", value.Mul,
		func(a, b ErrValue) ErrValue {
			return ErrValue{a.val * b.val, math.Abs(a.val*b.err) + math.Abs(b.val*a.err) + b.err*a.err}
		}),
	).
	AddOp("/", true, errOperation("/", value.Div,
		func(a, b ErrValue) ErrValue {
			val := a.val / b.val
			return ErrValue{val, (math.Abs(a.val)+a.err)/(math.Abs(b.val)-b.err) - math.Abs(val)}
		}),
	).
	AddOp("+-", false, func(a value.Value, b value.Value) value.Value {
		if v, ok := a.ToFloat(); ok {
			if e, ok := b.ToFloat(); ok {
				return ErrValue{v, math.Abs(e)}
			}
		}
		panic(fmt.Errorf("+- not allowed on %v/%v", a, b))
	}).
	AddStaticFunction("err", funcGen.Function[value.Value]{
		Func:   toErr,
		Args:   1,
		IsPure: true,
	}))
