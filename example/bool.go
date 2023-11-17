package example

import (
	"github.com/hneemann/parser2/funcGen"
)

// boolParser is a simple parser for bool expressions
// see test cases for usage example
var boolParser = funcGen.New[bool]().
	AddConstant("false", false).
	AddConstant("true", true).
	AddOp("^", true, func(a, b bool) bool { return a != b }).
	AddOp("=", true, func(a, b bool) bool { return a == b }).
	AddOp("|", true, func(a, b bool) bool { return a || b }).
	AddOp("&", true, func(a, b bool) bool { return a && b }).
	AddUnary("!", func(a bool) bool { return !a }).
	SetToBool(func(c bool) (bool, bool) { return c, true })
