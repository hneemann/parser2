// Package listMap implements a map based on a list for storage. Although the
// access to such a list is of order O(n), it is more performant than a map if
// the number of elements is small, as the overhead when accessing a map is
// comparatively large and is therefore only worthwhile if many elements are
// stored in the map. This implementation should no longer be used if there are
// more than around 20 elements.
package listMap

type listMapEntry[V any] struct {
	key   string
	value V
}

type ListMap[V any] []listMapEntry[V]

func New[V any](size int) ListMap[V] {
	return make(ListMap[V], 0, size)
}

func (l ListMap[V]) Get(key string) (V, bool) {
	for _, e := range l {
		if e.key == key {
			return e.value, true
		}
	}
	var zero V
	return zero, false
}

func (l ListMap[V]) Append(key string, v V) ListMap[V] {
	for i, e := range l {
		if e.key == key {
			l[i].value = v
			return l
		}
	}
	return append(l, listMapEntry[V]{key: key, value: v})
}

func (l ListMap[V]) Iter(yield func(key string, v V) bool) {
	for _, e := range l {
		if !yield(e.key, e.value) {
			return
		}
	}
}

func (l ListMap[V]) Size() int {
	return len(l)
}
