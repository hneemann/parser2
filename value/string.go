package value

import (
	"bytes"
	"errors"
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

func (s String) ToString() (string, error) {
	return string(s), nil
}

func (s String) Contains(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return Bool(strings.Contains(string(s), string(s2))), nil
	} else {
		return nil, errors.New("contains needs a string as argument")
	}
}

func (s String) IndexOf(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return Int(strings.Index(string(s), string(s2))), nil
	} else {
		return nil, errors.New("indexOf needs a string as argument")
	}
}

func (s String) Split(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return NewListConvert(func(s string) Value { return String(s) }, strings.Split(string(s), string(s2))...), nil
	} else {
		return nil, errors.New("split needs a string as argument")
	}
}

func (s String) Cut(st funcGen.Stack[Value]) (Value, error) {
	if p, ok := st.Get(1).ToInt(); ok {
		if n, ok := st.Get(2).ToInt(); ok {
			str := string(s)
			for i := 0; i < p; i++ {
				_, l := utf8.DecodeRuneInString(str)
				str = str[l:]
				if len(str) == 0 {
					return String(""), nil
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
					return String(res.String()), nil
				}
			}
			return String(res.String()), nil
		}
	}
	return nil, errors.New("cut requires integers as arguments (pos,len)")
}

var StringMethods = MethodMap{
	"len":    MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return Int(len(string(str))), nil }),
	"string": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str, nil }),
	"trim": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
		return String(strings.TrimSpace(string(str))), nil
	}),
	"toLower": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
		return String(strings.ToLower(string(str))), nil
	}),
	"toUpper": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
		return String(strings.ToUpper(string(str))), nil
	}),
	"contains": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Contains(stack) }),
	"indexOf":  MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.IndexOf(stack) }),
	"split":    MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Split(stack) }),
	"cut":      MethodAtType(2, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Cut(stack) }),
}

func (s String) GetMethod(name string) (funcGen.Function[Value], error) {
	return StringMethods.Get(name)
}
