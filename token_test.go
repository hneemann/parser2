package parser2

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"unicode"
)

func TestNewTokenizer(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		want []Token
	}{
		{
			name: "op1",
			exp:  "+-*/",
			want: []Token{{tOperate, "+", 1}, {tOperate, "-", 1}, {tOperate, "*", 1}, {tOperate, "/", 1}},
		},
		{
			name: "op2",
			exp:  "+->*-+",
			want: []Token{{tOperate, "+", 1}, {tOperate, "->", 1}, {tOperate, "*", 1}, {tOperate, "-", 1}, {tOperate, "+", 1}},
		},
		{
			name: "simple ident",
			exp:  "test",
			want: []Token{{tIdent, "test", 1}},
		},
		{
			name: "ident unicode",
			exp:  "tüb",
			want: []Token{{tIdent, "tüb", 1}},
		},
		{
			name: "ident blank",
			exp:  "a 'A b'",
			want: []Token{{tIdent, "a", 1}, {tIdent, "A b", 1}},
		},
		{
			name: "ident blank unicode",
			exp:  "'tüb'",
			want: []Token{{tIdent, "tüb", 1}},
		},
		{
			name: "ident blank unicode comment",
			exp:  "'t//b'",
			want: []Token{{tIdent, "t//b", 1}},
		},
		{
			name: "string",
			exp:  "\"tüb\"",
			want: []Token{{tString, "tüb", 1}},
		},
		{
			name: "string new line",
			exp:  "\"t\n",
			want: []Token{{tInvalid, "EOL", 1}},
		},
		{
			name: "string comment",
			exp:  "\"t//b\"",
			want: []Token{{tString, "t//b", 1}},
		},
		{
			name: "string escape",
			exp:  "\"t\\\\b\"",
			want: []Token{{tString, "t\\b", 1}},
		},
		{
			name: "string escape 2",
			exp:  "\"t\\n\\r\\tb\"",
			want: []Token{{tString, "t\n\r\tb", 1}},
		},
		{
			name: "string escape 3",
			exp:  "\"\\\"\"",
			want: []Token{{tString, "\"", 1}},
		},
		{
			name: "string escape 4",
			exp:  "\"\\#",
			want: []Token{{tInvalid, "Escape #", 1}},
		},
		{
			name: "exp",
			exp:  "(a\n)",
			want: []Token{{tOpen, "(", 1}, {tIdent, "a", 1}, {tClose, ")", 2}},
		},
		{
			name: "number",
			exp:  "5.5",
			want: []Token{{tNumber, "5.5", 1}},
		},
		{
			name: "comment 1",
			exp:  "a //test\n b",
			want: []Token{{tIdent, "a", 1}, {tIdent, "b", 2}},
		},
		{
			name: "comment 2",
			exp:  "a//->test\nb",
			want: []Token{{tIdent, "a", 1}, {tIdent, "b", 2}},
		},
		{
			name: "comment 5",
			exp:  "a/b",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/", 1}, {tIdent, "b", 1}},
		},
		{
			name: "comment 6",
			exp:  "a/=b",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/=", 1}, {tIdent, "b", 1}},
		},
		{
			name: "comment 7",
			exp:  "a/",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/", 1}},
		},
		{
			name: "comment 8",
			exp:  "a//",
			want: []Token{{tIdent, "a", 1}},
		},
		{
			name: "comment 9",
			exp:  "a//\n",
			want: []Token{{tIdent, "a", 1}},
		},
		{
			name: "comment 10",
			exp:  "a//ss\n//ss\n\na",
			want: []Token{{tIdent, "a", 1}, {tIdent, "a", 4}},
		},
		{
			name: "mod",
			exp:  "a % 10",
			want: []Token{{tIdent, "a", 1}, {tOperate, "%", 1}, {tNumber, "10", 1}},
		},
	}

	detect := NewOperatorDetector([]string{"+", "-", "*", "/", "->", "%", "/="})
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber, simpleIdentifier, detect).SetComments(true).Start()
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}

func TestOperatorDetect(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		op   []string
		want []Token
	}{
		{
			name: "op1",
			exp:  "+--->",
			op:   []string{"+", "--", "->"},
			want: []Token{{tOperate, "+", 1}, {tOperate, "--", 1}, {tOperate, "->", 1}},
		},
		{
			name: "op2",
			exp:  "+-+",
			op:   []string{"+", "--", "->"},
			want: []Token{{tOperate, "+", 1}, {tInvalid, "-", 1}, {tOperate, "+", 1}},
		},
		{
			name: "op3",
			exp:  "+-->",
			op:   []string{"+", "-", "->"},
			want: []Token{{tOperate, "+", 1}, {tOperate, "-", 1}, {tOperate, "->", 1}},
		},
		{
			name: "op3",
			exp:  "+-+---+",
			op:   []string{"+", "-", "---"},
			want: []Token{{tOperate, "+", 1}, {tOperate, "-", 1}, {tOperate, "+", 1}, {tOperate, "---", 1}, {tOperate, "+", 1}},
		},
		{
			name: "op3",
			exp:  "+-+--+",
			op:   []string{"+", "-", "---"},
			want: []Token{{tOperate, "+", 1}, {tOperate, "-", 1}, {tOperate, "+", 1}, {tInvalid, "--", 1}, {tOperate, "+", 1}},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber, simpleIdentifier, NewOperatorDetector(test.op)).Start()
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}

