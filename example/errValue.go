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

func (e ErrValue) String() (string, error) {
	va, err := value.Float(e.val).String()
	if err != nil {
		return "", err
	}
	er, err := value.Float(e.err).String()
	if err != nil {
		return "", err
	}
	return string(va + "±" + er), nil
}

func (e ErrValue) ToBool() (bool, bool) {
	return false, false
}

func (e ErrValue) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

var ErrValueMethods = value.MethodMap{
	"val": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		return value.Float(ev.val), nil
	}).
		SetMethodDescription("Returns the value of the error value"),
	"err": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		return value.Float(ev.err), nil
	}).
		SetMethodDescription("Returns the error of the error value"),
	"string": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		s, err := ev.String()
		return value.String(s), err
	}).
		SetMethodDescription("Returns the string representation of the error value"),
}

func (e ErrValue) GetMethod(name string) (funcGen.Function[value.Value], error) {
	return ErrValueMethods.Get(name)
}

func errOperation(name string,
	def func(a value.Value, b value.Value) (value.Value, error),
	f func(a, b ErrValue) (ErrValue, error)) func(a value.Value, b value.Value) (value.Value, error) {

	return func(a value.Value, b value.Value) (value.Value, error) {
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

func toErr(stack funcGen.Stack[value.Value], store []value.Value) (value.Value, error) {
	if err, ok := stack.Get(0).ToFloat(); ok {
		return ErrValue{err: math.Abs(err)}, nil
	}
	return nil, fmt.Errorf("err requires a float value")
}

var ErrValueParser = value.SetUpParser(value.New().
	AddOp("+", false, errOperation("+", value.Add,
		func(a, b ErrValue) (ErrValue, error) {
			return ErrValue{a.val + b.val, a.err + b.err}, nil
		}),
	).
	AddOp("-", false, errOperation("-", value.Sub,
		func(a, b ErrValue) (ErrValue, error) {
			return ErrValue{a.val - b.val, a.err + b.err}, nil
		}),
	).
	AddOp("*", true, errOperation("*", value.Mul,
		func(a, b ErrValue) (ErrValue, error) {
			return ErrValue{a.val * b.val, math.Abs(a.val*b.err) + math.Abs(b.val*a.err) + b.err*a.err}, nil
		}),
	).
	AddOp("/", true, errOperation("/", value.Div,
		func(a, b ErrValue) (ErrValue, error) {
			val := a.val / b.val
			return ErrValue{val, (math.Abs(a.val)+a.err)/(math.Abs(b.val)-b.err) - math.Abs(val)}, nil
		}),
	).
	AddOp("+-", false, func(a value.Value, b value.Value) (value.Value, error) {
		if v, ok := a.ToFloat(); ok {
			if e, ok := b.ToFloat(); ok {
				return ErrValue{v, math.Abs(e)}, nil
			}
		}
		return nil, fmt.Errorf("+- not allowed on %v+-%v", a, b)
	}).
	AddStaticFunction("err", funcGen.Function[value.Value]{
		Func:   toErr,
		Args:   1,
		IsPure: true,
	}.SetDescription("float", "Creates an error value with the given float as the error. The value is set to 0.")))
