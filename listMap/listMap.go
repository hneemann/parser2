package listMap

type listMapEntry[V any] struct {
	Key   string
	Value V
}

type ListMap[V any] []listMapEntry[V]

func New[V any](size int) ListMap[V] {
	return make(ListMap[V], 0, size)
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

func (l *ListMap[V]) Put(key string, v V) {
	for _, e := range *l {
		if e.Key == key {
			e.Value = v
			return
		}
	}
	*l = append(*l, listMapEntry[V]{Key: key, Value: v})
}
