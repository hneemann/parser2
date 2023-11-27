package value

import (
	"fmt"
	"github.com/hneemann/iterator"
	"math"
	"strings"
)

func Equal(a Value, b Value) bool {
	if aa, ok := a.(Bool); ok {
		if bb, ok := b.(Bool); ok {
			return aa == bb
		}
	}
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa == bb
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return aa == bb
		}
	}
	if aa, ok := a.(*List); ok {
		if bb, ok := b.(*List); ok {
			return aa.Equals(bb)
		}
	}
	if aa, ok := a.(Map); ok {
		if bb, ok := b.(Map); ok {
			return aa.Equals(bb)
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return aa == bb
		}
	}
	return false
}

func LessEqual(a Value, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Bool(aa <= bb)
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return Bool(aa <= bb)
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Bool(aa <= bb)
		}
	}
	panic(fmt.Errorf("less not allowed on %v<%v", a, b))
}

func In(a Value, b Value) Value {
	if list, ok := b.(*List); ok {
		if search, ok := a.(*List); ok {
			ls := list.ToSlice()
			for _, s := range search.ToSlice() {
				found := false
				for _, i := range ls {
					if Equal(s, i) {
						found = true
					}
				}
				if !found {
					return Bool(false)
				}
			}
			return Bool(true)
		} else {
			found := false
			list.iterable()(func(value Value) bool {
				if Equal(a, value) {
					found = true
					return false
				}
				return true
			})
			return Bool(found)
		}
	}
	if strToLookFor, ok := a.(String); ok {
		if strToLookIn, ok := b.(String); ok {
			return Bool(strings.Contains(string(strToLookIn), string(strToLookFor)))
		}
	}
	panic(fmt.Errorf("~ not allowed on %v~%v", a, b))
}

func Less(a Value, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Bool(aa < bb)
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return Bool(aa < bb)
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Bool(aa < bb)
		}
	}
	panic(fmt.Errorf("less not allowed on %v<%v", a, b))
}

func Swap(inner func(a, b Value) Value) func(a, b Value) Value {
	return func(a, b Value) Value {
		return inner(b, a)
	}
}

func Add(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa + bb
		}
	}
	if aa, ok := a.(String); ok {
		return aa + String(b.String())
	}
	if aa, ok := a.(*List); ok {
		if bb, ok := b.(*List); ok {
			return NewListFromIterable(iterator.Append(aa.iterable, bb.iterable))
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa + bb)
		}
	}
	panic(fmt.Errorf("add not allowed on %v+%v", a, b))
}

func Sub(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa - bb
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa - bb)
		}
	}
	panic(fmt.Errorf("sub not allowed on %v-%v", a, b))
}

func Left(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa << bb
		}
	}
	panic(fmt.Errorf("<< not allowed on %v<<%v", a, b))
}
func Right(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa >> bb
		}
	}
	panic(fmt.Errorf(">> not allowed on %v>>%v", a, b))
}
func Mod(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa % bb
		}
	}
	panic(fmt.Errorf("%% not allowed on %v%%%v", a, b))
}

func Mul(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return aa * bb
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa * bb)
		}
	}
	panic(fmt.Errorf("mul not allowed on %v*%v", a, b))
}

func Div(a, b Value) Value {
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(aa / bb)
		}
	}
	panic(fmt.Errorf("div not allowed on %v/%v", a, b))
}

func Pow(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			if bb > 0 && bb < 10 {
				n := int(aa)
				for j := 1; j < int(bb); j++ {
					n *= int(aa)
				}
				return Int(n)
			} else {
				return Int(math.Pow(float64(aa), float64(bb)))
			}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float(math.Pow(aa, bb))
		}
	}
	panic(fmt.Errorf("^ not allowed on %v^%v", a, b))
}

func Neg(a Value) Value {
	if aa, ok := a.(Int); ok {
		return -aa
	}
	if aa, ok := a.ToFloat(); ok {
		return Float(-aa)
	}
	panic(fmt.Errorf("neg not allowed on -%v", a))
}

func Not(a Value) Value {
	if aa, ok := a.(Bool); ok {
		return !aa
	}
	panic(fmt.Errorf("not not allowed on !%v", a))
}

func And(a, b Value) Value {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa && bb)
		}
	}
	panic(fmt.Errorf("& not allowed on %v&%v", a, b))
}

func Or(a, b Value) Value {
	if aa, ok := a.ToBool(); ok {
		if bb, ok := b.ToBool(); ok {
			return Bool(aa || bb)
		}
	}
	panic(fmt.Errorf("| not allowed on %v&%v", a, b))
}
