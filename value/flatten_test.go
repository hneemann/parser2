package value

import (
	"github.com/hneemann/parser2/listMap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		want []Value
	}{
		{
			name: "simple",
			val:  Int(1),
			want: []Value{Int(1)},
		},
		{
			name: "list",
			val:  NewList(Int(1), Int(2), Int(3)),
			want: []Value{Int(1), Int(2), Int(3)},
		},
		{
			name: "listList",
			val:  NewList(NewList(Int(1), Int(2)), NewList(Int(3), Int(4))),
			want: []Value{Int(1), Int(2), Int(3), Int(4)},
		},
		{
			name: "listMap",
			val: NewList(NewMap(listMap.New[Value](2).Append("a", Int(1)).Append("b", Int(2))),
				NewMap(listMap.New[Value](2).Append("a", Int(3)).Append("b", Int(4)))),
			want: []Value{Int(1), Int(2), Int(3), Int(4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := 0
			for v, err := range Flatten(tt.val) {
				assert.NoError(t, err)
				if i >= len(tt.want) {
					t.Errorf("Flatten() got more values than expected, got %v", v)
					return
				}
				assert.EqualValues(t, tt.want[i], v)
				i++
			}
			assert.Equal(t, len(tt.want), i, "Flatten() got %d values, want %d", i, len(tt.want))

			for n := 1; n < len(tt.want); n++ {
				i := 0
				for v, err := range Flatten(tt.val) {
					assert.NoError(t, err)
					assert.EqualValues(t, tt.want[i], v)
					if i == n {
						break
					} else {
						i++
					}
				}
			}

		})
	}
}
