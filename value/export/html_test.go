package export

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"math"
	"strconv"
	"testing"
)

func TestToHtml(t *testing.T) {
	tests := []struct {
		name        string
		value       value.Value
		maxListSize int
		html        string
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
		{"map", value.NewMap(listMap.New[value.Value](2).Append("a", value.Int(1)).Append("b", value.Int(2))), 10, "<table>\n\t<tr>\n\t\t<td>a:</td>\n\t\t<td>1</td>\n\t</tr>\n\t<tr>\n\t\t<td>b:</td>\n\t\t<td>2</td>\n\t</tr>\n</table>\n"},

		{"f1", style("zzz", value.String("test")), 10, "<span style=\"zzz\">test</span>\n"},
		{"f2", style("zzz", value.NewList(value.Int(4), value.Int(5))), 10, "<table style=\"zzz\">\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f3", value.NewList(style("zzz", value.Int(4)), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td style=\"zzz\">4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f41", value.NewList(style("zzz", value.NewList(value.Int(4), value.Int(5))), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>\n\t\t\t<table style=\"zzz\">\n\t\t\t\t<tr>\n\t\t\t\t\t<td>1.</td>\n\t\t\t\t\t<td>4</td>\n\t\t\t\t</tr>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>2.</td>\n\t\t\t\t\t<td>5</td>\n\t\t\t\t</tr>\n\t\t\t</table>\n\t\t</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},
		{"f42", value.NewList(styleCell(value.NewList(value.Int(4), value.Int(5))), value.Int(5)), 10, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td style=\"zzz\">\n\t\t\t<table>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>1.</td>\n\t\t\t\t\t<td>4</td>\n\t\t\t\t</tr>\n\t\t\t\t<tr>\n\t\t\t\t\t<td>2.</td>\n\t\t\t\t\t<td>5</td>\n\t\t\t\t</tr>\n\t\t\t</table>\n\t\t</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>5</td>\n\t</tr>\n</table>\n"},

		{"list1", value.NewList(value.Int(4), value.Int(5)), 1, "<table>\n\t<tr>\n\t\t<td>1.</td>\n\t\t<td>4</td>\n\t</tr>\n\t<tr>\n\t\t<td>2.</td>\n\t\t<td>more...</td>\n\t</tr>\n</table>\n"},
		{"table1", value.NewList(value.NewList(value.Int(1), value.Int(2)), value.NewList(value.Int(3), value.Int(4))), 1, "<table>\n\t<tr>\n\t\t<td>1</td>\n\t\t<td>more...</td>\n\t</tr>\n\t<tr>\n\t\t<td>more...</td>\n\t</tr>\n</table>\n"},

		{"link", link(value.Int(12)), 1, "<a href=\"link\">12</a>\n"},
		{"link", link(value.NewList(value.Int(1), value.Int(2))), 1, "<a href=\"link\">\n\t<table>\n\t\t<tr>\n\t\t\t<td>1.</td>\n\t\t\t<td>1</td>\n\t\t</tr>\n\t\t<tr>\n\t\t\t<td>2.</td>\n\t\t\t<td>more...</td>\n\t\t</tr>\n\t</table>\n</a>\n"},

		{"plainList", style("plainList", value.NewList(value.Int(1), value.Int(2))), 1, "12"},
		{"plainList", style("plainList", value.NewList(value.Int(1), link(value.String("inner")), value.Int(2))), 1, "1\n<a href=\"link\">inner</a>\n2"},
	}
	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			h, _, err := ToHtml(test.value, test.maxListSize, nil, true)
			assert.NoError(t, err)
			assert.Equal(t, test.html, string(h))
		})
	}
}

func link(v value.Value) value.Value {
	st, err := linkFunc.EvalSt(funcGen.NewStack[value.Value](), value.String("link"), v)
	if err != nil {
		panic(err)
	}
	return st
}

func style(s string, v value.Value) value.Value {
	st, err := styleFunc.EvalSt(funcGen.NewStack[value.Value](), value.String(s), v)
	if err != nil {
		panic(err)
	}
	return st
}

func styleCell(v value.Value) value.Value {
	st, err := styleFuncCell.EvalSt(funcGen.NewStack[value.Value](), value.String("zzz"), v)
	if err != nil {
		panic(err)
	}
	return st
}

