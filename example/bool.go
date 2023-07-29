package example

import (
	"github.com/hneemann/parser2"
)

// boolParser is a simple parser for bool expressions
// see test cases for usage example
var boolParser = parser2.New[bool]().
	AddConstant("false", false).
	AddConstant("true", true).
	AddOp("|", func(a, b bool) bool { return a || b }).
	AddOp("&", func(a, b bool) bool { return a && b }).
	AddUnary("!", func(a bool) bool { return !a })
