package export

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name string
		val  value.Value
		want string
	}{
		{
			name: "nil",
			val:  nil,
			want: "\"nil\"",
		},
		{
			name: "str",
			val:  value.String("test"),
			want: "\"test\"",
		},
		{
			name: "list",
			val:  makeList(),
			want: "[\"a\",\"b\",\"c\"]",
		},
		{
			name: "map",
			val:  makeMap(),
			want: "{\"a\":\"A\",\"b\":\"B\"}",
		},
		{
			name: "list-map",
			val:  value.NewList(makeMap(), makeMap()),
			want: "[{\"a\":\"A\",\"b\":\"B\"},{\"a\":\"A\",\"b\":\"B\"}]",
		},
		{
			name: "map-list",
			val:  value.NewMap(listMap.New[value.Value](2).Append("a", makeList()).Append("b", makeList())),
			want: "{\"a\":[\"a\",\"b\",\"c\"],\"b\":[\"a\",\"b\",\"c\"]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml := JSON()
			err := Export(funcGen.NewEmptyStack[value.Value](), tt.val, xml)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(xml.Result()))
		})
	}
}
