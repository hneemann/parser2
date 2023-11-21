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
		{name: "empty", a: Map{RealMap{}}, b: Map{RealMap{}}, want: true},
		{name: "one empty", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{}}, want: false},
		{name: "one empty", a: Map{RealMap{}}, b: Map{RealMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{RealMap{"a": Int(2)}}, b: Map{RealMap{"a": Int(1)}}, want: false},
		{name: "not equal", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{"b": Int(1)}}, want: false},
		{name: "equal", a: Map{RealMap{"a": Int(1)}}, b: Map{RealMap{"a": Int(1)}}, want: true},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			assert.Equalf(t, test.want, test.a.Equals(test.b), "Equals(%v)", test.b)
		})
	}
}
