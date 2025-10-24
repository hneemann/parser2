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

func (e ErrValue) MatchesFloat(b float64) bool {
	return e.GetMin() <= b && e.GetMax() >= b
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

func toErr(stack funcGen.Stack[value.Value], store []value.Value) (value.Value, error) {
	if err, ok := stack.Get(0).ToFloat(); ok {
		return ErrValue{err: math.Abs(err)}, nil
	}
	return nil, fmt.Errorf("err requires a float value")
}

var ErrValueParser = value.New().
	Modify(func(f *value.FunctionGenerator) {
		errValType = f.RegisterType("errValue")
		addAdd(f)
		addSub(f)
		addMul(f)
		addDiv(f)
		addEqual(f)
		addLess(f)
	}).
	RegisterMethods(errValType, createErrValueMethods()).
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

func addAdd(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("+")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		bv := b.(ErrValue)
		return ErrValue{val: av.val + bv.val, err: av.err + bv.err}, nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		return ErrValue{val: av.val + float64(b.(value.Float)), err: av.err}, nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		bv := b.(ErrValue)
		return ErrValue{val: bv.val + float64(a.(value.Float)), err: bv.err}, nil
	})
	m.Register(errValType, value.IntTypeId, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		return ErrValue{val: av.val + float64(b.(value.Int)), err: av.err}, nil
	})
	m.Register(value.IntTypeId, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		bv := b.(ErrValue)
		return ErrValue{val: bv.val + float64(a.(value.Int)), err: bv.err}, nil
	})
}

func addSub(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("-")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		bv := b.(ErrValue)
		return ErrValue{val: av.val - bv.val, err: av.err + bv.err}, nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		return ErrValue{val: av.val - float64(b.(value.Float)), err: av.err}, nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		bv := b.(ErrValue)
		return ErrValue{val: float64(a.(value.Float)) - bv.val, err: bv.err}, nil
	})
	m.Register(errValType, value.IntTypeId, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		av := a.(ErrValue)
		return ErrValue{val: av.val - float64(b.(value.Int)), err: av.err}, nil
	})
	m.Register(value.IntTypeId, errValType, func(_ funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		bv := b.(ErrValue)
		return ErrValue{val: float64(a.(value.Int)) - bv.val, err: bv.err}, nil
	})
}

func addMul(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("*")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := bv.(ErrValue)
		return ErrValue{a.val * b.val, math.Abs(a.val*b.err) + math.Abs(b.val*a.err) + b.err*a.err}, nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		return ErrValue{a.val * float64(bv.(value.Float)), math.Abs(float64(bv.(value.Float)) * a.err)}, nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		b := bv.(ErrValue)
		return ErrValue{float64(av.(value.Float)) * b.val, math.Abs(float64(av.(value.Float)) * b.err)}, nil
	})
	m.Register(errValType, value.IntTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		return ErrValue{a.val * float64(bv.(value.Int)), math.Abs(float64(bv.(value.Int)) * a.err)}, nil
	})
	m.Register(value.IntTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		b := bv.(ErrValue)
		return ErrValue{float64(av.(value.Int)) * b.val, math.Abs(float64(av.(value.Int)) * b.err)}, nil
	})
}

func addDiv(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("/")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := bv.(ErrValue)
		val := a.val / b.val
		return ErrValue{val, (math.Abs(a.val)+a.err)/(math.Abs(b.val)-b.err) - math.Abs(val)}, nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		val := a.val / float64(bv.(value.Float))
		return ErrValue{val, (math.Abs(a.val)+a.err)/math.Abs(float64(bv.(value.Float))) - math.Abs(val)}, nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		b := bv.(ErrValue)
		val := float64(av.(value.Float)) / b.val
		return ErrValue{val, math.Abs(float64(av.(value.Float)))/(math.Abs(b.val)-b.err) - math.Abs(val)}, nil
	})
	m.Register(errValType, value.IntTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		val := a.val / float64(bv.(value.Int))
		return ErrValue{val, (math.Abs(a.val)+a.err)/math.Abs(float64(bv.(value.Int))) - math.Abs(val)}, nil
	})
	m.Register(value.IntTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		b := bv.(ErrValue)
		val := float64(av.(value.Int)) / b.val
		return ErrValue{val, math.Abs(float64(av.(value.Int)))/(math.Abs(b.val)-b.err) - math.Abs(val)}, nil
	})
}

func addEqual(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("=")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := bv.(ErrValue)
		return value.Bool(a.Matches(b)), nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := float64(bv.(value.Float))
		return value.Bool(a.MatchesFloat(b)), nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := float64(av.(value.Float))
		b := bv.(ErrValue)
		return value.Bool(b.MatchesFloat(a)), nil
	})
}

func addLess(f *value.FunctionGenerator) {
	m := f.GetOpMatrix("<")
	m.Register(errValType, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := bv.(ErrValue)
		return value.Bool(a.GetMax() < b.GetMin()), nil
	})
	m.Register(errValType, value.FloatTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := float64(bv.(value.Float))
		return value.Bool(a.GetMax() < b), nil
	})
	m.Register(value.FloatTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := float64(av.(value.Float))
		b := bv.(ErrValue)
		return value.Bool(a < b.GetMin()), nil
	})
	m.Register(errValType, value.IntTypeId, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := av.(ErrValue)
		b := float64(bv.(value.Int))
		return value.Bool(a.GetMax() < b), nil
	})
	m.Register(value.IntTypeId, errValType, func(_ funcGen.Stack[value.Value], av, bv value.Value) (value.Value, error) {
		a := float64(av.(value.Int))
		b := bv.(ErrValue)
		return value.Bool(a < b.GetMin()), nil
	})
}
