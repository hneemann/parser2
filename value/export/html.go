package export

import (
	"errors"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"html/template"
	"log"
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
}.SetDescription("style", "value", "Formats the value with the given style.")

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
}.SetDescription("style", "value", "Formats the value with the given style. If used in a table, "+
	"the style is applied to the cell instead of the containing value.")

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

func (f Format) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return f.Value.ToString(st)
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

type CustomHTML func(value.Value) (template.HTML, bool, error)

// ToHtml creates an HTML representation of a value
// Lists and maps are converted to a html table.
// Everything else is converted to a string by calling the ToString() method.
func ToHtml(v value.Value, maxListSize int, custom CustomHTML) (res template.HTML, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Print("panic in ToHtml: ", rec)
			err = parser2.AnyToError(rec)
			res = ""
		}
	}()
	if maxListSize < 1 {
		maxListSize = 1
	}
	w := xmlWriter.New().AvoidShort()
	ex := htmlExporter{w: w, maxListSize: maxListSize, custom: custom}
	err = ex.toHtml(funcGen.NewEmptyStack[value.Value](), v, "")
	if err != nil {
		return "", err
	}
	return template.HTML(w.String()), nil
}

type htmlExporter struct {
	w           *xmlWriter.XMLWriter
	maxListSize int
	custom      CustomHTML
}

func (ex *htmlExporter) toHtml(st funcGen.Stack[value.Value], v value.Value, style string) error {
	if ex.custom != nil {
		if htm, ok, err := ex.custom(v); ok || err != nil {
			if err != nil {
				return err
			}
			ex.w.WriteHTML(htm)
			return nil
		}
	}
	switch t := v.(type) {
	case Format:
		return ex.toHtml(st, t.Value, t.Format)
	case *value.List:
		pit, f, err := iterator.Peek(t.Iterator(st))
		if err != nil {
			return err
		}
		if f == nil {
			return nil
		} else {
			if _, ok := f.(*value.List); ok {
				return ex.tableToHtml(st, pit, style)
			}
		}
		return ex.listToHtml(st, pit, style)
	case value.Map:
		ex.openWithStyle("table", style)
		var keys []string
		t.Iter(func(k string, v value.Value) bool {
			keys = append(keys, k)
			return true
		})
		sort.Strings(keys)
		for _, k := range keys {
			ex.w.Open("tr")
			ex.w.Open("td")
			ex.w.Write(k)
			ex.w.Write(":")
			ex.w.Close()
			v, _ := t.Get(k)
			err := ex.toTD(st, v)
			if err != nil {
				return err
			}
			ex.w.Close()
		}
		ex.w.Close()
	default:
		if v == nil {
			ex.w.Write("nil")
		} else {
			s, err := v.ToString(st)
			if err != nil {
				return err
			}
			ex.writeHtmlString(s, style)
		}
	}
	return nil
}

func (ex *htmlExporter) writeHtmlString(s string, style string) {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		ex.w.Open("a").Attr("href", s).Attr("target", "_blank")
		ex.w.Write("Link")
		ex.w.Close()
	} else if strings.HasPrefix(s, "host:") {
		ex.w.Open("a").Attr("href", s[5:]).Attr("target", "_blank")
		ex.w.Write("Link")
		ex.w.Close()
	} else {
		if style == "" {
			ex.w.Write(s)
		} else {
			ex.w.Open("span").Attr("style", style)
			ex.w.Write(s)
			ex.w.Close()
		}
	}
}

func (ex *htmlExporter) listToHtml(st funcGen.Stack[value.Value], it iterator.Iterator[value.Value], style string) error {
	ex.openWithStyle("table", style)
	i := 0
	var innerErr error
	_, err := it(func(e value.Value) bool {
		i++
		ex.w.Open("tr")
		ex.w.Open("td").Write(strconv.Itoa(i)).Write(".").Close()
		if i <= ex.maxListSize {
			err := ex.toTD(st, e)
			if err != nil {
				innerErr = err
				return false
			}
		} else {
			ex.w.Open("td").Write("more...").Close()
		}
		ex.w.Close()
		return i <= ex.maxListSize
	})
	if innerErr != nil {
		return innerErr
	}
	ex.w.Close()
	return err
}

func (ex *htmlExporter) openWithStyle(tag string, style string) {
	if style == "" {
		ex.w.Open(tag)
	} else {
		ex.w.Open(tag).Attr("style", style)
	}
}

func (ex *htmlExporter) tableToHtml(st funcGen.Stack[value.Value], it iterator.Iterator[value.Value], style string) error {
	ex.openWithStyle("table", style)
	i := 0
	var outerErr error
	_, err := it(func(v value.Value) bool {
		i++
		ex.w.Open("tr")
		if i <= ex.maxListSize {
			j := 0
			var innerErr error
			_, err := toList(v).Iterator(st)(func(c value.Value) bool {
				j++
				if j <= ex.maxListSize {
					err := ex.toTD(st, c)
					if err != nil {
						innerErr = err
						return false
					}
				} else {
					ex.w.Open("td").Write("more...").Close()
				}
				return j <= ex.maxListSize
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
			ex.w.Open("td").Write("more...").Close()
		}
		ex.w.Close()
		return i <= ex.maxListSize
	})
	if outerErr != nil {
		return outerErr
	}
	ex.w.Close()
	return err
}

func (ex *htmlExporter) toTD(st funcGen.Stack[value.Value], d value.Value) error {
	var err error
	if formatted, ok := d.(Format); ok {
		if _, isList := formatted.Value.(*value.List); isList && !formatted.Cell {
			ex.w.Open("td")
			err = ex.toHtml(st, formatted.Value, formatted.Format)
			ex.w.Close()
		} else {
			ex.w.Open("td").Attr("style", formatted.Format)
			err = ex.toHtml(st, formatted.Value, "")
			ex.w.Close()
		}
	} else {
		ex.w.Open("td")
		err = ex.toHtml(st, d, "")
		ex.w.Close()
	}
	return err
}

func toList(r value.Value) *value.List {
	if l, ok := r.(*value.List); ok {
		return l
	}
	return value.NewListFromIterable(iterator.Single[value.Value, funcGen.Stack[value.Value]](r))
}