func Test_byteSize_String(t *testing.T) {
	tests := []struct {
		name string
		b    byteSize
		want string
	}{
		{"1", byteSize(1), "1 Bytes"},
		{"2", byteSize(400), "400 Bytes"},
		{"3", byteSize(600), "600 Bytes"},
		{"4", byteSize(2400), "2400 Bytes"},
		{"5", byteSize(20400), "19 kBytes"},
		{"6", byteSize(20400000), "19 MBytes"},
		{"7", byteSize(1e12), "931 GBytes"},
		{"8", byteSize(1e17), "90949 TBytes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.b.String(), "String()")
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		v       float64
		unicode string
		ascii   string
		latex   string
	}{
		{0, "0", "0", "0"},
		{1e0, "1", "1", "1"},
		{1e1, "10", "10", "10"},
		{1e2, "100", "100", "100"},
		{1e3, "1000", "1000", "1000"},
		{1e4, "10000", "10000", "10000"},
		{1e5, "100000", "100000", "100000"},
		{1e6, "10⁶", "10^6", "10^{6}"},
		{1e7, "10⁷", "10^7", "10^{7}"},
		{1e8, "10⁸", "10^8", "10^{8}"},
		{1e9, "10⁹", "10^9", "10^{9}"},
		{1e10, "10¹⁰", "10^10", "10^{10}"},
		{1e11, "10¹¹", "10^11", "10^{11}"},
		{3e0, "3", "3", "3"},
		{3e1, "30", "30", "30"},
		{3e2, "300", "300", "300"},
		{3e3, "3000", "3000", "3000"},
		{3e4, "30000", "30000", "30000"},
		{3e5, "300000", "300000", "300000"},
		{3e6, "3⋅10⁶", "3*10^6", "3\\cdot 10^{6}"},
		{3e7, "3⋅10⁷", "3*10^7", "3\\cdot 10^{7}"},
		{3e8, "3⋅10⁸", "3*10^8", "3\\cdot 10^{8}"},
		{3e9, "3⋅10⁹", "3*10^9", "3\\cdot 10^{9}"},
		{3e10, "3⋅10¹⁰", "3*10^10", "3\\cdot 10^{10}"},
		{3e11, "3⋅10¹¹", "3*10^11", "3\\cdot 10^{11}"},
		{math.Pi * 1e7, "3.14159⋅10⁷", "3.14159*10^7", "3.14159\\cdot 10^{7}"},
		{-3e3, "-3000", "-3000", "-3000"},
		{-3e7, "-3⋅10⁷", "-3*10^7", "-3\\cdot 10^{7}"},
		{3e-3, "0.003", "0.003", "0.003"},
		{5e-7, "5⋅10⁻⁷", "5*10^-7", "5\\cdot 10^{-7}"},
		{1e-1, "0.1", "0.1", "0.1"},
		{1e-2, "0.01", "0.01", "0.01"},
		{1e-3, "0.001", "0.001", "0.001"},
		{1e-4, "0.0001", "0.0001", "0.0001"},
		{1e-5, "10⁻⁵", "10^-5", "10^{-5}"},
		{3.2e-5, "3.2⋅10⁻⁵", "3.2*10^-5", "3.2\\cdot 10^{-5}"},
		{1e-6, "10⁻⁶", "10^-6", "10^{-6}"},
		{1e-7, "10⁻⁷", "10^-7", "10^{-7}"},
		{1e-8, "10⁻⁸", "10^-8", "10^{-8}"},
		{1e-11, "10⁻¹¹", "10^-11", "10^{-11}"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.v), func(t *testing.T) {
			ff := NewFormattedFloat(tt.v, 6)
			assert.Equalf(t, tt.unicode, ff.Unicode(), "FormatFloat(%v)", tt.v)
			assert.Equalf(t, tt.ascii, ff.Ascii(), "FormatFloat(%v)", tt.v)
			assert.Equalf(t, tt.latex, ff.LaTeX(), "FormatFloat(%v)", tt.v)
		})
	}
}

func TestExpStr(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{n: 0, want: "⁰"},
		{n: 1, want: "¹"},
		{n: -1, want: "⁻¹"},
		{n: 4368231579, want: "⁴³⁶⁸²³¹⁵⁷⁹"},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.n), func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpStr(tt.n), "ExpStr(%v)", tt.n)
		})
	}
}
