package export

import (
	"bytes"
	"github.com/hneemann/parser2/listMap"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToText(t *testing.T) {
	tests := []struct {
		name  string
		value value.Value
		text  string
	}{
		{"nil", nil, "nil"},
		{"int", value.Int(5), "5"},
		{"bool", value.Bool(true), "true"},
		{"bool", value.Bool(false), "false"},
		{"str", value.String("test"), "test"},
		{"list", value.NewList(value.Int(4), value.Int(5)), "[\n  4,\n  5,\n]"},
		{"table", value.NewList(value.NewList(value.Int(1), value.Int(2)), value.NewList(value.Int(3), value.Int(4))), "[\n  [\n    1,\n    2,\n  ],\n  [\n    3,\n    4,\n  ],\n]"},
		{"map", value.NewMap(listMap.New[value.Value](2).Append("a", value.Int(1)).Append("b", value.Int(2))), "{\n  a: 1,\n  b: 2,\n}"},
	}
	for _, tt := range tests {
		test := tt
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			e := NewTextExporter(&b)
			err := e.ToText(test.value)
			assert.NoError(t, err)
			assert.Equal(t, test.text, b.String())
		})
	}

}
