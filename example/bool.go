package example

import (
	"github.com/hneemann/parser2/funcGen"
)

// boolParser is a simple parser for bool expressions
// see test cases for usage example
var boolParser = funcGen.New[bool]().
	AddConstant("false", false).
	AddConstant("true", true).
	AddOp("^", true, func(st funcGen.Stack[bool], a, b bool) (bool, error) { return a != b, nil }).
	AddOp("=", true, func(st funcGen.Stack[bool], a, b bool) (bool, error) { return a == b, nil }).
	AddOp("|", true, func(st funcGen.Stack[bool], a, b bool) (bool, error) { return a || b, nil }).
	AddOp("&", true, func(st funcGen.Stack[bool], a, b bool) (bool, error) { return a && b, nil }).
	AddUnary("!", func(a bool) (bool, error) { return !a, nil }).
	SetToBool(func(c bool) (bool, bool) { return c, true })
