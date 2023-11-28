package value

import "testing"

func TestString(t *testing.T) {
	runTest(t, []testType{
		{exp: "\"Hello World\".len()", res: Int(11)},
		{exp: "\"Hello World\".indexOf(\"Wo\")", res: Int(6)},
		{exp: "\"Hello World\".toLower()", res: String("hello world")},
		{exp: "\"Hello World\".toUpper()", res: String("HELLO WORLD")},
		{exp: "\"Hello World\".contains(\"Wo\")", res: Bool(true)},
		{exp: "\"Hello World\".contains(\"wo\")", res: Bool(false)},
		{exp: "\"Wo\" ~ \"Hello World\"", res: Bool(true)},
		{exp: "\" Hello \".trim()", res: String("Hello")},
		{exp: "\"Hello,World\".split(\",\").reduce((a,b)->a+\"|\"+b)", res: String("Hello|World")},
		{exp: "\"Hello , World\".split(\",\").reduce((a,b)->a+\"|\"+b)", res: String("Hello | World")},
	})
}
