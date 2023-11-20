package listMap

type listMapEntry[V any] struct {
	Key   string
	Value V
}

type ListMap[V any] []listMapEntry[V]

func New[V any](size int) ListMap[V] {
	return make(ListMap[V], 0, size)
}

func NewP[V any](size int) *ListMap[V] {
	lm := make(ListMap[V], 0, size)
	return &lm
}

func (l ListMap[V]) Get(key string) (V, bool) {
	for _, e := range l {
		if e.Key == key {
			return e.Value, true
		}
	}
	var zero V
	return zero, false
}

func (l *ListMap[V]) Put(key string, v V) *ListMap[V] {
	for _, e := range *l {
		if e.Key == key {
			e.Value = v
			return l
		}
	}
	*l = append(*l, listMapEntry[V]{Key: key, Value: v})
	return l
}

func (l ListMap[V]) Iter(yield func(key string, v V) bool) bool {
	for _, e := range l {
		if !yield(e.Key, e.Value) {
			return false
		}
	}
	return true
}

func (l ListMap[V]) Size() int {
	return len(l)
}
