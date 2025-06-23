package export

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
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
	Format value.Value
}

// AddHTMLStylingHelpers adds the following functions to the function generator:
//
//	style(style, value):  Formats the value with the given style.
//
//	styleCell(style, value):  Formats the value with the given style. If used in a table, the style is applied to the cell instead of the containing value.
//
//	link(name, link):  Used to create a link.
//
//	styleBins(fac):  Simple styling to format binning results. The width of the bars is scaled by the factor fac.
//
//	styleBinsSkipFirst(fac):  Simple styling to format binning results. The width of the bars is scaled by the factor fac. The first entry is skipped.
func AddHTMLStylingHelpers(f *value.FunctionGenerator) {
	f.AddStaticFunction("style", styleFunc)
	f.AddStaticFunction("styleCell", styleFuncCell)
	f.AddStaticFunction("link", linkFunc)
	//If the parser is used, further modifications are not possible
	//f.AddStaticFunction("styleBins", funcGen.Function[value.Value]{
	//	Func:   value.Must(f.GenerateFromString("m->m.descr.number((n,e)->[e.str, [[style(\"background:red;width:\"+(m.values[n]*fac)+\"px\",\"\"),m.values[n]]] ])", "fac")),
	//	Args:   1,
	//	IsPure: true,
	//}.SetDescription("fac", "Simple styling to format binning results. The width of the bars is scaled by the factor fac."))
	//f.AddStaticFunction("styleBinsSkipFirst", funcGen.Function[value.Value]{
	//	Func:   value.Must(f.GenerateFromString("m->m.descr.skip(1).number((n,e)->[e.str, [[style(\"background:red;width:\"+(m.values[n+1]*fac)+\"px\",\"\"),m.values[n+1]]] ])", "fac")),
	//	Args:   1,
	//	IsPure: true,
	//}.SetDescription("fac", "Simple styling to format binning results. The width of the bars is scaled by the factor fac."))
}

// linkFunc can be used to create a link
var linkFunc = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
		if l, ok := st.Get(0).(value.String); ok {
			return Link{
				Link:  string(l),
				Value: st.Get(1),
			}, nil
		}
		return nil, errors.New("link requires two strings (name,link)")
	},
	Args:   2,
	IsPure: true,
}.SetDescription("name", "link", "Used to create a link.")

// styleFunc can be used add a CSS style to a value
var styleFunc = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
		return Format{
			Value:  st.Get(1),
			Cell:   false,
			Format: st.Get(0),
		}, nil
	},
	Args:   2,
	IsPure: true,
}.SetDescription("style", "value", "Formats the value with the given style.")

// styleFuncCell can be used add a CSS stale to a value
// If used in a table the Format is applied to the cell instead of the containing value.
// It is only required in rare occasions.
var styleFuncCell = funcGen.Function[value.Value]{
	Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
		return Format{
			Value:  st.Get(1),
			Cell:   true,
			Format: st.Get(0),
		}, nil
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

func (f Format) GetType() value.Type {
	return value.FormatTypeId
}

type Link struct {
	Link  string
	Value value.Value
}

func (l Link) ToList() (*value.List, bool) {
	return l.Value.ToList()
}

func (l Link) ToMap() (value.Map, bool) {
	return l.Value.ToMap()
}

func (l Link) ToInt() (int, bool) {
	return l.Value.ToInt()
}

func (l Link) ToFloat() (float64, bool) {
	return l.Value.ToFloat()
}

func (l Link) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return l.Value.ToString(st)
}

func (l Link) ToBool() (bool, bool) {
	return l.Value.ToBool()
}

func (l Link) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (l Link) GetType() value.Type {
	return value.LinkTypeId
}

type File struct {
	Name string
	Data []byte
}

func (f File) ToList() (*value.List, bool) {
	return nil, false
}

func (f File) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (f File) ToInt() (int, bool) {
	return 0, false
}

func (f File) ToFloat() (float64, bool) {
	return 0, false
}

func (f File) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprintf("file %s (%d bytes)", f.Name, len(f.Data)), nil
}

