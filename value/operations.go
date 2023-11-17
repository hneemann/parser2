package value

import "fmt"

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
			return Bool{B: aa.I <= bb.I}
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return Bool{B: aa.S <= bb.S}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Bool{B: aa <= bb}
		}
	}
	panic(fmt.Errorf("less not allowed on %v<%v", a, b))
}

func Less(a Value, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Bool{B: aa.I < bb.I}
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return Bool{B: aa.S < bb.S}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Bool{B: aa < bb}
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
			return Int{I: aa.I + bb.I}
		}
	}
	if aa, ok := a.(String); ok {
		if bb, ok := b.(String); ok {
			return String{S: aa.S + bb.S}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float{F: aa + bb}
		}
	}
	panic(fmt.Errorf("add not allowed on %v+%v", a, b))
}

func Sub(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Int{I: aa.I - bb.I}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float{F: aa - bb}
		}
	}
	panic(fmt.Errorf("sub not allowed on %v-%v", a, b))
}

func Mul(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Int{I: aa.I * bb.I}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float{F: aa * bb}
		}
	}
	panic(fmt.Errorf("mul not allowed on %v*%v", a, b))
}

func Div(a, b Value) Value {
	if aa, ok := a.(Int); ok {
		if bb, ok := b.(Int); ok {
			return Int{I: aa.I / bb.I}
		}
	}
	if aa, ok := a.ToFloat(); ok {
		if bb, ok := b.ToFloat(); ok {
			return Float{F: aa / bb}
		}
	}
	panic(fmt.Errorf("div not allowed on %v/%v", a, b))
}

func Neg(a Value) Value {
	if aa, ok := a.(Int); ok {
		return Int{I: -aa.I}
	}
	if aa, ok := a.ToFloat(); ok {
		return Float{F: -aa}
	}
	panic(fmt.Errorf("neg not allowed on -%v", a))
}

func Not(a Value) Value {
	if aa, ok := a.(Bool); ok {
		return Bool{B: !aa.B}
	}
	panic(fmt.Errorf("not not allowed on !%v", a))
}
