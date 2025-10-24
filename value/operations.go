package value

import (
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"math"
)

// Equal does not cover lists and maps
func Equal(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "=")
	m.Register(BoolTypeId, BoolTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Bool) == b.(Bool)), nil
	})
	m.Register(IntTypeId, IntTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Int) == b.(Int)), nil
	})
	m.Register(StringTypeId, StringTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(String) == b.(String)), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Float) == b.(Float)), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(Float(a.(Int)) == b.(Float)), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Float) == Float(b.(Int))), nil
	})
	deepEqual := &operationMatrixDeepEqual{equal: m, ef: func(st funcGen.Stack[Value], a, b Value) (bool, error) {
		eq, err := m.Calc(st, a, b)
		return bool(eq.(Bool)), err
	}}

	ef := func(st funcGen.Stack[Value], a, b Value) (bool, error) {
		eq, err := deepEqual.Calc(st, a, b)
		return bool(eq.(Bool)), err
	}
	fg.equal = ef
	fg.FunctionGenerator.SetIsEqual(ef)
	return deepEqual
}

type operationMatrixDeepEqual struct {
	equal OperationMatrix
	ef    funcGen.BoolFunc[Value]
}

func (o *operationMatrixDeepEqual) Calc(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(*List); ok {
		if bb, ok := b.(*List); ok {
			equals, err := aa.Equals(st, bb, o.ef)
			return Bool(equals), err
		}
	}
	if aa, ok := a.(Map); ok {
		if bb, ok := b.(Map); ok {
			equals, err := aa.Equals(st, bb, o.ef)
			return Bool(equals), err
		}
	}
	return o.equal.Calc(st, a, b)
}

func (o *operationMatrixDeepEqual) Register(a, b Type, op funcGen.OperatorFunc[Value]) {
	o.equal.Register(a, b, op)
}

func Less(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "<")
	m.Register(IntTypeId, IntTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Int) < b.(Int)), nil
	})
	m.Register(StringTypeId, StringTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(String) < b.(String)), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Float) < b.(Float)), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(Float(a.(Int)) < b.(Float)), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(_ funcGen.Stack[Value], a, b Value) (Value, error) {
		return Bool(a.(Float) < Float(b.(Int))), nil
	})

	eq := func(st funcGen.Stack[Value], a, b Value) (bool, error) {
		eq, err := m.Calc(st, a, b)
		return bool(eq.(Bool)), err
	}
	fg.less = eq
	return m
}

func notAllowed(name string, a Value, b Value) error {
	return fmt.Errorf("'%s' not allowed on %s, %s", name, TypeName(a), TypeName(b))
}

type operationMatrixStringAdd struct {
	parent OperationMatrix
}

func (o operationMatrixStringAdd) Calc(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if a.GetType() == StringTypeId {
		str, err := b.ToString(st)
		return a.(String) + String(str), err
	}
	return o.parent.Calc(st, a, b)
}

func (o operationMatrixStringAdd) Register(a, b Type, op funcGen.OperatorFunc[Value]) {
	o.parent.Register(a, b, op)
}

func Add(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "+")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) + b.(Int), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) + b.(Float), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(a.(Int)) + b.(Float), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) + Float(b.(Int)), nil
	})
	m.Register(ListTypeId, ListTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return NewListFromIterable(iterator.Append(a.(*List).iterable, b.(*List).iterable)), nil
	})
	m.Register(MapTypeId, MapTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Map).Merge(b.(Map))
	})
	return operationMatrixStringAdd{m}
}

func Sub(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "-")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) - b.(Int), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) - b.(Float), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(a.(Int)) - b.(Float), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) - Float(b.(Int)), nil
	})
	return m
}

func Left(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "<<")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) << b.(Int), nil
	})
	return m
}

func Right(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, ">>")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) >> b.(Int), nil
	})
	return m
}

func Mod(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "%")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) % b.(Int), nil
	})
	return m
}

func Mul(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "*")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) * b.(Int), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) * b.(Float), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(a.(Int)) * b.(Float), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) * Float(b.(Int)), nil
	})
	return m
}

func Div(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "/")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(a.(Int)) / Float(b.(Int)), nil
	})
	m.Register(FloatTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) / b.(Float), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(a.(Int)) / b.(Float), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Float) / Float(b.(Int)), nil
	})
	return m
}

func Pow(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "^")
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		aa := a.(Int)
		bb := b.(Int)
		if bb > 0 && bb < 10 {
			n := int(aa)
			for j := 1; j < int(bb); j++ {
				n *= int(aa)
			}
			return Int(n), nil
		} else {
			return Int(math.Pow(float64(aa), float64(bb))), nil
		}
	})
	m.Register(FloatTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(math.Pow(float64(a.(Float)), float64(b.(Float)))), nil
	})
	m.Register(FloatTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(math.Pow(float64(a.(Float)), float64(b.(Int)))), nil
	})
	m.Register(IntTypeId, FloatTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return Float(math.Pow(float64(a.(Int)), float64(b.(Float)))), nil
	})
	return m
}

func Neg(a Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		return -aa, nil
	}
	if aa, ok := a.ToFloat(); ok {
		return Float(-aa), nil
	}
	return nil, fmt.Errorf("neg not allowed on -%s", TypeName(a))
}

func Not(a Value) (Value, error) {
	if aa, ok := a.(Bool); ok {
		return !aa, nil
	}
	return nil, fmt.Errorf("not not allowed on !%s", TypeName(a))
}

func And(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "&")
	m.Register(BoolTypeId, BoolTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Bool) && b.(Bool), nil
	})
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) & b.(Int), nil
	})
	return m
}

func Or(fg *FunctionGenerator) OperationMatrix {
	m := NewOperationMatrix(fg, "|")
	m.Register(BoolTypeId, BoolTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Bool) || b.(Bool), nil
	})
	m.Register(IntTypeId, IntTypeId, func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return a.(Int) | b.(Int), nil
	})
	return m
}
