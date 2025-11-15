package export

import (
	"errors"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataString(t *testing.T) {
	data := &Data{
		TimeName: "time",
		TimeUnit: "s",
		Time: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				return st.Get(0), nil
			},
			Args: 1,
		},
	}
	data = data.Add(DataContent{
		Name: "square",
		Unit: "1",
		Values: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if f, ok := st.Get(0).ToFloat(); ok {
					return value.Float(f * f), nil
				}
				return nil, errors.New("invalid argument")
			},
			Args: 1,
		},
	})
	data = data.Add(DataContent{
		Name: "square2",
		Unit: "1",
		Values: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if f, ok := st.Get(0).ToFloat(); ok {
					return value.Float(2 * f * f), nil
				}
				return nil, errors.New("invalid argument")
			},
			Args: 1,
		},
	})
	st := funcGen.NewEmptyStack[value.Value]()
	list := value.NewListFromIterable(func(f funcGen.Stack[value.Value]) iterator.Producer[value.Value] {
		return iterator.Generate(10, func(i int) (value.Value, error) {
			return value.Int(i), nil
		})
	})
	dataFile, err := data.DatFile(st, list)
	assert.NoError(t, err)

	assert.EqualValues(t, `#time[s]	square[1]	square2[1]
0	0	0
1	1	2
2	4	8
3	9	18
4	16	32
5	25	50
6	36	72
7	49	98
8	64	128
9	81	162`, string(dataFile))

	csvFile, err := data.CsvFile(st, list)
	assert.NoError(t, err)

	assert.EqualValues(t, `"time[s]","square[1]","square2[1]"
"0","0","0"
"1","1","2"
"2","4","8"
"3","9","18"
"4","16","32"
"5","25","50"
"6","36","72"
"7","49","98"
"8","64","128"
"9","81","162"`, string(csvFile))
}

func TestDataString2(t *testing.T) {
	data := &Data{
		TimeName: "time",
		TimeUnit: "s",
		Time: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				return st.Get(0), nil
			},
			Args: 1,
		},
	}
	data = data.Add(DataContent{
		Name: "square",
		Unit: "1",
		Values: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if f, ok := st.Get(0).ToFloat(); ok {
					if int(f)%2 == 0 {
						return value.String("lll"), nil
					}
					return value.Float(f * f), nil
				}
				return nil, errors.New("invalid argument")
			},
			Args: 1,
		},
	})
	data = data.Add(DataContent{
		Name: "square2",
		Unit: "1",
		Values: value.Closure{
			Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if f, ok := st.Get(0).ToFloat(); ok {
					return value.Float(2 * f * f), nil
				}
				return nil, errors.New("invalid argument")
			},
			Args: 1,
		},
	})
	st := funcGen.NewEmptyStack[value.Value]()
	list := value.NewListFromIterable(func(f funcGen.Stack[value.Value]) iterator.Producer[value.Value] {
		return iterator.Generate(10, func(i int) (value.Value, error) {
			return value.Int(i), nil
		})
	})
	dataFile, err := data.DatFile(st, list)
	assert.NoError(t, err)

	assert.EqualValues(t, `#time[s]	square[1]	square2[1]
0	-	0
1	1	2
2	-	8
3	9	18
4	-	32
5	25	50
6	-	72
7	49	98
8	-	128
9	81	162`, string(dataFile))

	csvFile, err := data.CsvFile(st, list)
	assert.NoError(t, err)

	assert.EqualValues(t, `"time[s]","square[1]","square2[1]"
"0","","0"
"1","1","2"
"2","","8"
"3","9","18"
"4","","32"
"5","25","50"
"6","","72"
"7","49","98"
"8","","128"
"9","81","162"`, string(csvFile))
}
