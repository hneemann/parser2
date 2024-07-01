package value

import (
	"errors"
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"strconv"
)

func newBinning(start, size float64, count int) *BinningData {
	bins := make([]float64, count+2)
	return &BinningData{axis{start, size, len(bins)}, bins}
}

type axis struct {
	start float64
	size  float64
	bins  int
}

func (a *axis) getIndex(v float64) int {
	index := int(math.Floor((v-a.start)/a.size)) + 1
	if index < 0 {
		index = 0
	} else if index >= a.bins {
		index = a.bins - 1
	}
	return index
}

type bin struct {
	IsMin bool
	Min   float64
	IsMax bool
	Max   float64
}

func (a *axis) getDescr(i int) bin {
	to := a.start + float64(i)*a.size
	from := to - a.size
	switch i {
	case 0:
		return bin{IsMax: true, Max: to}
	case a.bins - 1:
		return bin{IsMin: true, Min: from}
	default:
		return bin{IsMin: true, Min: from, IsMax: true, Max: to}
	}
}

type BinningData struct {
	a    axis
	bins []float64
}

func (s *BinningData) Add(value, toSum float64) {
	s.bins[s.a.getIndex(value)] += toSum
}

func (s *BinningData) Result(yield func(bin bin, value float64) bool) bool {
	for i, v := range s.bins {
		if !yield(s.a.getDescr(i), v) {
			return false
		}
	}
	return true
}

func New2d(xStart, xSize float64, xCount int, yStart, ySize float64, yCount int) *Binning2dData {
	bins := make([][]float64, xCount+2)
	for i := range bins {
		bins[i] = make([]float64, yCount+2)
	}
	return &Binning2dData{axis{xStart, xSize, len(bins)}, axis{yStart, ySize, len(bins[0])}, bins}
}

type Binning2dData struct {
	x    axis
	y    axis
	bins [][]float64
}

func (s *Binning2dData) Add(x, y, toSum float64) {
	xi := s.x.getIndex(x)
	yi := s.y.getIndex(y)
	s.bins[xi][yi] += toSum
}

func (s *Binning2dData) Result(yield func(bin, func(func(float64) bool) bool) bool) bool {
	for i, y := range s.bins {
		d := s.x.getDescr(i)
		if !yield(d, func(yie func(float64) bool) bool {
			for _, v := range y {
				if !yie(v) {
					return false
				}
			}
			return true
		}) {
			return false
		}
	}
	return true
}

func (s *Binning2dData) DescrY(yield func(bin) bool) bool {
	for i := range s.bins[0] {
		d := s.y.getDescr(i)
		if !yield(d) {
			return false
		}
	}
	return true
}

func Binning(l *List, st funcGen.Stack[Value]) (Value, error) {
	start, err := ToFloat("binning", st, 1)
	if err != nil {
		return nil, err
	}
	size, err := ToFloat("binning", st, 2)
	if err != nil {
		return nil, err
	}
	count, err := ToFloat("binning", st, 3)
	if err != nil {
		return nil, err
	}
	indFunc, err := ToFunc("binning", st, 4, 1)
	if err != nil {
		return nil, err
	}
	valFunc, err := ToFunc("binning", st, 5, 1)
	if err != nil {
		return nil, err
	}

	b := newBinning(start, size, int(count))
	err = l.Iterate(st, func(v Value) error {
		ind, err2 := MustFloat(indFunc.Eval(st, v))
		if err2 != nil {
			return err2
		}
		val, err2 := MustFloat(valFunc.Eval(st, v))
		if err2 != nil {
			return err2
		}
		b.Add(ind, val)
		return nil
	})
	if err != nil {
		return nil, err
	}
	var vals []Value
	var desc []Value
	b.Result(func(d bin, v float64) bool {
		vals = append(vals, Float(v))
		desc = append(desc, NewMap(d))
		return true
	})
	return NewMap(
		listMap.New[Value](2).
			Append("descr", NewList(desc...)).
			Append("values", NewList(vals...))), nil
}