func TestNewTokenizerNoComment(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		want []Token
	}{
		{
			name: "comment 1",
			exp:  "a //test\n b",
			want: []Token{{tIdent, "a", 1}, {tOperate, "//", 1}, {tIdent, "test", 1}, {tIdent, "b", 2}},
		},
		{
			name: "comment 2",
			exp:  "a//->test",
			want: []Token{{tIdent, "a", 1}, {tOperate, "//", 1}, {tOperate, "->", 1}, {tIdent, "test", 1}},
		},
		{
			name: "comment 4",
			exp:  "a-//->test",
			want: []Token{{tIdent, "a", 1}, {tOperate, "-", 1}, {tOperate, "//", 1}, {tOperate, "->", 1}, {tIdent, "test", 1}},
		},
		{
			name: "comment 5",
			exp:  "a/b",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/", 1}, {tIdent, "b", 1}},
		},
		{
			name: "comment 6",
			exp:  "a/=b",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/=", 1}, {tIdent, "b", 1}},
		},
		{
			name: "comment 7",
			exp:  "a/",
			want: []Token{{tIdent, "a", 1}, {tOperate, "/", 1}},
		},
		{
			name: "comment 8",
			exp:  "a//",
			want: []Token{{tIdent, "a", 1}, {tOperate, "//", 1}},
		},
		{
			name: "comment 9",
			exp:  "a//\n",
			want: []Token{{tIdent, "a", 1}, {tOperate, "//", 1}},
		},
	}

	detect := NewOperatorDetector([]string{"//", "->", "/=", "-", "/"})
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber, simpleIdentifier, detect).Start()
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}

func TestPeek(t *testing.T) {
	tok := NewTokenizer("=(a,b)", simpleNumber, simpleIdentifier, NewOperatorDetector([]string{"="})).Start()
	assert.Equal(t, "=", tok.Next().image)
	assert.Equal(t, "(", tok.Next().image)
	assert.Equal(t, "a", tok.Peek().image)
	assert.Equal(t, ",", tok.PeekPeek().image)
	assert.Equal(t, "a", tok.Next().image)
	assert.Equal(t, ",", tok.Next().image)
	assert.Equal(t, "b", tok.Next().image)
}

func TestUnicodeExp(t *testing.T) {
	assert.True(t, unicode.IsNumber('¹'))
	assert.True(t, unicode.IsNumber('²'))
	assert.True(t, unicode.IsNumber('³'))
	assert.True(t, unicode.IsNumber('⁴'))
	assert.True(t, unicode.IsNumber('⁵'))
	assert.True(t, unicode.IsNumber('⁶'))
	assert.False(t, unicode.IsLetter('¹'))
	assert.False(t, unicode.IsLetter('²'))
	assert.False(t, unicode.IsLetter('³'))
	assert.False(t, unicode.IsLetter('⁴'))
	assert.False(t, unicode.IsLetter('⁵'))
	assert.False(t, unicode.IsLetter('⁶'))

	tests := []struct {
		in   string
		want []Token
	}{
		{in: "a¹", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "1", 1}}},
		{in: "a²", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "2", 1}}},
		{in: "a³", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "3", 1}}},
		{in: "a⁴", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "4", 1}}},
		{in: "a⁵", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "5", 1}}},
		{in: "a⁶", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "6", 1}}},
		{in: "a⁷", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "7", 1}}},
		{in: "a⁸", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "8", 1}}},
		{in: "a⁹", want: []Token{{tIdent, "a", 1}, {tOperate, "^", 1}, {tNumber, "9", 1}}},
	}

	myIdent := func(r rune) (func(r rune) bool, bool) {
		if unicode.IsLetter(r) {
			return func(r rune) bool {
				return unicode.IsLetter(r) || (unicode.IsNumber(r) && !strings.ContainsRune("¹²³⁴⁵⁶⁷⁸⁹⁰", r)) || r == '_'
			}, true
		} else {
			return nil, false
		}
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.in, func(t *testing.T) {
			tok := NewTokenizer(test.in, simpleNumber, myIdent, NewOperatorDetector([]string{"^"})).Start()
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}
