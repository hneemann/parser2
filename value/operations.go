package value

import (
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"math"
)

// Equal does not cover lists and maps
func Equal(st funcGen.Stack[Value], a Value, b Value) (bool, error) {
	switch aa := a.(type) {
	case Bool:
		if bb, ok := b.(Bool); ok {
			return aa == bb, nil
		}
	case Int:
		if bb, ok := b.(Int); ok {
			return aa == bb, nil
		}
	case String:
		if bb, ok := b.(String); ok {
			return aa == bb, nil
		}
	case nilType:
		if _, ok := b.(nilType); ok {
			return true, nil
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return aa == bb, nil
		}
	}
	return false, nil
}

func Less(st funcGen.Stack[Value], a Value, b Value) (bool, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa < bb, nil
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return aa < bb, nil
		}
	}
	// allows int-float comparison also
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return aa < bb, nil
		}
	}
	return false, notAllowed("less", a, b)
}

func notAllowed(name string, a Value, b Value) error {
	return fmt.Errorf("'%s' not allowed on %s, %s", name, TypeName(a), TypeName(b))
}

func Add(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa + bb, nil
		}
	}
	if aa, ok := a.(String); ok {
		s, err := b.ToString(st)
		if err != nil {
			return nil, err
		}
		return aa + String(s), nil
	}
	if aa, ok := a.(*List); ok {
		if bb, ok := b.(*List); ok {
			return NewListFromIterable(iterator.Append(aa.iterable, bb.iterable)), nil
		}
	}
	if aa, ok := a.(Map); ok {
		if bb, ok := b.(Map); ok {
			return aa.Merge(bb)
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa + bb), nil
		}
	}
	return nil, notAllowed("add", a, b)
}

func Sub(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa - bb, nil
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa - bb), nil
		}
	}
	return nil, notAllowed("sub", a, b)
}

func Left(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa << bb, nil
		}
	}
	return nil, notAllowed("<<", a, b)
}
func Right(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa >> bb, nil
		}
	}
	return nil, notAllowed(">>", a, b)
}
func Mod(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa % bb, nil
		}
	}
	return nil, notAllowed("%", a, b)
}

func Mul(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa * bb, nil
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa * bb), nil
		}
	}
	return nil, notAllowed("mul", a, b)
}

func Div(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			if bb == 0 {
				return nil, errors.New("division by zero")
			}
			return Float(aa / bb), nil
		}
	}
	return nil, notAllowed("div", a, b)
}

func Pow(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			if bb > 0 && bb < 10 {
				n := int(aa)
				for j := 1; j < int(bb); j++ {
					n *= int(aa)
				}
				return Int(n), nil
			} else {
				return Int(math.Pow(float64(aa), float64(bb))), nil
			}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(math.Pow(aa, bb)), nil
		}
	}
	return nil, notAllowed("^", a, b)
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

func And(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa && bb), nil
		}
	}
	return nil, notAllowed("&", a, b)
}

func Or(st funcGen.Stack[Value], a, b Value) (Value, error) {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa || bb), nil
		}
	}
	return nil, notAllowed("|", a, b)
}
