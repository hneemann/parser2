package example

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"math"
)

var DynType = value.New().
	AddConstant("pi", value.Float{F: math.Pi}).
	AddConstant("true", value.Bool{B: true}).
	AddConstant("false", value.Bool{B: false}).
	AddStaticFunction("sprintf", funcGen.Function[value.Value]{
		Func:   sprintf,
		Args:   -1,
		IsPure: true,
	})

func sprintf(st funcGen.Stack[value.Value], cs []value.Value) value.Value {
	switch st.Size() {
	case 0:
		return value.String{}
	case 1:
		return value.String{S: fmt.Sprint(st.Get(0))}
	default:
		if s, ok := st.Get(0).(value.String); ok {
			values := make([]any, st.Size()-1)
			for i := 1; i < st.Size(); i++ {
				values[i-1] = st.Get(i)
			}
			return value.String{S: fmt.Sprintf(s.S, values...)}
		} else {
			panic("sprintf requires string as first argument")
		}
	}
}
