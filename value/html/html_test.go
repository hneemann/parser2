package html

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestToHtml(t *testing.T) {
	tests := []struct {
		name        string
		value       value.Value
		maxListSize int
		html        template.HTML
	}{
		{"nil", nil, 10, "nil"},
		{"int", value.Int(5), 10, "5"},
		{"bool", value.Bool(true), 10, "true"},
		{"bool", value.Bool(false), 10, "false"},
		{"str", value.String("test"), 10, "test"},
		{"http", value.String("http://a/b.html"), 10, "<a href=\"http://a/b.html\" target=\"_blank\">Link</a>\n"},
		{"host", value.String("host:/a/b.html"), 10, "<a href=\"/a/b.html\" target=\"_blank\">Link</a>\n"},
		{"list", value.NewList(value.Int(4), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"table", value.NewList(value.NewList(value.Int(1), value.Int(2)), value.NewList(value.Int(3), value.Int(4))), 10, "<table>\n\t<tr>\n\t\t<td>1</td>\n\t\t<td>2</td>\n\t</tr>\n\t<tr>\n\t\t<td>3</td>\n\t\t<td>4</td>\n\t</tr>\n</table>\n"},
		{"map", value.Map{M: listMap.NewP[value.Value](2).Put("a", value.Int(1)).Put("b", value.Int(2))}, 10, "<table>\n\t<tr>\n\t\t<td>a:</td>\n\t\t<td>1</td>\n\t</tr>\n\t<tr>\n\t\t<td>b:</td>\n\t\t<td>2</td>\n\t</tr>\n</table>\n"},

		{"f1", style(value.String("test")), 10, "<span style=\"zzz\">test</span>\n"},
		{"f2", style(value.NewList(value.Int(4), value.Int(5))), 10, "<table style=\"zzz\">\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f3", value.NewList(style(value.Int(4)), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td style=\"zzz\">4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f41", value.NewList(style(value.NewList(value.Int(4), value.Int(5))), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>\n\t\t\t<table style=\"zzz\">\n\t\t\t\t<tr>\n\t\t\t\t\t<td>1.</td>\n\t\t\t\t\t<td>4</td>\n\t\t\t\t</tr>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>2.</td>\n\t\t\t\t\t<td>5</td>\n\t\t\t\t</tr>\n\t\t\t</table>\n\t\t</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f42", value.NewList(styleCell(value.NewList(value.Int(4), value.Int(5))), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td style=\"zzz\">\n\t\t\t<table>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>1.</td>\n\t\t\t\t\t<td>4</td>\n\t\t\t\t</tr>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>2.</td>\n\t\t\t\t\t<td>5</td>\n\t\t\t\t</tr>\n\t\t\t</table>\n\t\t</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},

		{"list1", value.NewList(value.Int(4), value.Int(5)), 1, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>more...</td>\n\t</tr>\n</table>\n"},
		{"table1", value.NewList(value.NewList(value.Int(1), value.Int(2)), value.NewList(value.Int(3), value.Int(4))), 1, "<table>\n\t<tr>\n\t\t<td>1</td>\n\t\t<td>more...</td>\n\t</tr>\n\t<tr>\n\t\t<td>more...</td>\n\t</tr>\n</table>\n"},
	}
	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, test.html, ToHtml(test.value, test.maxListSize))
		})
	}
}

func style(v value.Value) value.Value {
	return StyleFunc.EvalSt(funcGen.NewStack[value.Value](), value.String("zzz"), v)
}

func styleCell(v value.Value) value.Value {
	return StyleFuncCell.EvalSt(funcGen.NewStack[value.Value](), value.String("zzz"), v)
}

func TestFormat_GetMethod(t *testing.T) {
	v := style(value.String("test"))
	m, err := v.GetMethod("len")
	assert.NoError(t, err)
	got := m.Func(funcGen.NewStack(v), nil)
	assert.Equal(t, value.Int(4), got)
}
