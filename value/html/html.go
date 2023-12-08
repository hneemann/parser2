// Package html is used to create an HTML representation of a value
// It's ToHtml function returns a sanitized html string.
package html

import (
	"errors"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/html/xmlWriter"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

type Format struct {
	Value  value.Value
	Cell   bool
	Format string
}

// StyleFunc can be used add a CSS style to a value
var StyleFunc = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
		if s, ok := st.Get(0).(value.String); ok {
			return Format{
				Value:  st.Get(1),
				Cell:   false,
				Format: string(s),
			}, nil
		} else {
			return nil, errors.New("style must be a string")
		}
	},
	Args:   2,
	IsPure: true,
}

// StyleFuncCell can be used add a CSS stale to a value
// If used in a table the Format is applied to the cell instead of the containing value.
// It is only required in rare occasions.
var StyleFuncCell = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
		if s, ok := st.Get(0).(value.String); ok {
			return Format{
				Value:  st.Get(1),
				Cell:   true,
				Format: string(s),
			}, nil
		} else {
			return nil, errors.New("style must be a string")
		}
	},
	Args:   2,
	IsPure: true,
}

func (f Format) ToList() (*value.List, bool) {
	return f.Value.ToList()
}

func (f Format) ToMap() (value.Map, bool) {
	return f.Value.ToMap()
}

func (f Format) ToInt() (int, bool) {
	return f.Value.ToInt()
}

func (f Format) ToFloat() (float64, bool) {
	return f.Value.ToFloat()
}

func (f Format) String() (string, error) {
	return f.Value.String()
}

func (f Format) ToBool() (bool, bool) {
	return f.Value.ToBool()
}

func (f Format) ToClosure() (funcGen.Function[value.Value], bool) {
	return f.Value.ToClosure()
}

func (f Format) GetMethod(name string) (funcGen.Function[value.Value], error) {
	m, err := f.Value.GetMethod(name)
	if err != nil {
		return funcGen.Function[value.Value]{}, err
	}
	return funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			ss := st.Size()
			st.Push((st.Get(0).(Format)).Value)
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
		if rec := recover(); rec != nil {
			err = parser2.AnyToError(rec)
			res = ""
		}
	}()
	if maxListSize < 1 {
		maxListSize = 1
	}
	w := xmlWriter.New()
	err = toHtml(v, w, "", maxListSize)
	if err != nil {
		return "", err
	}
	return template.HTML(w.String()), nil
}

func toHtml(v value.Value, w *xmlWriter.XMLWriter, style string, maxListSize int) error {
	switch t := v.(type) {
	case Format:
		return toHtml(t.Value, w, t.Format, maxListSize)
	case *value.List:
		pit, f, err := iterator.Peek(t.Iterator())
		if err != nil {
			return err
		}
		if _, ok := f.(*value.List); ok {
			return tableToHtml(pit, w, style, maxListSize)
		}
		return listToHtml(pit, w, style, maxListSize)
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
			err := toTD(v, w, maxListSize)
			if err != nil {
				return err
			}
			w.Close()
		}
		w.Close()
	default:
		if v == nil {
			w.Write("nil")
		} else {
			s, err := v.String()
			if err != nil {
				return err
			}
			writeHtmlString(s, style, w)
		}
	}
	return nil
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

func listToHtml(it iterator.Iterator[value.Value], w *xmlWriter.XMLWriter, style string, maxListSize int) error {
	openWithStyle("table", style, w)
	i := 0
	var innerErr error
	_, err := it(func(e value.Value) bool {
		i++
		w.Open("tr")
		w.Open("td").Write(strconv.Itoa(i)).Write(".").Close()
		if i <= maxListSize {
			err := toTD(e, w, maxListSize)
			if err != nil {
				innerErr = err
				return false
			}
		} else {
			w.Open("td").Write("more...").Close()
		}
		w.Close()
		return i <= maxListSize
	})
	if innerErr != nil {
		return innerErr
	}
	w.Close()
	return err
}

func openWithStyle(tag string, style string, w *xmlWriter.XMLWriter) {
	if style == "" {
		w.Open(tag)
	} else {
		w.Open(tag).Attr("style", style)
	}
}

func tableToHtml(it iterator.Iterator[value.Value], w *xmlWriter.XMLWriter, style string, maxListSize int) error {
	openWithStyle("table", style, w)
	i := 0
	var outerErr error
	_, err := it(func(v value.Value) bool {
		i++
		w.Open("tr")
		if i <= maxListSize {
			j := 0
			var innerErr error
			_, err := toList(v).Iterator()(func(c value.Value) bool {
				j++
				if j <= maxListSize {
					err := toTD(c, w, maxListSize)
					if err != nil {
						innerErr = err
						return false
					}
				} else {
					w.Open("td").Write("more...").Close()
				}
				return j <= maxListSize
			})
			if innerErr != nil {
				outerErr = innerErr
				return false
			}
			if err != nil {
				outerErr = err
				return false
			}
		} else {
			w.Open("td").Write("more...").Close()
		}
		w.Close()
		return i <= maxListSize
	})
	if outerErr != nil {
		return outerErr
	}
	w.Close()
	return err
}

func toTD(d value.Value, w *xmlWriter.XMLWriter, maxListSize int) error {
	var err error
	if formatted, ok := d.(Format); ok {
		if _, isList := formatted.Value.(*value.List); isList && !formatted.Cell {
			w.Open("td")
			err = toHtml(formatted.Value, w, formatted.Format, maxListSize)
			w.Close()
		} else {
			w.Open("td").Attr("style", formatted.Format)
			err = toHtml(formatted.Value, w, "", maxListSize)
			w.Close()
		}
	} else {
		w.Open("td")
		err = toHtml(d, w, "", maxListSize)
		w.Close()
	}
	return err
}

func toList(r value.Value) *value.List {
	if l, ok := r.(*value.List); ok {
		return l
	}
	return value.NewListFromIterable(iterator.Single(r))
}