func (f File) ToBool() (bool, bool) {
	return false, false
}

func (f File) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (f File) GetType() value.Type {
	return value.FileTypeId
}

func AddZipHelpers(f *value.FunctionGenerator) {
	f.AddStaticFunction("zipFiles", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
			if name, ok := st.Get(0).(value.String); ok {
				if list, ok := st.Get(1).ToList(); ok {
					var buffer bytes.Buffer
					zip := zip.NewWriter(&buffer)
					err := list.Iterate(st, func(v value.Value) error {
						if f, ok := v.(File); ok {
							w, err := zip.Create(f.Name)
							if err != nil {
								return err
							}
							_, err = w.Write(f.Data)
							return err
						}
						return errors.New("zipFiles requires a list of files")
					})
					if err != nil {
						return nil, err
					}
					err = zip.Close()
					if err != nil {
						return nil, err
					}

					return File{
						Name: string(name) + ".zip",
						Data: buffer.Bytes(),
					}, nil
				}
			}
			return nil, errors.New("zipFiles requires a string and a list of files")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("name", "list of files", "Used to create a zip file."))
}

type CustomHTML func(value.Value) (template.HTML, bool, error)

// ToHtml creates an HTML representation of a value
// Lists and maps are converted to a html table.
// Everything else is converted to a string by calling the ToString() method.
func ToHtml(v value.Value, maxListSize int, custom CustomHTML, inlineStyle bool) (res template.HTML, list []Class, err error) {
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
	w := xmlWriter.New().AvoidShort().PrettyPrint()
	ex := htmlExporter{w: w, maxListSize: maxListSize, custom: custom, styleMap: make(map[string]string), inlineStyle: inlineStyle}
	err = ex.toHtml(funcGen.NewEmptyStack[value.Value](), v, nil)
	if err != nil {
		return "", list, err
	}
	return template.HTML(w.String()), ex.classList, nil
}

type Class struct {
	Name  string
	Style template.CSS
}

type htmlExporter struct {
	w           *xmlWriter.XMLWriter
	maxListSize int
	custom      CustomHTML
	classList   []Class
	styleMap    map[string]string
	inlineStyle bool
}

func (ex *htmlExporter) getClassName(style string) string {
	if className, ok := ex.styleMap[style]; ok {
		return className
	}
	className := "c" + strconv.Itoa(len(ex.styleMap))
	ex.styleMap[style] = className
	ex.classList = append(ex.classList, Class{Name: className, Style: template.CSS(style)})
	return className
}

type byteSize int

var byteUnits = []string{"Bytes", "kBytes", "MBytes", "GBytes", "TBytes"}

func (b byteSize) String() string {
	us := b
	unit := 0
	for us > 10000 && unit < len(byteUnits)-1 {
		unit++
		us = us / 1024
	}
	return strconv.Itoa(int(us)) + " " + byteUnits[unit]
}

type ToHtmlInterface interface {
	ToHtml(st funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error
}

func (ex *htmlExporter) toHtml(st funcGen.Stack[value.Value], v, style value.Value) error {
	if style != nil {
		if cl, ok := style.ToClosure(); ok {
			if cl.Args == 1 {
				if res, err := cl.Eval(st, v); err == nil {
					return ex.toHtml(st, res, nil)
				} else {
					return err
				}
			}
		}
	}

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
	case Link:
		ex.w.Open("a").Attr("href", t.Link)
		err := ex.toHtml(st, t.Value, style)
		ex.w.Close()
		return err
	case File:
		dataStr := base64.StdEncoding.EncodeToString(t.Data)
		dataStr = "data:application/octet-stream;base64," + dataStr
		ex.w.Open("a").Attr("href", dataStr)
		ex.w.Attr("download", t.Name)
		ex.w.Write("File: " + t.Name + " (" + byteSize(len(t.Data)).String() + ")")
		ex.w.Close()
	case *value.List:
		if hasKey(style, "plainList") {
			return t.Iterate(st, func(v value.Value) error {
				return ex.toHtml(st, v, nil)
			})
		} else {
			var le listExporter
			err := t.Iterate(st, func(v value.Value) error {
				if le == nil {
					le = ex.createListExporter(st, v)
					le.open(style)
				}
				ok, err := le.add(v)
				if !ok && err == nil {
					return iterator.SBC
				}
				return err
			})
			if le != nil {
				le.close()
			}
			if err == iterator.SBC {
				return nil
			}
			return err
		}
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
			if th, ok := v.(ToHtmlInterface); ok {
				err := th.ToHtml(st, ex.w)
				if err != nil {
					return err
				}
			} else {
				s, err := v.ToString(st)
				if err != nil {
					return err
				}
				ex.writeHtmlString(s, style)
			}
		}
	}
	return nil
}

