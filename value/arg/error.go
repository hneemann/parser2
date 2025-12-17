package arg

import (
	"errors"
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"strings"
)

type pError struct {
	err error
}

func (a pError) Error() string {
	return a.err.Error()
}

func PanicErr(str string) error {
	return pError{err: errors.New(str)}
}

func PanicfErr(str string, args ...any) error {
	return pError{err: fmt.Errorf(str, args...)}
}

func PanicToError(work func() value.Value) (v value.Value, e error) {
	defer CatchErr(&e)
	return work(), nil
}

func CatchErr(e *error) {
	if r := recover(); r != nil {
		if ae, ok := r.(pError); ok {
			*e = ae.err
		} else {
			panic(r)
		}
	}
}

func GetFromStack[R value.Value](st funcGen.Stack[value.Value], n int) (r R) {
	vv := st.Get(n)

	switch any(r).(type) {
	case value.Float:
		if v, ok := vv.ToFloat(); ok {
			vv = value.Float(v)
		}
	case *value.List:
		if v, ok := vv.ToList(); ok {
			vv = v
		}
	case *value.Map:
		if v, ok := vv.ToMap(); ok {
			vv = v
		}
	}

	if v, ok := vv.(R); ok {
		return v
	} else {
		var zero R
		panic(PanicfErr("expected %d. argument to be of type %s but got %s", n, typeStr(zero), typeStr(vv)))
	}
}

func GetFromStackOptional[R value.Value](st funcGen.Stack[value.Value], n int, def R) R {
	if n < st.Size() {
		return GetFromStack[R](st, n)
	}
	return def
}

func typeStr(v value.Value) string {
	t := fmt.Sprintf("%T", v)
	p := strings.Index(t, ".")
	if p >= 0 {
		return t[p+1:]
	}
	return t
}

func Try(err error) {
	if err != nil {
		panic(pError{err})
	}
}

func TryArg[T any](t T, err error) T {
	if err != nil {
		panic(pError{err})
	}
	return t
}

func TryArgs[A any, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(pError{err})
	}
	return a, b
}

func checkStackM(stack funcGen.Stack[value.Value], min int, max int) {
	s := stack.Size()
	if s < min {
		panic(PanicfErr("expected at least %d arguments but got %d", min-1, s-1))
	}
	if s > max {
		panic(PanicfErr("expected at most %d arguments but got %d", max-1, s-1))
	}
}

func checkStackF(stack funcGen.Stack[value.Value], min int, max int) {
	s := stack.Size()
	if s < min {
		panic(PanicfErr("expected at least %d arguments but got %d", min, s))
	}
	if s > max {
		panic(PanicfErr("expected at most %d arguments but got %d", max, s))
	}
}
