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

func methodAtString(args int, method func(str String, stack funcGen.Stack[Value]) Value) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) Value {
		if obj, ok := stack.Get(0).ToString(); ok {
			return method(String(obj), stack)
		}
		panic("call of list method on non list")
	}, Args: args, IsPure: true}
}

var StringMethods = map[string]funcGen.Function[Value]{
	"len":      methodAtString(1, func(str String, stack funcGen.Stack[Value]) Value { return Int(len(string(str))) }),
	"toLower":  methodAtString(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToLower(string(str))) }),
	"toUpper":  methodAtString(1, func(str String, stack funcGen.Stack[Value]) Value { return String(strings.ToUpper(string(str))) }),
	"contains": methodAtString(2, func(str String, stack funcGen.Stack[Value]) Value { return str.Contains(stack) }),
	"indexOf":  methodAtString(2, func(str String, stack funcGen.Stack[Value]) Value { return str.IndexOf(stack) }),
}

func (s String) GetMethod(name string) (funcGen.Function[Value], bool) {
	m, ok := StringMethods[name]
	return m, ok
}
