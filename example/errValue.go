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

func (e ErrValue) ToString(st funcGen.Stack[value.Value]) (string, error) {
	va, err := value.Float(e.val).ToString(st)
	if err != nil {
		return "", err
	}
	er, err := value.Float(e.err).ToString(st)
	if err != nil {
		return "", err
	}
	return va + "±" + er, nil
}

func (e ErrValue) ToBool() (bool, bool) {
	return false, false
}

func (e ErrValue) GetMin() float64 {
	return e.val - e.err
}

func (e ErrValue) GetMax() float64 {
	return e.val + e.err
}

func (e ErrValue) Matches(b ErrValue) bool {
	return e.GetMin() <= b.GetMax() && e.GetMax() >= b.GetMin()
}

func (e ErrValue) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func createErrValueMethods() value.MethodMap {
	return value.MethodMap{
		"val": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(ev.val), nil
		}).
			SetMethodDescription("Returns the value of the error value."),
		"err": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(ev.err), nil
		}).
			SetMethodDescription("Returns the error of the error value."),
		"string": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			s, err := ev.ToString(stack)
			return value.String(s), err
		}).
			SetMethodDescription("Returns the string representation of the error value."),
	}
}

var errValType value.Type

func (e ErrValue) GetType() value.Type {
	return errValType
}

func Equal(st funcGen.Stack[value.Value], aVal, bVal value.Value) (bool, error) {
	if a, ok := aVal.(ErrValue); ok {
		if b, ok := bVal.(ErrValue); ok {
			return a.Matches(b), nil
		} else {
			if bf, ok := bVal.ToFloat(); ok {
				return a.Matches(ErrValue{val: bf}), nil
			} else {
				return false, fmt.Errorf("= not alowed on %s=%s", value.TypeName(a), value.TypeName(b))
			}
		}
	} else {
		if b, ok := bVal.(ErrValue); ok {
			if af, ok := aVal.ToFloat(); ok {
				return b.Matches(ErrValue{val: af}), nil
			} else {
				return false, fmt.Errorf("= not alowed on %s=%s", value.TypeName(a), value.TypeName(b))
			}
		}
	}
	return value.Equal(st, aVal, bVal)
}

func errOperation(name string,
	def func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error),
	f func(a, b ErrValue) (value.Value, error)) func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {

	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(ErrValue); ok {
			if be, ok := b.(ErrValue); ok {
				// both are error values
				return f(ae, be)
			} else {
				// a is error value, b is'nt
				if bf, ok := b.ToFloat(); ok {
					return f(ae, ErrValue{val: bf})
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			}
		} else {
			if be, ok := b.(ErrValue); ok {
				// b is error value, a is'nt
				if af, ok := a.ToFloat(); ok {
					return f(ErrValue{val: af}, be)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			} else {
				// no error value at all
				return def(st, a, b)
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

var ErrValueParser = value.New().
	Modify(func(f *value.FunctionGenerator) {
		errValType = f.RegisterType("errValue")
		f.AddOpBehind(">", ">>>", false, errOperation(">>>", f.GetOpImpl(">"),
			func(a, b ErrValue) (value.Value, error) {
				return value.Bool(a.GetMin() > b.GetMax()), nil
			}), true,
		)
		f.AddOpBehind("<", "<<<", false, errOperation("<<<", f.GetOpImpl("<"),
			func(a, b ErrValue) (value.Value, error) {
				return value.Bool(a.GetMax() < b.GetMin()), nil
			}), true,
		)
	}).
	RegisterMethods(errValType, createErrValueMethods()).
	SetEqual(Equal).
	ReplaceOp("+", false, true, func(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
		return errOperation("+", old,
			func(a, b ErrValue) (value.Value, error) {
				return ErrValue{a.val + b.val, a.err + b.err}, nil
			})
	}).
	ReplaceOp("-", false, true, func(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
		return errOperation("-", old,
			func(a, b ErrValue) (value.Value, error) {
				return ErrValue{a.val - b.val, a.err + b.err}, nil
			})
	}).
	ReplaceOp("*", true, true, func(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
		return errOperation("*", old,
			func(a, b ErrValue) (value.Value, error) {
				return ErrValue{a.val * b.val, math.Abs(a.val*b.err) + math.Abs(b.val*a.err) + b.err*a.err}, nil
			})
	}).
	ReplaceOp("/", true, true, func(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
		return errOperation("/", old,
			func(a, b ErrValue) (value.Value, error) {
				val := a.val / b.val
				return ErrValue{val, (math.Abs(a.val)+a.err)/(math.Abs(b.val)-b.err) - math.Abs(val)}, nil
			})
	}).
	ReplaceUnary("-", func(orig funcGen.UnaryOperatorImpl[value.Value]) funcGen.UnaryOperatorImpl[value.Value] {
		return func(a value.Value) (value.Value, error) {
			if v, ok := a.(ErrValue); ok {
				return ErrValue{val: -v.val, err: v.err}, nil
			}
			return orig(a)
		}
	}).
	AddOp("+-", false, func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
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
	}.SetDescription("float", "Creates an error value with the given float as the error. The value is set to 0."))