// func (v *List) Binning2d(xStart, xSize, xCount, yStart, ySize, yCount Value, xIndex, yIndex, toSum Value) Value {
func Binning2d(l *List, st funcGen.Stack[Value]) (Value, error) {
	xStart, err := ToFloat("binning2d", st, 1)
	if err != nil {
		return nil, err
	}
	xSize, err := ToFloat("binning2d", st, 2)
	if err != nil {
		return nil, err
	}
	xCount, err := ToFloat("binning2d", st, 3)
	if err != nil {
		return nil, err
	}
	yStart, err := ToFloat("binning2d", st, 4)
	if err != nil {
		return nil, err
	}
	ySize, err := ToFloat("binning2d", st, 5)
	if err != nil {
		return nil, err
	}
	yCount, err := ToFloat("binning2d", st, 6)
	if err != nil {
		return nil, err
	}
	xIndFunc, err := ToFunc("binning2d", st, 7, 1)
	if err != nil {
		return nil, err
	}
	yIndFunc, err := ToFunc("binning2d", st, 8, 1)
	if err != nil {
		return nil, err
	}
	toSumFunc, err := ToFunc("binning2d", st, 9, 1)
	if err != nil {
		return nil, err
	}

	b := New2d(xStart, xSize, int(xCount), yStart, ySize, int(yCount))
	err = l.Iterate(st, func(v Value) error {
		xInd, err2 := MustFloat(xIndFunc.Eval(st, v))
		if err2 != nil {
			return err2
		}
		yInd, err2 := MustFloat(yIndFunc.Eval(st, v))
		if err2 != nil {
			return err2
		}
		toSum, err2 := MustFloat(toSumFunc.Eval(st, v))
		if err2 != nil {
			return err2
		}
		b.Add(xInd, yInd, toSum)
		return nil
	})
	if err != nil {
		return nil, err
	}
	var vals []Value
	b.Result(func(xd bin, y func(func(float64) bool) bool) bool {
		var bin []Value
		y(func(v float64) bool {
			bin = append(bin, Float(v))
			return true
		})
		vals = append(vals, NewMap(listMap.New[Value](2).Append("xd", NewMap(xd)).Append("row", NewList(bin...))))
		return true
	})
	var yDesc []Value
	b.DescrY(func(b bin) bool {
		yDesc = append(yDesc, NewMap(b))
		return true
	})
	return NewMap(listMap.New[Value](2).Append("yDescr", NewList(yDesc...)).Append("values", NewList(vals...))), nil
}

const (
	binMin = "min"
	binMax = "max"
	binStr = "str"
)

func (b bin) Get(key string) (Value, bool) {
	if key == binStr {
		return String(b.String()), true
	} else if key == binMin {
		return Float(b.Min), b.IsMin
	} else if key == binMax {
		return Float(b.Max), b.IsMax
	} else {
		return nil, false
	}
}

func (b bin) Iter(yield func(string, Value) bool) bool {
	if !yield(binStr, String(b.String())) {
		return false
	}
	if b.IsMin {
		if !yield(binMin, Float(b.Min)) {
			return false
		}
	}
	if b.IsMax {
		if !yield(binMax, Float(b.Max)) {
			return false
		}
	}
	return true
}

func (b bin) Size() int {
	return 3
}

func (b bin) String() string {
	var format string
	if b.Min == math.Round(b.Min) && b.Max == math.Round(b.Max) {
		format = "%1.0f"
	} else {
		n := nks(b.Max - b.Min)
		format = "%1." + strconv.Itoa(n) + "f"
	}
	if !b.IsMin {
		return fmt.Sprintf("<"+format, b.Max)
	} else if !b.IsMax {
		return fmt.Sprintf(">"+format, b.Min)
	} else {
		return fmt.Sprintf(format+"-"+format, b.Min, b.Max)
	}
}

func nks(delta float64) int {
	if delta == 0 {
		return 0
	}
	delta = math.Abs(delta)
	n := int(math.Trunc(-math.Log10(delta)-0.000001)) + 1
	if n >= 0 {
		return n
	}
	return 0
}

type binningCollector interface {
	add(st funcGen.Stack[Value], m Map) error
	result() Value
}

