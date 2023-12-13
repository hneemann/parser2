package export

import (
	"bytes"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export/xmlWriter"
)

type xmlListExporter struct {
	x xmlExporter
}

func (x xmlListExporter) Open() error {
	x.x.w.Open("list")
	return nil
}

func (x xmlListExporter) Add(item value.Value) error {
	x.x.w.Open("entry")
	err := Export[[]byte](funcGen.NewEmptyStack[value.Value](), item, x.x)
	if err != nil {
		return err
	}
	x.x.w.Close()
	return nil
}

func (x xmlListExporter) Close() error {
	x.x.w.Close()
	return nil
}

type xmlMapExporter struct {
	x        xmlExporter
	isSimple bool
}

func (x xmlMapExporter) Open() error {
	x.x.w.Open("map")
	return nil
}

func (x xmlMapExporter) Add(key string, val value.Value) error {
	st := funcGen.NewEmptyStack[value.Value]()
	if x.isSimple {
		str, err := val.ToString(st)
		if err != nil {
			return err
		}
		x.x.w.Attr(key, str)
	} else {
		x.x.w.Open("entry").Attr("key", key)
		err := Export[[]byte](st, val, x.x)
		if err != nil {
			return err
		}
		x.x.w.Close()
	}
	return nil
}

func (x xmlMapExporter) Close() error {
	x.x.w.Close()
	return nil
}

type xmlExporter struct {
	w *xmlWriter.XMLWriter
}

func (x xmlExporter) Result() []byte {
	return x.w.Bytes()
}

func (x xmlExporter) String(str string) error {
	x.w.Write(str)
	return nil
}

func (x xmlExporter) List() ListExporter {
	return &xmlListExporter{x: x}
}

func (x xmlExporter) Map(m value.Map) MapExporter {
	return &xmlMapExporter{x: x, isSimple: isSimpleMap(m)}
}

func isSimpleMap(m value.Map) bool {
	isSimple := true
	m.Iter(func(key string, e value.Value) bool {
		if _, ok := e.ToMap(); ok {
			isSimple = false
		}
		if _, ok := e.ToList(); ok {
			isSimple = false
		}
		if _, ok := e.(Format); ok {
			isSimple = false
		}
		return true
	})
	return isSimple
}

func (x xmlExporter) Custom(value.Value) (bool, error) {
	return false, nil
}

func XML() Exporter[[]byte] {
	var b bytes.Buffer
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\" ?>\n")
	return &xmlExporter{
		w: xmlWriter.NewWithBuffer(&b),
	}
}
