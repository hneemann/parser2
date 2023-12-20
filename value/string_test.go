package value

import "testing"

func TestString(t *testing.T) {
	runTest(t, []testType{
		{exp: "\"Hello World\".len()", res: Int(11)},
		{exp: "\"Hello World\".string()", res: String("Hello World")},
		{exp: "\"Hello World\".indexOf(\"Wo\")", res: Int(6)},
		{exp: "\"Hello World\".toLower()", res: String("hello world")},
		{exp: "\"Hello World\".toUpper()", res: String("HELLO WORLD")},
		{exp: "\"Hello World\".contains(\"Wo\")", res: Bool(true)},
		{exp: "\"Hello World\".contains(\"wo\")", res: Bool(false)},
		{exp: "\"Wo\" ~ \"Hello World\"", res: Bool(true)},
		{exp: "\" Hello \".trim()", res: String("Hello")},
		{exp: "\"Hello,World\".split(\",\").reduce((a,b)->a+\"|\"+b)", res: String("Hello|World")},
		{exp: "\"Hello , World\".split(\",\").reduce((a,b)->a+\"|\"+b)", res: String("Hello | World")},

		{exp: "\"0123456789\".cut(1,3)", res: String("123")},
		{exp: "\"0123456789\".cut(0,3)", res: String("012")},
		{exp: "\"0123456789\".cut(8,2)", res: String("89")},
		{exp: "\"0123456789\".cut(8,6)", res: String("89")},
		{exp: "\"0123456789\".cut(5,0)", res: String("56789")}, // zero takes all runes left
		{exp: "\"01\\nval:23\\n456789\".behind(\"val:\")", res: String("23")},
		{exp: "\"01\\nval:23\".behind(\"val:\")", res: String("23")},
		{exp: "\"12.4\".toFloat()", res: Float(12.4)},
	})
}
