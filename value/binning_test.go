package value

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	b := newBinning(0, 10, 1)
	b.Add(-100, 1)
	b.Add(1, 1)
	b.Add(100, 1)
	var d []bin
	var v []float64
	b.Result(func(desc bin, value float64) bool {
		d = append(d, desc)
		v = append(v, value)
		return true
	})
	assert.EqualValues(t, []bin{{IsMin: false, Min: 0, IsMax: true, Max: 0}, {IsMin: true, Min: 0, IsMax: true, Max: 10}, {IsMin: true, Min: 10, IsMax: false, Max: 0}}, d)
	assert.EqualValues(t, []float64{1, 1, 1}, v)
}

func TestNew2d(t *testing.T) {
	b := New2d(0, 10, 1, 10, 1, 1)
	b.Add(-100, 1, 1)
	b.Add(1, 10.5, 1)
	b.Add(100, 13, 1)
	var dx []bin
	var v [][]float64
	b.Result(func(d bin, y func(func(float64) bool) bool) bool {
		dx = append(dx, d)
		var yv []float64
		y(func(v float64) bool {
			yv = append(yv, v)
			return true
		})
		v = append(v, yv)
		return true
	})
	var dy []bin
	b.DescrY(func(d bin) bool {
		dy = append(dy, d)
		return true
	})
	assert.EqualValues(t, []bin{{IsMin: false, Min: 0, IsMax: true, Max: 0}, {IsMin: true, Min: 0, IsMax: true, Max: 10}, {IsMin: true, Min: 10, IsMax: false, Max: 0}}, dx)
	assert.EqualValues(t, []bin{{IsMin: false, Min: 0, IsMax: true, Max: 10}, {IsMin: true, Min: 10, IsMax: true, Max: 11}, {IsMin: true, Min: 11, IsMax: false, Max: 0}}, dy)
	assert.EqualValues(t, [][]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, v)
}