func detectBinning(m Map) binningCollector {
	if d, ok := m.Get("descr"); ok {
		return &collectBinning1d{descr: d}
	}
	if d, ok := m.Get("yDescr"); ok {
		return &collectBinning2d{yDescr: d}
	}
	return nil
}

type collectBinning1d struct {
	descr Value
	vals  []float64
}

func (c *collectBinning1d) add(st funcGen.Stack[Value], m Map) error {
	if v, ok := m.Get("values"); ok {
		if list, ok := v.ToList(); ok {
			entries, err := list.ToSlice(st)
			if err != nil {
				return err
			}
			if c.vals == nil {
				c.vals = make([]float64, len(entries))
			} else {
				if len(c.vals) != len(entries) {
					return errors.New("CollectBinning: not all value lists have the same size")
				}
			}
			for i, e := range entries {
				fl, err := MustFloat(e, nil)
				if err != nil {
					return err
				}
				c.vals[i] += fl
			}
		} else {
			errors.New("CollectBinning: item map does not contain values as list")
		}
	} else {
		errors.New("CollectBinning: item map does not contain values")
	}
	return nil
}

func (c *collectBinning1d) result() Value {
	res := make([]Value, len(c.vals))
	for i, e := range c.vals {
		res[i] = Float(e)
	}
	return NewMap(listMap.New[Value](2).Append("descr", c.descr).Append("values", NewList(res...)))
}

type collectBinning2d struct {
	yDescr Value
	xd     []Value
	vals   [][]float64
}

func (c *collectBinning2d) add(st funcGen.Stack[Value], m Map) error {
	if v, ok := m.Get("values"); ok {
		if list, ok := v.ToList(); ok {
			entries, err := list.ToSlice(st) // list of maps {xd:String, row:[]float}
			if err != nil {
				return err
			}
			if c.vals == nil {
				c.vals = make([][]float64, len(entries))
				c.xd = make([]Value, len(entries))
			} else {
				if len(c.vals) != len(entries) {
					errors.New("CollectBinning: not all value lists have the same size")
				}
			}
			for i, e := range entries {
				if em, ok := e.ToMap(); ok {
					if row, ok := em.Get("row"); ok {
						if rowL, ok := row.ToList(); ok {
							ro, err := rowL.ToSlice(st)
							if err != nil {
								return err
							}
							if c.vals[i] == nil {
								c.vals[i] = make([]float64, len(ro))
								if xd, ok := em.Get("xd"); ok {
									c.xd[i] = xd
								}
							} else {
								if len(c.vals[i]) != len(ro) {
									errors.New("row len does not match")
								}
							}
							for j, e := range ro {
								float, err := MustFloat(e, nil)
								if err != nil {
									return err
								}
								c.vals[i][j] += float
							}
						} else {
							errors.New("row is not a list")
						}
					} else {
						errors.New("no row found")
					}
				} else {
					errors.New("entry is not a map")
				}
			}
		} else {
			errors.New("CollectBinning: item map does not contain values as list")
		}
	} else {
		errors.New("CollectBinning: item map does not contain values")
	}
	return nil
}

func (c *collectBinning2d) result() Value {
	values := []Value{}
	for i, xd := range c.xd {
		row := make([]Value, len(c.vals[i]))
		for j, v := range c.vals[i] {
			row[j] = Float(v)
		}
		values = append(values, NewMap(listMap.New[Value](2).Append("xd", xd).Append("row", NewList(row...))))
	}

	return NewMap(listMap.New[Value](2).Append("yDescr", c.yDescr).Append("values", NewList(values...)))
}

func CollectBinning(l *List, st funcGen.Stack[Value]) (Value, error) {
	var bc binningCollector
	err := l.Iterate(st, func(v Value) error {
		if m, ok := v.ToMap(); ok {
			if bc == nil {
				bc = detectBinning(m)
				if bc == nil {
					return errors.New("invalid binning type")
				}
			}
			bc.add(st, m)
		} else {
			return errors.New("CollectBinning: item is not a map")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if bc == nil {
		return nil, errors.New("collectBinning: no items")
	}
	return bc.result(), nil
}
