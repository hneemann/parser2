package value

import (
	"fmt"
	"github.com/hneemann/iterator"
	"math"
	"strings"
)

func Equal(a Value, b Value) bool {
	switch aa := a.(type) {
	case Bool:
		if bb, ok := b.(Bool); ok {
			return aa == bb
		}
	case Int:
		if bb, ok := b.(Int); ok {
			return aa == bb
		}
	case String:
		if bb, ok := b.(String); ok {
			return aa == bb
		}
	case *List:
		if bb, ok := b.(*List); ok {
			return aa.Equals(bb)
		}
	case Map:
		if bb, ok := b.(Map); ok {
			return aa.Equals(bb)
		}
	default:
		if aa, ok := a.ToFloat(); ok {
			if bb, ok := b.ToFloat(); ok {
				return aa == bb
			}
		}
	}
	return false
}

func LessEqual(a Value, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Bool(aa <= bb), nil
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return Bool(aa <= bb), nil
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Bool(aa <= bb), nil
		}
	}
	return nil, fmt.Errorf("less not allowed on %v<%v", a, b)
}

func In(a Value, b Value) (Value, error) {
	if list, ok := b.(*List); ok {
		if search, ok := a.(*List); ok {
			return Bool(list.containsAllItems(search)), nil
		} else {
			return Bool(list.containsItem(a)), nil
		}
	}
	if m, ok := b.(Map); ok {
		if key, ok := a.(String); ok {
			return m.ContainsKey(key), nil
		}
	}
	if strToLookFor, ok := a.(String); ok {
		if strToLookIn, ok := b.(String); ok {
			return Bool(strings.Contains(string(strToLookIn), string(strToLookFor))), nil
		}
	}
	return nil, fmt.Errorf("~ not allowed on %v~%v", a, b)
}

func Less(a Value, b Value) (bool, error) {
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
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return aa < bb, nil
		}
	}
	return false, fmt.Errorf("less not allowed on %v<%v", a, b)
}

func Swap(inner func(a, b Value) (Value, error)) func(a, b Value) (Value, error) {
	return func(a, b Value) (Value, error) {
		return inner(b, a)
	}
}

func Add(a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa + bb, nil
		}
	}
	if aa, ok := a.(String); ok {
		return aa + String(b.String()), nil
	}
	if aa, ok := a.(*List); ok {
		if bb, ok := b.(*List); ok {
			return NewListFromIterable(iterator.Append(aa.iterable, bb.iterable)), nil
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa + bb), nil
		}
	}
	return nil, fmt.Errorf("add not allowed on %v+%v", a, b)
}

func Sub(a, b Value) (Value, error) {
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
	return nil, fmt.Errorf("sub not allowed on %v-%v", a, b)
}

func Left(a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa << bb, nil
		}
	}
	return nil, fmt.Errorf("<< not allowed on %v<<%v", a, b)
}
func Right(a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa >> bb, nil
		}
	}
	return nil, fmt.Errorf(">> not allowed on %v>>%v", a, b)
}
func Mod(a, b Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa % bb, nil
		}
	}
	return nil, fmt.Errorf("%% not allowed on %v%%%v", a, b)
}

func Mul(a, b Value) (Value, error) {
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
	return nil, fmt.Errorf("mul not allowed on %v*%v", a, b)
}

func Div(a, b Value) (Value, error) {
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa / bb), nil
		}
	}
	return nil, fmt.Errorf("div not allowed on %v/%v", a, b)
}

func Pow(a, b Value) (Value, error) {
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
	return nil, fmt.Errorf("^ not allowed on %v^%v", a, b)
}

func Neg(a Value) (Value, error) {
	if aa, ok := a.(Int); ok {
		return -aa, nil
	}
	if aa, ok := a.ToFloat(); ok {
		return Float(-aa), nil
	}
	return nil, fmt.Errorf("neg not allowed on -%v", a)
}

func Not(a Value) (Value, error) {
	if aa, ok := a.(Bool); ok {
		return !aa, nil
	}
	return nil, fmt.Errorf("not not allowed on !%v", a)
}

func And(a, b Value) (Value, error) {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa && bb), nil
		}
	}
	return nil, fmt.Errorf("& not allowed on %v&%v", a, b)
}

func Or(a, b Value) (Value, error) {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa || bb), nil
		}
	}
	return nil, fmt.Errorf("| not allowed on %v&%v", a, b)
}
