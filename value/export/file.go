package export

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"math"
	"time"
)

type File struct {
	Name     string
	MimeType string
	Data     []byte
}

func (f File) ToList() (*value.List, bool) {
	return nil, false
}

func (f File) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (f File) ToFloat() (float64, bool) {
	return 0, false
}

func (f File) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprintf("file %s (%d bytes)", f.Name, len(f.Data)), nil
}

func (f File) GetType() value.Type {
	return value.FileTypeId
}

func AddFileHelpers(f *value.FunctionGenerator) {
	DataTypeId = f.RegisterType("DataFile")
	f.RegisterMethods(value.ListTypeId, value.MethodMap{
		"zip": value.MethodAtType(1, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				var buffer bytes.Buffer
				zip := zip.NewWriter(&buffer)
				for v, err := range list.Iterate(st) {
					if err != nil {
						return nil, err
					}
					if f, ok := v.(File); ok {
						w, err := zip.Create(f.Name)
						if err != nil {
							return nil, err
						}
						_, err = w.Write(f.Data)
					} else {
						return nil, errors.New("zipFiles requires a list of files")
					}
				}
				err := zip.Close()
				if err != nil {
					return nil, err
				}

				return File{
					Name:     string(name) + ".zip",
					MimeType: "application/zip",
					Data:     buffer.Bytes(),
				}, nil
			}
			return nil, errors.New("zip requires a filename as argument")
		}).SetMethodDescription("name", "Creates a zip file from the list of files."),
	})
	f.RegisterMethods(DataTypeId, value.MethodMap{
		"add": value.MethodAtType(3, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if unit, ok := st.Get(2).(value.String); ok {
					if fu, ok := st.Get(3).(value.Closure); ok {
						if fu.Args != 1 {
							return nil, errors.New("data column function must have exactly one argument")
						}
						d := DataContent{
							Name:   string(name),
							Unit:   string(unit),
							Values: fu,
						}
						return data.Add(d), nil
					}
				}
			}
			return nil, errors.New("add requires a name, a unit and a function")
		}).SetMethodDescription("name", "unit", "func", "Creates a column in the data file."),
		"addIf": value.MethodAtType(4, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cond, ok := st.Get(1).(value.Bool); ok {
				if !cond {
					return data, nil
				}
				if name, ok := st.Get(2).(value.String); ok {
					if unit, ok := st.Get(3).(value.String); ok {
						if fu, ok := st.Get(4).(value.Closure); ok {
							if fu.Args != 1 {
								return nil, errors.New("data column function must have exactly one argument")
							}
							d := DataContent{
								Name:   string(name),
								Unit:   string(unit),
								Values: fu,
							}
							return data.Add(d), nil
						}
					}
				}
			}
			return nil, errors.New("addIf requires a bool, a name, a unit and a function")
		}).SetMethodDescription("cond", "name", "unit", "func", "Creates a column in the data file if the condition is true."),
		"timeIsDate": value.MethodAtType(0, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			data.TimeIsDate = true
			return data, nil
		}).SetMethodDescription("The time function returns a date given in seconds since 01.01.1970."),
		"timeFormat": value.MethodAtType(1, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if format, ok := st.Get(1).(value.String); ok {
				data.TimeFormat = string(format)
				return data, nil
			}
			return nil, errors.New("timeFormat requires a format string")
		}).SetMethodDescription("format", "Sets the time format."),
		"dateFormat": value.MethodAtType(1, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if format, ok := st.Get(1).(value.String); ok {
				data.DateFormat = string(format)
				return data, nil
			}
			return nil, errors.New("dateFormat requires a format string")
		}).SetMethodDescription("format", "Sets the date format."),
		"dat": value.MethodAtType(2, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if list, ok := st.Get(2).(*value.List); ok {
					d, err := data.DatFile(st, list)
					if err != nil {
						return nil, err
					}
					return File{
						Name:     string(name) + ".dat",
						MimeType: "text/text",
						Data:     d,
					}, nil
				}
			}
			return nil, errors.New("dat requires a name and a list")
		}).SetMethodDescription("name", "list", "Creates a dat file."),
		"csv": value.MethodAtType(2, func(data *Data, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if list, ok := st.Get(2).(*value.List); ok {
					d, err := data.CsvFile(st, list)
					if err != nil {
						return nil, err
					}
					return File{
						Name:     string(name) + ".csv",
						MimeType: "text/csv",
						Data:     d,
					}, nil
				}
			}
			return nil, errors.New("csv requires a name and a list")
		}).SetMethodDescription("name", "list", "Creates a csv file."),
	})
	f.AddStaticFunction("dataFile", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if name, ok := st.Get(0).(value.String); ok {
				if unit, ok := st.Get(1).(value.String); ok {
					if fu, ok := st.Get(2).(value.Closure); ok {
						if fu.Args != 1 {
							return nil, errors.New("data column function must have exactly one argument")
						}

						return &Data{
							Time:     fu,
							TimeName: string(name),
							TimeUnit: string(unit),
						}, nil
					}
				}
			}
			return nil, errors.New("dataFile requires a name, a unit and a function")
		},
		Args:   3,
		IsPure: true,
	})
}

type DataContent struct {
	Name   string
	Unit   string
	Values value.Closure
}

