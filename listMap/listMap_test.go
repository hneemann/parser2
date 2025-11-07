package listMap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListReplace(t *testing.T) {
	m := New[int](1)
	m = m.Append("a", 1)
	assert.Equal(t, 1, m.Size())
	exp, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, exp)

	m = m.Append("a", 4)
	assert.Equal(t, 1, m.Size())
	exp, ok = m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 4, exp)
}

func TestListGetFails(t *testing.T) {
	m := New[int](1)
	m = m.Append("a", 1)
	assert.Equal(t, 1, m.Size())
	exp, ok := m.Get("b")
	assert.False(t, ok)
	assert.Equal(t, 0, exp)
}

func TestNilAppend(t *testing.T) {
	var m ListMap[int]
	m = m.Append("a", 1)
	assert.Equal(t, 1, m.Size())
	exp, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, exp)
}

func TestIter(t *testing.T) {
	m := New[int](1).
		Append("a", 1).
		Append("b", 2).
		Append("c", 3)

	assert.Equal(t, 3, m.Size())

	var sum int
	for _, v := range m.Iter {
		sum += v
	}
	assert.Equal(t, 6, sum)
}

func TestIterBreak(t *testing.T) {
	m := New[int](1).
		Append("a", 1).
		Append("b", 2).
		Append("c", 3)

	assert.Equal(t, 3, m.Size())

	var sum int
	for _, v := range m.Iter {
		sum += v
		break
	}
	assert.Equal(t, 1, sum)
}
