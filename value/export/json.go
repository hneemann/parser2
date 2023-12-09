package export

import (
	"bytes"
	"github.com/hneemann/parser2/value"
)

type jsonListExporter struct {
	j     jsonExporter
	first bool
}

func (j *jsonListExporter) Open() error {
	_, err := j.j.b.WriteString("[")
	return err
}

func (j *jsonListExporter) Add(item value.Value) error {
	if j.first {
		j.first = false
	} else {
		_, err := j.j.b.WriteString(",")
		if err != nil {
			return err
		}
	}
	return Export[[]byte](item, j.j)
}

func (j *jsonListExporter) Close() error {
	_, err := j.j.b.WriteString("]")
	return err
}

type jsonMapExporter struct {
	j     jsonExporter
	first bool
}

func (j *jsonMapExporter) Open() error {
	_, err := j.j.b.WriteString("{")
	return err
}

func (j *jsonMapExporter) Add(key string, val value.Value) error {
	if j.first {
		j.first = false
	} else {
		_, err := j.j.b.WriteString(",")
		if err != nil {
			return err
		}
	}
	err := j.j.String(key)
	if err != nil {
		return err
	}
	j.j.b.WriteString(":")
	return Export[[]byte](val, j.j)
}

func (j *jsonMapExporter) Close() error {
	_, err := j.j.b.WriteString("}")
	return err
}

type jsonExporter struct {
	b *bytes.Buffer
}

func (j jsonExporter) String(str string) error {
	j.b.WriteString("\"")
	for _, r := range str {
		switch r {
		case '"':
			j.b.WriteString("\\\"")
		case '\t':
			j.b.WriteString("\\t")
		case '\r':
			j.b.WriteString("\\r")
		case '\n':
			j.b.WriteString("\\n")
		default:
			j.b.WriteRune(r)
		}
	}
	j.b.WriteString("\"")
	return nil
}

func (j jsonExporter) List() ListExporter {
	return &jsonListExporter{j: j, first: true}
}

func (j jsonExporter) Map(value.Map) MapExporter {
	return &jsonMapExporter{j: j, first: true}
}

func (j jsonExporter) Custom(val value.Value) (bool, error) {
	return false, nil
}

func (j jsonExporter) Result() []byte {
	return j.b.Bytes()
}

func JSON() Exporter[[]byte] {
	var b bytes.Buffer
	return &jsonExporter{b: &b}
}
