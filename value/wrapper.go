package value

type ToMap[S any] struct {
	attr map[string]func(S) Value
}

func NewToMap[S any]() *ToMap[S] {
	return &ToMap[S]{attr: make(map[string]func(S) Value)}
}

func (wt *ToMap[S]) Attr(name string, val func(S) Value) *ToMap[S] {
	wt.attr[name] = val
	return wt
}

func (wt *ToMap[S]) Create(container S) Map {
	return Map{toMapWrapper[S]{container: container, attr: wt.attr}}
}

type toMapWrapper[S any] struct {
	container S
	attr      map[string]func(S) Value
}

func (w toMapWrapper[S]) Get(key string) (Value, bool) {
	f, ok := w.attr[key]
	if ok {
		return f(w.container), true
	}
	return nil, false
}

func (w toMapWrapper[S]) Iter(yield func(string, Value) bool) bool {
	for k, f := range w.attr {
		if !yield(k, f(w.container)) {
			return false
		}
	}
	return true
}

func (w toMapWrapper[S]) Size() int {
	return len(w.attr)
}
