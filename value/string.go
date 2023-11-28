package value

import (
	"bytes"
	"github.com/hneemann/parser2/funcGen"
	"math"
	"strings"
	"unicode/utf8"
)

type String string

func (s String) ToList() (*List, bool) {
	return nil, false
}

func (s String) ToMap() (Map, bool) {
	return Map{}, false
}

func (s String) ToInt() (int, bool) {
	return 0, false
}

func (s String) ToFloat() (float64, bool) {
	return 0, false
}

func (s String) ToBool() (bool, bool) {
	return false, false
}

func (s String) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (s String) String() string {
	return string(s)
}

func (s String) Contains(st funcGen.Stack[Value]) Value {
	return Bool(strings.Contains(string(s), st.Get(1).String()))
}

func (s String) IndexOf(st funcGen.Stack[Value]) Value {
	return Int(strings.Index(string(s), st.Get(1).String()))
}

func (s String) Split(st funcGen.Stack[Value]) Value {
	return NewListCreate(func(s string) Value { return String(s) }, strings.Split(string(s), st.Get(1).String())...)
}

func (s String) Cut(st funcGen.Stack[Value]) Value {
	if p, ok := st.Get(1).ToInt(); ok {
		if n, ok := st.Get(2).ToInt(); ok {
			str := string(s)
			for i := 0; i < p; i++ {
				_, l := utf8.DecodeRuneInString(str)
				str = str[l:]
				if len(str) == 0 {
					return String("")
				}
			}
			var res bytes.Buffer
			if n <= 0 {
				n = math.MaxInt
			}
			for i := 0; i < n; i++ {
				r, l := utf8.DecodeRuneInString(str)
				res.WriteRune(r)
				str = str[l:]
				if len(str) == 0 {
					return String(res.String())
				}
			}
			return String(res.String())
		}
	}
	panic("cur requires in arguments (pos,len)")
}

var StringMethods = MethodMap{
	"len":      methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return Int(len(string(str))) }),
	"trim":     methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.TrimSpace(string(str))) }),
	"toLower":  methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToLower(string(str))) }),
	"toUpper":  methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToUpper(string(str))) }),
	"contains": methodAtType(2, func(str String, stack funcGen.Stack[Value]) Value { return str.Contains(stack) }),
	"indexOf":  methodAtType(2, func(str String, stack funcGen.Stack[Value]) Value { return str.IndexOf(stack) }),
	"split":    methodAtType(2, func(str String, stack funcGen.Stack[Value]) Value { return str.Split(stack) }),
	"cut":      methodAtType(3, func(str String, stack funcGen.Stack[Value]) Value { return str.Cut(stack) }),
}

func (s String) GetMethod(name string) (funcGen.Function[Value], error) {
	return StringMethods.Get(name)
}
