package arg

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testFunc() (err error) {
	defer CatchErr(&err)
	Try(fmt.Errorf("throw error"))
	return nil
}

func TestError(t *testing.T) {
	err := testFunc()
	assert.Error(t, err)
	assert.Equal(t, "throw error", err.Error())
}

func testFunc2() (err error) {
	defer CatchErr(&err)

	panic("test panic")

	return nil
}

func TestErrorPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Equal(t, "test panic", r)
	}()
	testFunc2()
	assert.Fail(t, "should not reach here")
}

func testInnerFuncArg(e bool) (int, error) {
	if e {
		return 1, fmt.Errorf("throw error")
	} else {
		return 2, nil
	}
}

func testFuncArg() (i int, err error) {
	defer CatchErr(&err)
	j := TryArg(testInnerFuncArg(true))
	return j, nil
}

func testFuncArg2() (i int, err error) {
	defer CatchErr(&err)
	j := TryArg(testInnerFuncArg(false))
	return j, nil
}

func TestErrorArg(t *testing.T) {
	i, err := testFuncArg()
	assert.Error(t, err)
	assert.Equal(t, 0, i)
	assert.Equal(t, "throw error", err.Error())

	i, err = testFuncArg2()
	assert.NoError(t, err)
	assert.Equal(t, 2, i)
}

type TestType struct {
	a float64
}

func (t TestType) ToList() (*value.List, bool) {
	return value.NewList(value.Float(t.a)), true
}

func (t TestType) ToMap() (value.Map, bool) {
	return value.NewMap(listMap.New[value.Value](1).Append("a", value.Float(t.a))), true
}

func (t TestType) ToFloat() (float64, bool) {
	return t.a, true
}

func (t TestType) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprintf("TestType(%v)", t.a), nil
}

func (t TestType) GetType() value.Type {
	return 0
}

func TestGetFromStack(t *testing.T) {
	st := funcGen.NewStack[value.Value](TestType{1.5})

	f := GetFromStack[value.Float](st, 0)
	assert.Equal(t, value.Float(1.5), f)

	l := GetFromStack[*value.List](st, 0)
	slice, err := l.ToSlice(st)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(slice))
	assert.Equal(t, value.Float(1.5), slice[0])
}

func TestMethod2(t *testing.T) {
	v := Function2[value.Float, *value.List](func(a value.Float, b *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
		l := TryArg(b.ToSlice(st))
		if l0, ok := l[0].ToFloat(); ok {
			return a + value.Float(l0), nil
		}
		return nil, fmt.Errorf("expected float in list")
	})
	st := funcGen.NewEmptyStack[value.Value]()
	eval, err := v.EvalSt(st, TestType{1.5}, TestType{1.5})
	assert.NoError(t, err)
	assert.Equal(t, value.Float(3.0), eval)
}

func TestMethod2opt(t *testing.T) {
	list, ok := TestType{1.5}.ToList()
	assert.True(t, ok)
	v := Function2[value.Float, *value.List](func(a value.Float, b *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
		l := TryArg(b.ToSlice(st))
		if l0, ok := l[0].ToFloat(); ok {
			return a + value.Float(l0), nil
		}
		return nil, fmt.Errorf("expected float in list")
	}, list)
	st := funcGen.NewEmptyStack[value.Value]()
	eval, err := v.EvalSt(st, TestType{1.5})
	assert.NoError(t, err)
	assert.Equal(t, value.Float(3.0), eval)
}
