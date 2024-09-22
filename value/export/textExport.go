package export

import (
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"io"
	"sort"
)

type textExporter struct {
	w       io.Writer
	spaces  string
	newline bool
}

func NewTextExporter(w io.Writer) *textExporter {
	return &textExporter{w: w}
}

func (e *textExporter) write(s string) {
	if e.newline {
		e.w.Write([]byte(e.spaces))
		e.newline = false
	}
	e.w.Write([]byte(s))
}

func (e *textExporter) newLine() {
	e.w.Write([]byte("\n"))
	e.newline = true
}

func (e *textExporter) inc() {
	e.spaces += "  "
}

func (e *textExporter) dec() {
	if len(e.spaces) > 2 {
		e.spaces = e.spaces[:len(e.spaces)-2]
	} else {
		e.spaces = ""
	}
}

func (e *textExporter) ToText(v value.Value) error {
	return e.toText(funcGen.NewEmptyStack[value.Value](), v)
}

func (e *textExporter) toText(st funcGen.Stack[value.Value], v value.Value) error {
	switch t := v.(type) {
	case *value.List:
		e.write("[")
		e.newLine()
		e.inc()
		so := sepOut{e: e}
		err := t.Iterate(st, func(v value.Value) error {
			return so.out(v, st)
		})
		err = so.finish(st, err)
		e.dec()
		e.write("]")
		if err == iterator.SBC {
			return nil
		}
		return err
	case value.Map:
		var keys []string
		t.Iter(func(k string, v value.Value) bool {
			keys = append(keys, k)
			return true
		})
		sort.Strings(keys)
		e.write("{")
		e.newLine()
		e.inc()
		for i, k := range keys {
			e.write(k)
			e.write(": ")
			v, _ := t.Get(k)
			err := e.toText(st, v)
			if err != nil {
				return err
			}
			if i < len(keys)-1 {
				e.write(",")
			}
			e.newLine()
		}
		e.dec()
		e.write("}")
	default:
		if v == nil {
			e.write("nil")
		} else {
			s, err := v.ToString(st)
			if err != nil {
				return err
			}
			e.write(s)
		}
	}
	return nil
}

type sepOut struct {
	e    *textExporter
	last value.Value
}

func (so *sepOut) out(v value.Value, st funcGen.Stack[value.Value]) error {
	if so.last != nil {
		err := so.e.toText(st, so.last)
		so.e.write(",")
		so.e.newLine()
		so.last = v
		return err
	} else {
		so.last = v
		return nil
	}
}

func (so *sepOut) finish(st funcGen.Stack[value.Value], err error) error {
	if err != nil && err != iterator.SBC {
		return err
	}
	if so.last != nil {
		err := so.e.toText(st, so.last)
		so.e.newLine()
		return err
	}
	return err
}

func (e *textExporter) sepOut(v value.Value, last value.Value, st funcGen.Stack[value.Value]) error {
	if last != nil {
		err := e.toText(st, last)
		e.write(",")
		e.newLine()
		return err
	} else {
		return nil
	}
}
