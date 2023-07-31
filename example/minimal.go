package example

import (
	"github.com/hneemann/parser2"
	"math"
	"strconv"
)

// minimal is a minimal float64 parser example
// see test cases for usage example
var minimal = parser2.New[float64]().
	AddConstant("pi", math.Pi).
	AddOp("+", true, func(a, b float64) float64 { return a + b }).
	AddOp("-", false, func(a, b float64) float64 { return a - b }).
	AddOp("*", true, func(a, b float64) float64 { return a * b }).
	AddOp("/", false, func(a, b float64) float64 { return a / b }).
	AddOp("^", false, func(a, b float64) float64 { return math.Pow(a, b) }).
	AddUnary("-", func(a float64) float64 { return -a }).
	AddSimpleFunction("sin", math.Sin).
	AddSimpleFunction("sqrt", math.Sqrt).
	SetNumberParser(
		parser2.NumberParserFunc[float64](
			func(n string) (float64, error) {
				return strconv.ParseFloat(n, 64)
			},
		),
	)
