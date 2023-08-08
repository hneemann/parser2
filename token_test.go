package parser2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTokenizer(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		want []Token
	}{
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
			name: "string comment",
			exp:  "\"t//b\"",
			want: []Token{{tString, "t//b", 1}},
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
			name: "comment 3",
			exp:  "a-//->test\nb",
			want: []Token{{tIdent, "a", 1}, {tOperate, "-", 1}, {tIdent, "b", 2}},
		},
		{
			name: "comment 4",
			exp:  "a-//->test",
			want: []Token{{tIdent, "a", 1}, {tOperate, "-", 1}},
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
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber, simpleIdentifier, simpleOperator, map[string]string{}, true)
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
			want: []Token{{tIdent, "a", 1}, {tOperate, "//->", 1}, {tIdent, "test", 1}},
		},
		{
			name: "comment 4",
			exp:  "a-//->test",
			want: []Token{{tIdent, "a", 1}, {tOperate, "-//->", 1}, {tIdent, "test", 1}},
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

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber, simpleIdentifier, simpleOperator, map[string]string{}, false)
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}

func TestPeek(t *testing.T) {
	tok := NewTokenizer("=(a,b)", simpleNumber, simpleIdentifier, simpleOperator, map[string]string{}, false)
	assert.Equal(t, "=", tok.Next().image)
	assert.Equal(t, "(", tok.Next().image)
	assert.Equal(t, "a", tok.Peek().image)
	assert.Equal(t, ",", tok.PeekPeek().image)
	assert.Equal(t, "a", tok.Next().image)
	assert.Equal(t, ",", tok.Next().image)
	assert.Equal(t, "b", tok.Next().image)
}
