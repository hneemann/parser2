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
			want: []Token{{tIdent, "test"}},
		},
		{
			name: "ident unicode",
			exp:  "tüb",
			want: []Token{{tIdent, "tüb"}},
		},
		{
			name: "ident blank",
			exp:  "a 'A b'",
			want: []Token{{tIdent, "a"}, {tIdent, "A b"}},
		},
		{
			name: "ident blank unicode",
			exp:  "'tüb'",
			want: []Token{{tIdent, "tüb"}},
		},
		{
			name: "ident blank unicode comment",
			exp:  "'t//b'",
			want: []Token{{tIdent, "t//b"}},
		},
		{
			name: "string",
			exp:  "\"tüb\"",
			want: []Token{{tString, "tüb"}},
		},
		{
			name: "string comment",
			exp:  "\"t//b\"",
			want: []Token{{tString, "t//b"}},
		},
		{
			name: "exp",
			exp:  "(a)",
			want: []Token{{tOpen, "("}, {tIdent, "a"}, {tClose, ")"}},
		},
		{
			name: "number",
			exp:  "5.5",
			want: []Token{{tNumber, "5.5"}},
		},
		{
			name: "comment 1",
			exp:  "a //test\n b",
			want: []Token{{tIdent, "a"}, {tIdent, "b"}},
		},
		{
			name: "comment 2",
			exp:  "a//->test\nb",
			want: []Token{{tIdent, "a"}, {tIdent, "b"}},
		},
		{
			name: "comment 3",
			exp:  "a-//->test\nb",
			want: []Token{{tIdent, "a"}, {tOperate, "-"}, {tIdent, "b"}},
		},
		{
			name: "comment 4",
			exp:  "a-//->test",
			want: []Token{{tIdent, "a"}, {tOperate, "-"}},
		},
		{
			name: "comment 5",
			exp:  "a/b",
			want: []Token{{tIdent, "a"}, {tOperate, "/"}, {tIdent, "b"}},
		},
		{
			name: "comment 6",
			exp:  "a/=b",
			want: []Token{{tIdent, "a"}, {tOperate, "/="}, {tIdent, "b"}},
		},
		{
			name: "comment 7",
			exp:  "a/",
			want: []Token{{tIdent, "a"}, {tOperate, "/"}},
		},
		{
			name: "comment 8",
			exp:  "a//",
			want: []Token{{tIdent, "a"}},
		},
		{
			name: "comment 9",
			exp:  "a//\n",
			want: []Token{{tIdent, "a"}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber{}, simpleIdentifier{}, simpleOperator{}, map[string]string{}, true)
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
			want: []Token{{tIdent, "a"}, {tOperate, "//"}, {tIdent, "test"}, {tIdent, "b"}},
		},
		{
			name: "comment 2",
			exp:  "a//->test",
			want: []Token{{tIdent, "a"}, {tOperate, "//->"}, {tIdent, "test"}},
		},
		{
			name: "comment 4",
			exp:  "a-//->test",
			want: []Token{{tIdent, "a"}, {tOperate, "-//->"}, {tIdent, "test"}},
		},
		{
			name: "comment 5",
			exp:  "a/b",
			want: []Token{{tIdent, "a"}, {tOperate, "/"}, {tIdent, "b"}},
		},
		{
			name: "comment 6",
			exp:  "a/=b",
			want: []Token{{tIdent, "a"}, {tOperate, "/="}, {tIdent, "b"}},
		},
		{
			name: "comment 7",
			exp:  "a/",
			want: []Token{{tIdent, "a"}, {tOperate, "/"}},
		},
		{
			name: "comment 8",
			exp:  "a//",
			want: []Token{{tIdent, "a"}, {tOperate, "//"}},
		},
		{
			name: "comment 9",
			exp:  "a//\n",
			want: []Token{{tIdent, "a"}, {tOperate, "//"}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tok := NewTokenizer(test.exp, simpleNumber{}, simpleIdentifier{}, simpleOperator{}, map[string]string{}, false)
			for _, to := range test.want {
				assert.EqualValues(t, to, tok.Next())
			}
			assert.EqualValues(t, TokenEof, tok.Next())
			assert.EqualValues(t, TokenEof, tok.Next())
		})
	}
}
