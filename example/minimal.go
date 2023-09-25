package example

import (
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"math"
	"strconv"
)

// minimal is a minimal float64 parser example
// see test cases for usage example
var minimal = funcGen.New[float64]().
	AddConstant("pi", math.Pi).
	AddOp("=", true, func(a, b float64) float64 { return fromBool(a == b) }).
	AddOp("<", false, func(a, b float64) float64 { return fromBool(a < b) }).
	AddOp(">", false, func(a, b float64) float64 { return fromBool(a > b) }).
	AddOp("+", true, func(a, b float64) float64 { return a + b }).
	AddOp("-", false, func(a, b float64) float64 { return a - b }).
	AddOp("*", true, func(a, b float64) float64 { return a * b }).
	AddOp("/", false, func(a, b float64) float64 { return a / b }).
	AddOp("^", false, func(a, b float64) float64 { return math.Pow(a, b) }).
	AddUnary("-", func(a float64) float64 { return -a }).
	AddSimpleFunction("sin", math.Sin).
	AddSimpleFunction("sqrt", math.Sqrt).
	SetToBool(func(c float64) bool { return c != 0 }).
	SetNumberParser(
		parser2.NumberParserFunc[float64](
			func(n string) (float64, error) {
				return strconv.ParseFloat(n, 64)
			},
		),
	)

func fromBool(b bool) float64 {
	if b {
		return 1
	} else {
		return 0
	}
}
