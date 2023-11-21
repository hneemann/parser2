package value

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap_Equals(t *testing.T) {
	tests := []struct {
		name string
		a, b Map
		want bool
	}{
		{name: "empty", a: Map{SimpleMap{}}, b: Map{SimpleMap{}}, want: true},
		{name: "one empty", a: Map{SimpleMap{"a": Int(1)}}, b: Map{SimpleMap{}}, want: false},
		{name: "one empty", a: Map{SimpleMap{}}, b: Map{SimpleMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{SimpleMap{"a": Int(2)}}, b: Map{SimpleMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{SimpleMap{"a": Int(1)}}, b: Map{SimpleMap{"b": Int(1)}}, want: false},
		{name: "equal", a: Map{SimpleMap{"a": Int(1)}}, b: Map{SimpleMap{"a": Int(1)}}, want: true},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			assert.Equalf(t, test.want, test.a.Equals(test.b), "Equals(%v)", test.b)
		})
	}
}
