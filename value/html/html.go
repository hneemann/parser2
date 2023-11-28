// Package html is used to create an HTML representation of a value
// It's ToHtml function returns a sanitized html string.
package html

import (
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/html/xmlWriter"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

type format struct {
	Value  value.Value
	Cell   bool
	Format string
}

// StyleFunc can be used add a CSS style to a value
var StyleFunc = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) value.Value {
		return format{
			Value:  st.Get(1),
			Cell:   false,
			Format: st.Get(0).String(),
		}
	},
	Args:   2,
	IsPure: true,
}

// StyleFuncCell can be used add a CSS stale to a value
// If used in a table the format is applied to the cell instead of the containing value.
// It is only required in rare occasions.
var StyleFuncCell = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) value.Value {
		return format{
			Value:  st.Get(1),
			Cell:   true,
			Format: st.Get(0).String(),
		}
	},
	Args:   2,
	IsPure: true,
}

func (f format) ToList() (*value.List, bool) {
	return f.Value.ToList()
}

func (f format) ToMap() (value.Map, bool) {
	return f.Value.ToMap()
}

func (f format) ToInt() (int, bool) {
	return f.Value.ToInt()
}

func (f format) ToFloat() (float64, bool) {
	return f.Value.ToFloat()
}

func (f format) String() string {
	return f.Value.String()
}

func (f format) ToBool() (bool, bool) {
	return f.Value.ToBool()
}

func (f format) ToClosure() (funcGen.Function[value.Value], bool) {
	return f.Value.ToClosure()
}

func (f format) GetMethod(name string) (funcGen.Function[value.Value], error) {
	m, err := f.Value.GetMethod(name)
	if err != nil {
		return funcGen.Function[value.Value]{}, err
	}
	return funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) value.Value {
			ss := st.Size()
			st.Push((st.Get(0).(format)).Value)
			for i := 1; i < ss; i++ {
				st.Push(st.Get(i))
			}
			return m.Func(st.CreateFrame(ss), closureStore)
		},
		Args:   m.Args,
		IsPure: m.IsPure,
	}, nil
}

// ToHtml creates an HTML representation of a value
// Lists and maps are converted to a html table.
// Everything else is converted to a string by calling the String() method.
func ToHtml(v value.Value, maxListSize int) (res template.HTML, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			if e, ok := rec.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("error: %v", rec)
			}
			res = ""
		}
	}()
	if maxListSize < 1 {
		maxListSize = 1
	}
	w := xmlWriter.New()
	toHtml(v, w, "", maxListSize)
	return template.HTML(w.String()), nil
}

func toHtml(v value.Value, w *xmlWriter.XMLWriter, style string, maxListSize int) {
	switch t := v.(type) {
	case format:
		toHtml(t.Value, w, t.Format, maxListSize)
	case *value.List:
		pit, f, ok := iterator.Peek(t.Iterator())
		if ok {
			if _, ok := f.(*value.List); ok {
				tableToHtml(pit, w, style, maxListSize)
				return
			}
		}
		listToHtml(pit, w, style, maxListSize)
	case value.Map:
		openWithStyle("table", style, w)
		var keys []string
		t.Iter(func(k string, v value.Value) bool {
			keys = append(keys, k)
			return true
		})
		sort.Strings(keys)
		for _, k := range keys {
			w.Open("tr")
			w.Open("td")
			w.Write(k)
			w.Write(":")
			w.Close()
			v, _ := t.Get(k)
			toTD(v, w, maxListSize)
			w.Close()
		}
		w.Close()
	default:
		if v == nil {
			w.Write("nil")
		} else {
			writeHtmlString(v.String(), style, w)
		}
	}
}

func writeHtmlString(s string, style string, w *xmlWriter.XMLWriter) {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		w.Open("a").Attr("href", s).Attr("target", "_blank")
		w.Write("Link")
		w.Close()
	} else if strings.HasPrefix(s, "host:") {
		w.Open("a").Attr("href", s[5:]).Attr("target", "_blank")
		w.Write("Link")
		w.Close()
	} else {
		if style == "" {
			w.Write(s)
		} else {
			w.Open("span").Attr("style", style)
			w.Write(s)
			w.Close()
		}
	}
}

func listToHtml(it iterator.Iterator[value.Value], w *xmlWriter.XMLWriter, style string, maxListSize int) {
	openWithStyle("table", style, w)
	i := 0
	it(func(e value.Value) bool {
		i++
		w.Open("tr")
		w.Open("td").Write(strconv.Itoa(i)).Write(".").Close()
		if i <= maxListSize {
			toTD(e, w, maxListSize)
		} else {
			w.Open("td").Write("more...").Close()
		}
		w.Close()
		return i <= maxListSize
	})
	w.Close()
}

func openWithStyle(tag string, style string, w *xmlWriter.XMLWriter) {
	if style == "" {
		w.Open(tag)
	} else {
		w.Open(tag).Attr("style", style)
	}
}

func tableToHtml(it iterator.Iterator[value.Value], w *xmlWriter.XMLWriter, style string, maxListSize int) {
	openWithStyle("table", style, w)
	i := 0
	it(func(v value.Value) bool {
		i++
		w.Open("tr")
		if i <= maxListSize {
			j := 0
			toList(v).Iterator()(func(c value.Value) bool {
				j++
				if j <= maxListSize {
					toTD(c, w, maxListSize)
				} else {
					w.Open("td").Write("more...").Close()
				}
				return j <= maxListSize
			})
		} else {
			w.Open("td").Write("more...").Close()
		}
		w.Close()
		return i <= maxListSize
	})
	w.Close()
}

func toTD(d value.Value, w *xmlWriter.XMLWriter, maxListSize int) {
	if formatted, ok := d.(format); ok {
		if _, isList := formatted.Value.(*value.List); isList && !formatted.Cell {
			w.Open("td")
			toHtml(formatted.Value, w, formatted.Format, maxListSize)
			w.Close()
		} else {
			w.Open("td").Attr("style", formatted.Format)
			toHtml(formatted.Value, w, "", maxListSize)
			w.Close()
		}
	} else {
		w.Open("td")
		toHtml(d, w, "", maxListSize)
		w.Close()
	}
}

func toList(r value.Value) *value.List {
	if l, ok := r.(*value.List); ok {
		return l
	}
	return value.NewListFromIterable(iterator.Single(r))
}