var DataTypeId value.Type

type Data struct {
	Time        value.Closure
	TimeIsDate  bool
	TimeUnit    string
	TimeFormat  string
	DateFormat  string
	TimeName    string
	DataContent []DataContent
}

func (d *Data) ToList() (*value.List, bool) {
	return nil, false
}

func (d *Data) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (d *Data) ToFloat() (float64, bool) {
	return 0, false
}

func (d *Data) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	sb := &bytes.Buffer{}
	sb.WriteString("DataFile(")
	for i, content := range d.DataContent {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(content.Name)
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func (d *Data) GetType() value.Type {
	return DataTypeId
}

func (d *Data) Add(content DataContent) *Data {
	var n = *d
	n.DataContent = append(n.DataContent, content)
	return &n
}

func (d *Data) DatFile(st funcGen.Stack[value.Value], list *value.List) ([]byte, error) {
	return d.writeFile(dat{}, st, list)
}

func (d *Data) CsvFile(st funcGen.Stack[value.Value], list *value.List) ([]byte, error) {
	df := d.DateFormat
	if df == "" {
		df = csvDateFormat
	}
	tf := d.TimeFormat
	if tf == "" {
		tf = csvTimeFormat
	}
	return d.writeFile(&csv{isDate: d.TimeIsDate, dateFormat: df, timeFormat: tf}, st, list)
}

func (d *Data) writeFile(f format, st funcGen.Stack[value.Value], rows *value.List) ([]byte, error) {
	var b bytes.Buffer

	f.writeHeader(&b, d)

	type errorHolder struct {
		err             error
		someRowsWritten bool
	}

	columns := make([]errorHolder, len(d.DataContent))
	for row, err := range rows.Iterate(st) {
		if err != nil {
			return nil, err
		}
		tVal, err := d.Time.Eval(st, row)
		if err != nil {
			return nil, err
		}
		if t, ok := tVal.ToFloat(); ok {
			f.writeTime(&b, t)

			for i, content := range d.DataContent {
				vVal, err := content.Values.Eval(st, row)
				if err == nil {
					if v, ok := vVal.ToFloat(); ok {
						f.writeValue(&b, v)
						columns[i].someRowsWritten = true
					} else {
						f.skipValue(&b)
					}
				} else {
					f.skipValue(&b)
					if columns[i].err == nil {
						columns[i].err = err
					}
				}
			}
		} else {
			return nil, fmt.Errorf("time value is not a float")
		}
	}

	var buf bytes.Buffer
	for i, column := range columns {
		if !column.someRowsWritten {
			if buf.Len() > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(d.DataContent[i].Name)
			if column.err != nil {
				buf.WriteString(": " + column.err.Error())
			}
		}
	}
	if buf.Len() > 0 {
		return nil, errors.New("no entry created in data column(s): " + buf.String())
	}

	return b.Bytes(), nil
}

type format interface {
	writeHeader(*bytes.Buffer, *Data)
	writeTime(*bytes.Buffer, float64)
	writeValue(*bytes.Buffer, float64)
	skipValue(*bytes.Buffer)
}

type dat struct{}

func (d dat) writeHeader(b *bytes.Buffer, data *Data) {
	if data.TimeIsDate {
		b.WriteString("#time is unix date\n")
	}

	b.WriteString("#" + data.TimeName + "[" + data.TimeUnit + "]")
	for _, content := range data.DataContent {
		b.WriteString("\t" + content.Name + "[" + content.Unit + "]")
	}
}

func (d dat) writeTime(b *bytes.Buffer, t float64) {
	b.WriteString(fmt.Sprintf("\n%g", t))
}

func (d dat) writeValue(b *bytes.Buffer, v float64) {
	b.WriteString(fmt.Sprintf("\t%g", v))
}

func (d dat) skipValue(b *bytes.Buffer) {
	b.WriteString("\t-")
}

const (
	csvDateFormat = "2006-01-02"
	csvTimeFormat = "15:04:05"
)

type csv struct {
	isDate     bool
	dateFormat string
	timeFormat string
}

func (c *csv) writeHeader(b *bytes.Buffer, data *Data) {
	if data.TimeIsDate {
		b.WriteString("\"date\",\"time\"")
	} else {
		b.WriteString("\"" + data.TimeName + "[" + data.TimeUnit + "]\"")
	}
	for _, content := range data.DataContent {
		b.WriteString(",\"" + content.Name + "[" + content.Unit + "]\"")
	}
}

func (c *csv) writeTime(b *bytes.Buffer, t float64) {
	if c.isDate {
		sec := int64(math.Trunc(t))
		nsec := int64((t - float64(sec)) * 1e9)
		unix := time.Unix(sec, nsec)
		b.WriteString(fmt.Sprintf("\n\"%s\"", unix.Format(c.dateFormat)))
		b.WriteString(fmt.Sprintf(",\"%s\"", unix.Format(c.timeFormat)))
	} else {
		b.WriteString(fmt.Sprintf("\n\"%g\"", t))
	}
}

func (c *csv) writeValue(b *bytes.Buffer, f float64) {
	b.WriteString(fmt.Sprintf(",\"%g\"", f))
}

func (c *csv) skipValue(b *bytes.Buffer) {
	b.WriteString(",\"\"")
}
