package export

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXML(t *testing.T) {
	tests := []struct {
		name string
		val  value.Value
		want string
	}{
		{
			name: "nil",
			val:  nil,
			want: "nil",
		},
		{
			name: "str",
			val:  value.String("test"),
			want: "test",
		},
		{
			name: "list",
			val:  makeList(),
			want: "<list>\n\t<entry>a</entry>\n\t<entry>b</entry>\n\t<entry>c</entry>\n</list>\n",
		},
		{
			name: "map",
			val:  makeMap(),
			want: "<map a=\"A\" b=\"B\"/>\n",
		},
		{
			name: "list-map",
			val:  value.NewList(makeMap(), makeMap()),
			want: "<list>\n\t<entry>\n\t\t<map a=\"A\" b=\"B\"/>\n\t</entry>\n\t<entry>\n\t\t<map a=\"A\" b=\"B\"/>\n\t</entry>\n</list>\n",
		},
		{
			name: "map-list",
			val:  value.NewMap(listMap.New[value.Value](2).Append("a", makeList()).Append("b", makeList())),
			want: "<map>\n\t<entry key=\"a\">\n\t\t<list>\n\t\t\t<entry>a</entry>\n\t\t\t<entry>b</entry>\n\t\t\t<entry>c</entry>\n\t\t</list>\n\t</entry>\n\t<entry key=\"b\">\n\t\t<list>\n\t\t\t<entry>a</entry>\n\t\t\t<entry>b</entry>\n\t\t\t<entry>c</entry>\n\t\t</list>\n\t</entry>\n</map>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml := XML()
			err := Export(funcGen.NewEmptyStack[value.Value](), tt.val, xml)
			assert.NoError(t, err)
			assert.Equal(t, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\" ?>\n"+tt.want, string(xml.Result()))
		})
	}
}

func makeList() *value.List {
	return value.NewList(value.String("a"), value.String("b"), value.String("c"))
}

func makeMap() value.Map {
	return value.NewMap(listMap.New[value.Value](2).Append("a", value.String("A")).Append("b", value.String("B")))
}