func hasKey(style value.Value, key string) bool {
	if style != nil {
		if m, ok := style.ToMap(); ok {
			_, ok = m.Get(key)
			return ok
		}
		if s, ok := style.(value.String); ok {
			return string(s) == key
		}
	}
	return false
}

func toStyleStr(style value.Value) (string, bool) {
	switch t := style.(type) {
	case value.String:
		return string(t), true
	case value.Map:
		type kv struct {
			k, v string
		}
		var keys []kv
		t.Iter(func(k string, v value.Value) bool {
			ck := strings.ReplaceAll(k, "_", "-")
			switch s := v.(type) {
			case value.String:
				keys = append(keys, kv{ck, string(s)})
			case value.Int:
				keys = append(keys, kv{ck, strconv.Itoa(int(s))})
			case value.Float:
				keys = append(keys, kv{ck, strconv.FormatFloat(float64(s), 'f', -1, 64)})
			}
			return true
		})
		if len(keys) == 0 {
			return "", false
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].k < keys[j].k
		})
		var res strings.Builder
		for _, k := range keys {
			res.WriteString(k.k)
			res.WriteString(":")
			res.WriteString(k.v)
			res.WriteString(";")
		}
		return res.String(), true
	default:
		return "", false
	}
}

func (ex *htmlExporter) writeHtmlString(s string, style value.Value) {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		ex.w.Open("a").Attr("href", s).Attr("target", "_blank")
		ex.w.Write("Link")
		ex.w.Close()
	} else if strings.HasPrefix(s, "host:") {
		ex.w.Open("a").Attr("href", s[5:]).Attr("target", "_blank")
		ex.w.Write("Link")
		ex.w.Close()
	} else {
		if strStyle, ok := toStyleStr(style); ok {
			if ex.inlineStyle {
				ex.w.Open("span").Attr("style", strStyle)
			} else {
				ex.w.Open("span").Attr("class", ex.getClassName(strStyle))
			}
			ex.w.Write(s)
			ex.w.Close()
		} else {
			ex.w.Write(s)
		}
	}
}

type listExporter interface {
	open(style value.Value)
	add(value value.Value) (bool, error)
	close()
}

type dummy struct{}

func (d dummy) open(style value.Value) {
}
func (d dummy) add(v value.Value) (bool, error) {
	return false, nil
}
func (d dummy) close() {
}

func (ex *htmlExporter) createListExporter(st funcGen.Stack[value.Value], v value.Value) listExporter {
	if v == nil {
		return dummy{}
	}
	if _, ok := v.(*value.List); ok {
		return &tableExporter{ex: ex, st: st}
	}
	return &simpleListExporter{ex: ex, st: st}
}

type simpleListExporter struct {
	ex *htmlExporter
	st funcGen.Stack[value.Value]
	i  int
}

func (r *simpleListExporter) open(style value.Value) {
	r.ex.openWithStyle("table", style)
}

func (r *simpleListExporter) add(value value.Value) (bool, error) {
	r.i++
	r.ex.w.Open("tr")
	r.ex.w.Open("td").Write(strconv.Itoa(r.i)).Write(".").Close()
	if r.i <= r.ex.maxListSize {
		err := r.ex.toTD(r.st, value)
		if err != nil {
			return false, err
		}
	} else {
		r.ex.w.Open("td").Write("more...").Close()
	}
	r.ex.w.Close()
	return r.i <= r.ex.maxListSize, nil
}

