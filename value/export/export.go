package export

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"sort"
)

type ListExporter interface {
	Open() error
	Add(item value.Value) error
	Close() error
}

type MapExporter interface {
	Open() error
	Add(key string, val value.Value) error
	Close() error
}

type Exporter[R any] interface {
	String(str string) error
	List() ListExporter
	Map(value.Map) MapExporter
	Custom(val value.Value) (bool, error)
	Result() R
}

func Export[V any](st funcGen.Stack[value.Value], val value.Value, exporter Exporter[V]) error {
	if ok, err := exporter.Custom(val); ok || err != nil {
		if err != nil {
			return err
		}
		return nil
	}
	switch v := val.(type) {
	case Format:
		return Export(st, v.Value, exporter)
	case Link:
		return Export(st, v.Value, exporter)
	case *value.List:
		le := exporter.List()
		err := le.Open()
		if err != nil {
			return err
		}
		err = v.Iterator()(st, func(e value.Value) error {
			return le.Add(e)
		})
		if err != nil {
			return err
		}
		return le.Close()
	case value.Map:
		var keys []string
		v.Iter(func(k string, v value.Value) bool {
			keys = append(keys, k)
			return true
		})
		sort.Strings(keys)
		ma := exporter.Map(v)
		err := ma.Open()
		if err != nil {
			return err
		}
		for _, k := range keys {
			if item, ok := v.Get(k); ok {
				err := ma.Add(k, item)
				if err != nil {
					return err
				}
			}
		}
		return ma.Close()
	default:
		if v == nil {
			return exporter.String("nil")
		}
		str, err := v.ToString(st)
		if err != nil {
			return err
		}
		return exporter.String(str)
	}
}
