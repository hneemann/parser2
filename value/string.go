package value

import (
	"github.com/hneemann/parser2/funcGen"
	"strings"
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

func (s String) ToString() (string, bool) {
	return string(s), true
}

func (s String) Contains(st funcGen.Stack[Value]) Value {
	if substr, ok := st.Get(1).ToString(); ok {
		return Bool(strings.Contains(string(s), substr))
	} else {
		panic("contains requires a string as argument")
	}
}

func (s String) IndexOf(st funcGen.Stack[Value]) Value {
	if substr, ok := st.Get(1).ToString(); ok {
		return Int(strings.Index(string(s), substr))
	} else {
		panic("contains requires a string as argument")
	}
}

var StringMethods = MethodMap{
	"len":      methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return Int(len(string(str))) }),
	"toLower":  methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToLower(string(str))) }),
	"toUpper":  methodAtType(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToUpper(string(str))) }),
	"contains": methodAtType(2, func(str String, stack funcGen.Stack[Value]) Value { return str.Contains(stack) }),
	"indexOf":  methodAtType(2, func(str String, stack funcGen.Stack[Value]) Value { return str.IndexOf(stack) }),
}

func (s String) GetMethod(name string) (funcGen.Function[Value], error) {
	return StringMethods.Get(name)
}
