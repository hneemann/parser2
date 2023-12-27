package value

import (
	"reflect"
)

type ToMapInterface[S any] interface {
	Create(S) Map
}

type funcMap[S any] map[string]func(S) Value

type ToMap[S any] struct {
	attr funcMap[S]
}

func NewToMap[S any]() *ToMap[S] {
	return &ToMap[S]{attr: make(funcMap[S])}
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
	attr      funcMap[S]
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

type ToMapReflection[S any] struct {
	ToMap[reflect.Value]
}

func (wt *ToMapReflection[S]) Create(s S) Map {
	return Map{toMapWrapper[reflect.Value]{container: reflect.ValueOf(s), attr: wt.attr}}
}

func NewToMapReflection[S any]() ToMapInterface[S] {
	var zero S
	t := reflect.TypeOf(zero)
	tm := &ToMapReflection[S]{ToMap[reflect.Value]{attr: make(funcMap[reflect.Value])}}
	for i := 0; i < t.NumField(); i++ {
		i := i
		field := t.Field(i)
		name := field.Name
		switch field.Type.Kind() {
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int:
			tm.Attr(name, func(s reflect.Value) Value { return Int(s.Field(i).Int()) })
		case reflect.Bool:
			tm.Attr(name, func(s reflect.Value) Value { return Bool(s.Field(i).Bool()) })
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			tm.Attr(name, func(s reflect.Value) Value { return Float(s.Field(i).Float()) })
		case reflect.String:
			tm.Attr(name, func(s reflect.Value) Value { return String(s.Field(i).String()) })
		}
	}
	return tm
}
