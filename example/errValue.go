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
			SetMethodDescription("Returns the value of the error value"),
		"err": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(ev.err), nil
		}).
			SetMethodDescription("Returns the error of the error value"),
		"string": value.MethodAtType(0, func(ev ErrValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			s, err := ev.ToString(stack)
			return value.String(s), err
		}).
			SetMethodDescription("Returns the string representation of the error value"),
	}
}

var errValType value.Type

func (e ErrValue) GetType() value.Type {
	return errValType
}

var ErrValueParser = value.New().
	Modify(func(f *value.FunctionGenerator) {
		errValType = f.RegisterType()
		fromInt := value.UpCast{
			From: value.IntTypeId,
			To:   errValType,
			Cast: func(a value.Value) (value.Value, error) {
				if aa, ok := a.(value.Int); ok {
					return ErrValue{val: float64(aa)}, nil
				}
				return nil, fmt.Errorf("cannot convert %s to ErrValue", value.TypeName(a))
			},
		}
		fromFloat := value.UpCast{
			From: value.FloatTypeId,
			To:   errValType,
			Cast: func(a value.Value) (value.Value, error) {
				if aa, ok := a.(value.Float); ok {
					return ErrValue{val: float64(aa)}, nil
				}
				return nil, fmt.Errorf("cannot convert %s to ErrValue", value.TypeName(a))
			},
		}

		f.AddTypedOpFunc("+", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return ErrValue{aa.val + bb.val, aa.err + bb.err}, nil
				}
			}
			return nil, nil
		}).AddUpCast("+", fromInt, fromFloat)

		f.AddTypedOpFunc("-", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return ErrValue{aa.val - (bb.val), aa.err + bb.err}, nil
				}
			}
			return nil, nil
		}).AddUpCast("-", fromInt, fromFloat)

		f.AddTypedOpFunc("*", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return ErrValue{aa.val * bb.val, math.Abs(aa.val*bb.err) + math.Abs(bb.val*aa.err) + bb.err*aa.err}, nil
				}
			}
			return nil, nil
		}).AddUpCast("*", fromInt, fromFloat)

		f.AddTypedOpFunc("/", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					val := aa.val / bb.val
					return ErrValue{val, (math.Abs(aa.val)+aa.err)/(math.Abs(bb.val)-bb.err) - math.Abs(val)}, nil
				}
			}
			return nil, nil
		}).AddUpCast("/", fromInt, fromFloat)

		f.AddTypedOpFunc("=", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return value.Bool(aa.Matches(bb)), nil
				}
			}
			return nil, nil
		}).AddUpCast("=", fromInt, fromFloat)

		f.AddTypedOpFunc("<", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return value.Bool(aa.val < bb.val), nil
				}
			}
			return nil, nil
		}).AddUpCast("<", fromInt, fromFloat)

		f.AddOpBehind(">", ">>>", false, f.CreateOpTable(">>>"), true)
		f.AddTypedOpFunc(">>>", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return value.Bool(aa.GetMin() > bb.GetMax()), nil
				}
			}
			return nil, fmt.Errorf(">>> not allowed on %s>>>%s", value.TypeName(a), value.TypeName(b))
		}).AddUpCast(">>>", fromInt, fromFloat)

		f.AddOpBehind("<", "<<<", false, f.CreateOpTable("<<<"), true)
		f.AddTypedOpFunc("<<<", errValType, errValType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
			if aa, ok := a.(ErrValue); ok {
				if bb, ok := b.(ErrValue); ok {
					return value.Bool(aa.GetMax() < bb.GetMin()), nil
				}
			}
			return nil, fmt.Errorf("<<< not allowed on %s>>>%s", value.TypeName(a), value.TypeName(b))
		}).AddUpCast("<<<", fromInt, fromFloat)
	}).
	RegisterMethods(errValType, createErrValueMethods()).
	AddOp("+-", false, func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if v, ok := a.ToFloat(); ok {
			if e, ok := b.ToFloat(); ok {
				return ErrValue{v, math.Abs(e)}, nil
			}
		}
		return nil, fmt.Errorf("+- not allowed on %v+-%v", a, b)
	}).
	AddStaticFunction("err", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], store []value.Value) (value.Value, error) {
			if err, ok := stack.Get(0).ToFloat(); ok {
				return ErrValue{err: math.Abs(err)}, nil
			}
			return nil, fmt.Errorf("err requires a float value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("float", "Creates an error value with the given float as the error. The value is set to 0."))
