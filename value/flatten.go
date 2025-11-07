package value

import (
	"github.com/hneemann/parser2/funcGen"
)

// Flatten takes a Value and returns a function that yields all non-list, non-map Values contained within it.
// If the Value is a list or map, it recursively flattens its contents.
// The returned function takes a yield function as an argument, which is called with each flattened Value and
// any error encountered.
// The yield function should return true to continue yielding values or false to stop.
func Flatten(v Value) func(yield func(v Value, err error) bool) {
	return func(yield func(v Value, err error) bool) {
		flatten(v, nil, yield)
	}
}

// FlattenStack takes a stack of Values and an integer start index, and returns a function that yields all non-list,
// non-map Values contained within the Values in the stack starting from the specified index.
// It uses the same flattening logic as Flatten.
func FlattenStack(st funcGen.Stack[Value], start int) func(yield func(v Value, err error) bool) {
	return func(yield func(v Value, err error) bool) {
		for i := start; i < st.Size(); i++ {
			if !flatten(st.Get(i), nil, yield) {
				return
			}
		}
	}
}

func flatten(v Value, err error, yield func(v Value, err error) bool) bool {
	if list, ok := v.ToList(); ok && err == nil {
		for v, err := range list.Iterate(funcGen.NewEmptyStack[Value]()) {
			if !flatten(v, err, yield) {
				return false
			}
		}
		return true
	} else if m, ok := v.ToMap(); ok && err == nil {
		for _, value := range m.Iter {
			if !flatten(value, nil, yield) {
				return false
			}
		}
		return true
	} else {
		return yield(v, err)
	}
}