func (r *simpleListExporter) close() {
	r.ex.w.Close()
}

func (ex *htmlExporter) openWithStyle(tag string, style value.Value) {
	if strStyle, ok := toStyleStr(style); ok {
		if ex.inlineStyle {
			ex.w.Open(tag).Attr("style", strStyle)
		} else {
			ex.w.Open(tag).Attr("class", ex.getClassName(strStyle))
		}
	} else {
		ex.w.Open(tag)
	}
}

type tableExporter struct {
	ex          *htmlExporter
	st          funcGen.Stack[value.Value]
	row         int
	tableFormat value.MapStorage
}

func (t *tableExporter) open(style value.Value) {
	t.ex.openWithStyle("table", style)
	if style != nil {
		if m, ok := style.ToMap(); ok {
			if ta, ok := m.Get("table"); ok {
				if m, ok := ta.ToMap(); ok {
					t.tableFormat = m.Storage()
				}
			}
		}
	}
}

func (t *tableExporter) add(val value.Value) (bool, error) {
	t.row++
	t.ex.w.Open("tr")
	if t.row <= t.ex.maxListSize {
		col := 0
		err := toList(val).Iterate(t.st, func(item value.Value) error {
			col++
			if col <= t.ex.maxListSize {
				err := t.ex.toTD(t.st, t.format(t.row, col, item))
				if err != nil {
					return err
				}
			} else {
				t.ex.w.Open("td").Write("more...").Close()
			}
			if col <= t.ex.maxListSize {
				return nil
			} else {
				return iterator.SBC
			}
		})
		if err != nil && err != iterator.SBC {
			return false, err
		}
	} else {
		t.ex.w.Open("td").Write("more...").Close()
	}
	t.ex.w.Close()
	return t.row <= t.ex.maxListSize, nil
}

func (t *tableExporter) close() {
	t.ex.w.Close()
}

func (t *tableExporter) format(row, col int, item value.Value) value.Value {
	if t.tableFormat == nil {
		return item
	}
	var format value.Value
	if v, ok := t.tableFormat.Get("r" + strconv.Itoa(row) + "c" + strconv.Itoa(col)); ok {
		format = v
	} else if v, ok := t.tableFormat.Get("r" + strconv.Itoa(row)); ok {
		format = v
	} else if v, ok := t.tableFormat.Get("c" + strconv.Itoa(col)); ok {
		format = v
	} else if v, ok := t.tableFormat.Get("all"); ok {
		format = v
	}
	if format == nil {
		return item
	}

	if cl, ok := format.ToClosure(); ok {
		if cl.Args == 1 {
			if res, err := cl.Eval(t.st, item); err == nil {
				return res
			}
		} else if cl.Args == 3 {
			if res, err := cl.EvalSt(t.st, value.Int(row), value.Int(col), item); err == nil {
				return res
			}
		}
	}
	return Format{Value: item, Format: format, Cell: true}
}

func (ex *htmlExporter) toTD(st funcGen.Stack[value.Value], d value.Value) error {
	var err error
	if formatted, ok := d.(Format); ok {
		if _, isList := formatted.Value.(*value.List); isList && !formatted.Cell {
			ex.w.Open("td")
			err = ex.toHtml(st, formatted.Value, formatted.Format)
			ex.w.Close()
		} else {
			ex.w.Open("td")
			if strStyle, ok := toStyleStr(formatted.Format); ok {
				if ex.inlineStyle {
					ex.w.Attr("style", strStyle)
				} else {
					ex.w.Attr("class", ex.getClassName(strStyle))
				}
			}
			err = ex.toHtml(st, formatted.Value, nil)
			ex.w.Close()
		}
	} else {
		ex.w.Open("td")
		err = ex.toHtml(st, d, nil)
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
