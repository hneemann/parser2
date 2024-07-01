package export

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestMapStyle(t *testing.T) {
	table := value.NewList(value.NewList(value.Int(1), value.Int(2)), value.NewList(value.Int(3), value.Int(4)))
	tests := []struct {
		name        string
		value       value.Value
		maxListSize int
		html        string
	}{
		{"f1", mapStyle(map[string]value.Value{"a": value.String("aa")}, value.String("test")), 10, "<span style=\"a:aa;\">test</span>\n"},
		{"f2", mapStyle(map[string]value.Value{"a": value.String("aa"), "b": value.String("bb")}, value.String("test")), 10, "<span style=\"a:aa;b:bb;\">test</span>\n"},
		{"f3", mapStyle(map[string]value.Value{"b": value.String("bb"), "a": value.String("aa")}, value.String("test")), 10, "<span style=\"a:aa;b:bb;\">test</span>\n"},
		{"f4", mapStyle(map[string]value.Value{"b": value.Int(4), "a": value.Float(1.5)}, value.String("test")), 10, "<span style=\"a:1.5;b:4;\">test</span>\n"},
		{"table1", mapStyle(map[string]value.Value{}, table), 10,
			`<table>
	<tr>
		<td>1</td>
		<td>2</td>
	</tr>
	<tr>
		<td>3</td>
		<td>4</td>
	</tr>
</table>
`},
		{"table2", mapStyle(map[string]value.Value{
			"table": value.NewMap(value.RealMap{
				"r2c2": value.String("r2c2"),
				"c2":   value.String("c2"),
				"r2":   value.String("r2"),
			}),
		}, table), 10,
			`<table>
	<tr>
		<td>1</td>
		<td style="c2">2</td>
	</tr>
	<tr>
		<td style="r2">3</td>
		<td style="r2c2">4</td>
	</tr>
</table>
`},
		{"table2", mapStyle(map[string]value.Value{
			"table": value.NewMap(value.RealMap{
				"all": value.String("all"),
			}),
		}, table), 10,
			`<table>
	<tr>
		<td style="all">1</td>
		<td style="all">2</td>
	</tr>
	<tr>
		<td style="all">3</td>
		<td style="all">4</td>
	</tr>
</table>
`},
		{"table2", mapStyle(map[string]value.Value{
			"table": value.NewMap(value.RealMap{
				"all": value.Closure{
					Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
						return value.String("all"), nil
					},
					Args:   1,
					IsPure: true,
				},
			}),
		}, table), 10,
			`<table>
	<tr>
		<td>all</td>
		<td>all</td>
	</tr>
	<tr>
		<td>all</td>
		<td>all</td>
	</tr>
</table>
`},
		{"table2", mapStyle(map[string]value.Value{
			"table": value.NewMap(value.RealMap{
				"all": value.Closure{
					Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
						r, _ := st.Get(0).ToInt()
						c, _ := st.Get(1).ToInt()
						s, _ := st.Get(2).ToString(st)
						return value.String(strconv.Itoa(r) + "_" + strconv.Itoa(c) + "_" + s), nil
					},
					Args:   3,
					IsPure: true,
				},
			}),
		}, table), 10,
			`<table>
	<tr>
		<td>1_1_1</td>
		<td>1_2_2</td>
	</tr>
	<tr>
		<td>2_1_3</td>
		<td>2_2_4</td>
	</tr>
</table>
`},
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

func mapStyle(s value.RealMap, v value.Value) value.Value {
	st, err := styleFunc.EvalSt(funcGen.NewStack[value.Value](), value.NewMap(s), v)
	if err != nil {
		panic(err)
	}
	return st
}
