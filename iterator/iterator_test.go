package iterator

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"runtime"
	"strings"
	"testing"
	"time"
)

func addOne(i int) (int, error) {
	return i + 1, nil
}

func add(a, b int) (int, error) {
	return a + b, nil
}

func addSlow(a, b int) (int, error) {
	time.Sleep(time.Millisecond * 10)
	return a + b, nil
}

func square(i, v int) (int, error) {
	return v * v, nil
}

func squareSlow(i, v int) (int, error) {
	time.Sleep(time.Microsecond * itemProcessingTimeMicroSec * 2)
	return v * v, nil
}

func isEven(i int) (bool, error) {
	return i&1 == 0, nil
}

func equal[V comparable](a, b V) (bool, error) {
	return a == b, nil
}

func check[V comparable](t *testing.T, it Producer[V], items ...V) {
	for n := 0; n < 2; n++ {
		checkIterator[V](t, it, items...)
	}
}

func checkIterator[V comparable](t *testing.T, it Producer[V], items ...V) {
	slice, err := ToSlice(it)
	assert.NoError(t, err)
	assert.EqualValues(t, items, slice)
}

func checkErr[V comparable](t *testing.T, it Producer[V], errMsg string, items ...V) {
	for n := 0; n < 2; n++ {
		checkErrIterator[V](t, it, errMsg, items...)
	}
}

func checkErrIterator[V comparable](t *testing.T, it Producer[V], errMsg string, items ...V) {
	slice, err := ToSlice(it)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), errMsg), "err does not contain "+errMsg)
	assert.EqualValues(t, items, slice)
}

func ints(n int) Producer[int] {
	return Generate[int](n, func(i int) (int, error) {
		return i, nil
	})
}

func TestIterables(t *testing.T) {
	type testCase[V any] struct {
		name string
		it   Producer[V]
		want []V
	}
	var empty []int
	tests := []testCase[int]{
		{name: "empty", it: Empty[int](), want: empty},
		{name: "single", it: Single[int](2), want: []int{2}},
		{name: "slice", it: Slice[int]([]int{1, 2, 3}), want: []int{1, 2, 3}},
		{name: "append", it: Append(Single[int](1), Single[int](2)), want: []int{1, 2}},
		{name: "appendSlice", it: Append(Slice[int]([]int{1, 2}), Slice[int]([]int{3, 4})), want: []int{1, 2, 3, 4}},
		{name: "generate", it: Generate[int](4, addOne), want: []int{1, 2, 3, 4}},
		{name: "map", it: Map(Slice[int]([]int{1, 2, 3}), square), want: []int{1, 4, 9}},
		{name: "filter", it: Filter(Slice[int]([]int{1, 2, 3, 4}), isEven), want: []int{2, 4}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := ToSlice(test.it)
			assert.NoError(t, err)
			assert.EqualValues(t, test.want, got)
			got, err = ToSlice(test.it)
			assert.NoError(t, err)
			assert.EqualValues(t, test.want, got)
		})
	}
}

func intEqual(a int, b int) (bool, error) {
	return a == b, nil
}

func TestEquals(t *testing.T) {
	type testCase[V any] struct {
		name  string
		it1   Producer[V]
		it2   Producer[V]
		equal bool
	}
	tests := []testCase[int]{
		{name: "empty", it1: Empty[int](), it2: Empty[int](), equal: true},
		{name: "slice", it1: Slice[int]([]int{1, 2}), it2: Slice[int]([]int{1, 2}), equal: true},
		{name: "slice", it1: Slice[int]([]int{1, 2}), it2: Slice[int]([]int{1, 3}), equal: false},
		{name: "slice", it1: Slice[int]([]int{1, 2}), it2: Slice[int]([]int{1}), equal: false},
		{name: "slice", it1: Slice[int]([]int{1}), it2: Slice[int]([]int{1, 2}), equal: false},
		{name: "append", it1: Append(Single[int](1), Single[int](2)), it2: Slice[int]([]int{1, 2}), equal: true},
		{name: "generate", it1: ints(4), it2: Slice[int]([]int{0, 1, 2, 3}), equal: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			eq, err := Equals(test.it1, test.it2, intEqual)
			assert.NoError(t, err)
			assert.Equal(t, test.equal, eq)
		})
	}
}

func TestEqualsPanic(t *testing.T) {
	l1 := Generate[int](20, func(n int) (int, error) { return n, nil })
	l2 := Generate[int](20, func(n int) (int, error) {
		if n == 10 {
			return 0, errors.New("test")
		}
		return n, nil
	})

	eq, err := Equals(l1, l2, intEqual)
	assert.Error(t, err)
	assert.False(t, eq)
}

func TestFirst(t *testing.T) {
	type testCase[V any] struct {
		name string
		it   Producer[V]
		want V
		ok   bool
	}
	tests := []testCase[int]{
		{name: "empty", it: Empty[int](), want: 0, ok: false},
		{name: "single", it: Single[int](2), want: 2, ok: true},
		{name: "slice", it: Slice[int]([]int{1, 2}), want: 1, ok: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := First(test.it)
			assert.EqualValues(t, test.want, got)
			assert.True(t, (err == nil) == test.ok)
		})
	}
}

func TestReduce(t *testing.T) {
	reduce, err := Reduce(ints(11), add)
	assert.NoError(t, err)
	assert.Equal(t, 55, reduce)
	reduce, err = Reduce(Empty[int](), add)
	assert.Error(t, err)
	assert.Equal(t, 0, reduce)
}

func TestReduceParallel(t *testing.T) {
	reduce, err := ReduceParallel(ints(11), func() func(int, int) (int, error) { return addSlow })
	assert.NoError(t, err)
	assert.Equal(t, 55, reduce)
	reduce, err = ReduceParallel(Empty[int](), func() func(int, int) (int, error) { return addSlow })
	assert.Error(t, err)
	assert.Equal(t, 0, reduce)
}

func TestReduceParallelPanic1(t *testing.T) {
	list := Generate[int](20, func(n int) (int, error) {
		if n == 10 {
			return 0, errors.New("test")
		}
		return n, nil
	})

	_, err := ReduceParallel(list, func() func(int, int) (int, error) { return addSlow })
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestReduceParallelPanic2(t *testing.T) {
	fire := runtime.NumCPU() * 2
	_, err := ReduceParallel(ints(10000), func() func(a int, b int) (int, error) {
		return func(a int, b int) (int, error) {
			time.Sleep(time.Millisecond * 30)
			if b == fire {
				log.Print("fire error")
				return 0, errors.New("test")
			}
			return a + b, nil
		}
	})
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestMapReduce(t *testing.T) {
	reduce, err := MapReduce[float64, int](ints(11), 0.0, func(s float64, i int) (float64, error) {
		return s + float64(i), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 55.0, reduce)
}

func checkBreak[V any](t *testing.T, in Producer[V], n int) {
	i := 0
	for _, err := range in {
		assert.NoError(t, err)
		i++
		if i == n {
			return
		}
	}
	assert.Equal(t, n, i, "not enough items")
}

func TestBreak(t *testing.T) {
	type testCase[V any] struct {
		name  string
		it    Producer[V]
		count int
	}
	appended := Append(Slice[int]([]int{1, 2}), Slice[int]([]int{3, 4}))
	toChan, _ := ToChan[int](ints(10))
	tests := []testCase[int]{
		{name: "slice", it: Slice[int]([]int{1, 2, 3}), count: 2},
		{name: "append", it: appended, count: 1},
		{name: "append", it: appended, count: 2},
		{name: "append", it: appended, count: 3},
		{name: "generate", it: Generate[int](10, addOne), count: 3},
		{name: "map", it: Map[int](ints(4), square), count: 2},
		{name: "parallelMap", it: MapParallel[int](ints(40), func() func(int, int) (int, error) { return squareSlow }), count: 10},
		{name: "filter", it: Filter[int](ints(8), isEven), count: 2},
		{name: "chan", it: FromChan(toChan), count: 2},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			checkBreak(t, test.it, test.count)
		})
	}
}

func FromChan[V any](src <-chan container[V]) Producer[V] {
	return func(yield Consumer[V]) {
		for v := range src {
			if !yield(v.val, v.err) {
				return
			}
		}
	}
}

func TestChannel(t *testing.T) {
	input := ints(10)
	toChan, _ := ToChan[int](input)
	res, err := ToSlice(FromChan(toChan))
	assert.NoError(t, err)
	assert.EqualValues(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, res)
}

func TestIterableCombine(t *testing.T) {
	check(t, Combine(Slice[int]([]int{1, 2, 3, 4}), add), 3, 5, 7)
	check(t, Combine(Slice[int]([]int{1}), add))
	check(t, Combine(Slice[int]([]int{}), add))
}

func add3(a, b, c int) (int, error) {
	return a + b + c, nil
}

func TestIterableCombine3(t *testing.T) {
	check(t, Combine3(Slice[int]([]int{1, 2, 3, 4, 5}), add3), 6, 9, 12)
	check(t, Combine3(Slice[int]([]int{1, 2, 3}), add3), 6)
	check(t, Combine3(Slice[int]([]int{1, 2}), add3))
	check(t, Combine3(Slice[int]([]int{1}), add3))
	check(t, Combine3(Slice[int]([]int{}), add3))
}

func add3n(pos int, v []int) (int, error) {
	return v[0] + v[1] + v[2], nil
}

func TestIterableCombineN(t *testing.T) {
	check(t, CombineN(Slice[int]([]int{1, 2, 3, 4, 5, 6}), 3, add3n), 6, 9, 12, 15)
	check(t, CombineN(Slice[int]([]int{1, 2, 3}), 3, add3n), 6)
	check(t, CombineN(Slice[int]([]int{1, 2}), 3, add3n))
	check(t, CombineN(Slice[int]([]int{1}), 3, add3n))
	check(t, CombineN(Slice[int]([]int{}), 3, add3n))
}

func TestIterableMerge(t *testing.T) {
	less := func(i1, i2 int) (bool, error) {
		return i1 < i2, nil
	}
	check(t, Merge(Slice[int]([]int{1, 3, 5}), Slice[int]([]int{2, 4, 6}), less), 1, 2, 3, 4, 5, 6)
	check(t, Merge(Slice[int]([]int{2, 4, 6}), Slice[int]([]int{1, 3, 5}), less), 1, 2, 3, 4, 5, 6)
	check(t, Merge(Slice[int]([]int{1, 2, 4}), Slice[int]([]int{3, 5, 6}), less), 1, 2, 3, 4, 5, 6)
	check(t, Merge(Slice[int]([]int{1, 2, 4, 6}), Slice[int]([]int{3, 5}), less), 1, 2, 3, 4, 5, 6)
	check(t, Merge(Slice[int]([]int{1, 3, 5}), Slice[int]([]int{2, 4, 6, 7, 8}), less), 1, 2, 3, 4, 5, 6, 7, 8)
	check(t, Merge(Slice[int]([]int{1, 3, 5, 7, 8}), Slice[int]([]int{2, 4, 6}), less), 1, 2, 3, 4, 5, 6, 7, 8)
}

func TestIterableMergeElements(t *testing.T) {
	combine := func(i1, i2 int) (int, error) {
		return i1 + i2, nil
	}
	check(t, MergeElements(Slice[int]([]int{1, 3, 5}), Slice[int]([]int{2, 4, 6}), combine), 3, 7, 11)

	checkErr(t, MergeElements(Slice[int]([]int{1, 3, 5}), Slice[int]([]int{2, 4, 6, 8}), combine), "same size", 3, 7, 11)

	checkErr(t, MergeElements(Slice[int]([]int{1, 3, 5, 7}), Slice[int]([]int{2, 4, 6}), combine), "same size", 3, 7, 11)
}

func TestIterableCross(t *testing.T) {
	cross := func(i1, i2 int) (int, error) {
		return i1 + i2, nil
	}
	check(t, Cross(Slice[int]([]int{10, 20, 30}), Slice[int]([]int{1, 2, 3}), cross), 11, 12, 13, 21, 22, 23, 31, 32, 33)
	check(t, Cross(Slice[int]([]int{1, 2, 3}), Slice[int]([]int{10, 20, 30}), cross), 11, 21, 31, 12, 22, 32, 13, 23, 33)
	check(t, Cross(Slice[int]([]int{1}), Slice[int]([]int{10}), cross), 11)
}

func TestIirMap(t *testing.T) {
	all := Generate[int](10, func(n int) (int, error) { return n + 1, nil })
	ints := IirMap(all, func(item int) (int, error) {
		return item, nil
	}, func(item int, lastItem int, last int) (int, error) {
		return item + last, nil
	})
	expected := Generate[int](10, func(n int) (int, error) { return (n + 2) * (n + 1) / 2, nil })

	equals, err := Equals(ints, expected, intEqual)
	assert.NoError(t, err)
	assert.True(t, equals)
}

func TestIterableFirst(t *testing.T) {
	ints := Slice[int]([]int{1, 2, 3, 4})
	check(t, FirstN(ints, 2), 1, 2)
	check(t, FirstN(ints, 6), 1, 2, 3, 4)
}

func TestIterableSkip(t *testing.T) {
	ints := Slice[int]([]int{1, 2, 3, 4})
	check(t, Skip(ints, 2), 3, 4)
	check(t, Skip(ints, 6))
}

func TestIterableThinning(t *testing.T) {
	ints := Slice[int]([]int{1, 2, 3, 4, 5, 6, 7})
	check(t, Thinning(ints, 1), 1, 3, 5, 7)
	check(t, Thinning(ints, 2), 1, 4, 7)
	check(t, Thinning(ints, 3), 1, 5, 7)
	check(t, Thinning(ints, 4), 1, 6, 7)
	check(t, Thinning(ints, 5), 1, 7)
	check(t, Thinning(ints, 10), 1, 7)
	ints = Slice[int]([]int{1, 2, 3, 4, 5, 6})
	check(t, Thinning(ints, 1), 1, 3, 5, 6)
	check(t, Thinning(ints, 2), 1, 4, 6)
	check(t, Thinning(ints, 3), 1, 5, 6)
	check(t, Thinning(ints, 4), 1, 6)
	check(t, Thinning(ints, 10), 1, 6)
}

func TestParallelMap(t *testing.T) {
	src := Generate[int](20, func(n int) (int, error) { return n, nil })
	ints := MapParallel(src, func() func(i, v int) (int, error) { return func(i, v int) (int, error) { return v * 2, nil } })
	expected, err := ToSlice(Generate[int](20, func(n int) (int, error) { return n * 2, nil }))
	assert.NoError(t, err)
	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, slice)
}

type mapResult struct {
	res    int
	number int
}

func TestParallelMapSlow(t *testing.T) {
	const count = 500
	const delay = time.Millisecond * 10
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := MapParallel(all, func() func(n, v int) (mapResult, error) {
		return func(n, v int) (mapResult, error) {
			time.Sleep(delay)
			return mapResult{res: v * 2, number: n}, nil
		}
	})
	expected, err := ToSlice(Generate[mapResult](count, func(n int) (mapResult, error) {
		return mapResult{res: (n + 1) * 2, number: n}, nil
	}))
	assert.NoError(t, err)

	start := time.Now()
	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, slice)

	measuredTime := time.Now().Sub(start)
	worstTime := time.Duration(count) * delay
	expectedTime := count * delay / time.Duration(runtime.NumCPU())

	fmt.Println("count:", count)
	fmt.Println("infinit cores:", delay)
	fmt.Println("single core:", worstTime)
	fmt.Println("expected", expectedTime)
	fmt.Println("measured:", measuredTime)

	assert.True(t, measuredTime < (expectedTime+worstTime)/2, "to slow")
}

func TestParallelFilter(t *testing.T) {
	ints := Slice[int]([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	ints = FilterParallel[int](ints, func() func(v int) (bool, error) {
		return func(v int) (bool, error) {
			time.Sleep(time.Millisecond * 10)
			return v%2 == 0, nil
		}
	})

	check[int](t, ints, 2, 4, 6, 8, 10)
}

func TestParallelMapPanic1(t *testing.T) {
	src := Generate[int](20, func(n int) (int, error) { return n, nil })
	ints := MapParallel[int, int](src, func() func(i, v int) (int, error) {
		return func(i, v int) (int, error) {
			if v == 7 {
				return 0, errors.New("test")
			}
			return v * 2, nil
		}
	})

	var err error
	for _, err = range ints {
		if err != nil {
			break
		}
	}
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestParallelMapPanic2(t *testing.T) {
	src := Generate[int](20, func(n int) (int, error) {
		if n == 7 {
			return 0, errors.New("test")
		}
		return n, nil
	})
	ints := MapParallel[int, int](src, func() func(i, v int) (int, error) {
		return func(i, v int) (int, error) {
			return v * 2, nil
		}
	})

	var err error
	for _, err = range ints {
		if err != nil {
			break
		}
	}
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestAutoMapFast(t *testing.T) {
	all := Generate[int](100000, func(n int) (int, error) { return n + 1, nil })

	start := time.Now()
	r1, _ := Reduce[int](MapAuto[int, int](all, func() func(n, v int) (int, error) {
		return func(n, v int) (int, error) {
			return v, nil
		}
	}), func(a int, b int) (int, error) {
		return a + b, nil
	})
	t1 := time.Now().Sub(start)

	start = time.Now()
	r2, _ := Reduce[int](Map[int, int](all, func(n, v int) (int, error) {
		return v, nil
	}), func(a int, b int) (int, error) {
		return a + b, nil
	})
	t2 := time.Now().Sub(start)

	assert.EqualValues(t, r1, r2)
	fmt.Println("time auto map", t1)
	fmt.Println("time simple map", t2)
	assert.True(t, t1 < t2*2, "auto map to slow with no load")
}

func TestAutoMapSlow(t *testing.T) {
	const count = itemsToMeasure * 50
	const delay = time.Millisecond * 10
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := MapAuto[int, mapResult](all, func() func(n, v int) (mapResult, error) {
		return func(n, v int) (mapResult, error) {
			time.Sleep(delay)
			return mapResult{res: v * 2, number: n}, nil
		}
	})
	expected, err := ToSlice(Generate[mapResult](count, func(n int) (mapResult, error) {
		return mapResult{res: (n + 1) * 2, number: n}, nil
	}))
	assert.NoError(t, err)

	start := time.Now()
	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	measuredTime := time.Now().Sub(start)
	worstTime := time.Duration(count) * delay
	expectedTime := time.Duration(itemsToMeasure)*delay + (count-itemsToMeasure)*delay/time.Duration(runtime.NumCPU())
	assert.EqualValues(t, expected, slice)

	fmt.Println("count:", count)
	fmt.Println("infinite cores:", time.Duration(itemsToMeasure+1)*delay)
	fmt.Println("single core:", worstTime)
	fmt.Println("expected", expectedTime)
	fmt.Println("measured:", measuredTime)
	assert.True(t, measuredTime < (expectedTime+worstTime)/2, "auto map to slow with heavy load")

	ints = MapAuto[int, mapResult](all, func() func(n, v int) (mapResult, error) {
		return func(n, v int) (mapResult, error) {
			return mapResult{res: v * 2, number: n}, nil
		}
	})
	toSlice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, toSlice)
}

func TestAutoMapShort(t *testing.T) {
	count := itemsToMeasure
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := MapAuto[int, int](all, func() func(n, v int) (int, error) {
		return func(n, v int) (int, error) {
			return v * 2, nil
		}
	})
	expected, err := ToSlice(Generate[int](count, func(n int) (int, error) { return (n + 1) * 2, nil }))
	assert.NoError(t, err)
	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, slice)
	toSlice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, toSlice)
}

func TestAutoMapPanicEarly1(t *testing.T) {
	src := Generate[int](20, func(n int) (int, error) {
		if n == 5 {
			return 0, errors.New("test")
		}
		return n, nil
	})
	ints := MapAuto[int, int](src, func() func(i, v int) (int, error) {
		return func(i, v int) (int, error) {
			return v * 2, nil
		}
	})

	var err error
	for _, err = range ints {
		if err != nil {
			break
		}
	}
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestAutoMapPanicEarly2(t *testing.T) {
	src := Generate[int](20, func(n int) (int, error) { return n, nil })
	ints := MapAuto[int, int](src, func() func(i, v int) (int, error) {
		return func(i, v int) (int, error) {
			if i == 4 {
				return 0, errors.New("test")
			}
			return v * 2, nil
		}
	})

	var err error
	for _, err = range ints {
		if err != nil {
			break
		}
	}
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestAutoMapPanic(t *testing.T) {
	src := Generate[int](200, func(n int) (int, error) {
		return n, nil
	})
	ints := MapAuto[int, int](src, func() func(i, v int) (int, error) {
		return func(i, v int) (int, error) {
			time.Sleep(5 * time.Millisecond)
			if i == 100 {
				return 0, errors.New("test")
			}
			return v * 2, nil
		}
	})

	var err error
	for _, err = range ints {
		if err != nil {
			break
		}
	}
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestAutoFilter(t *testing.T) {
	const count = itemsToMeasure * 50
	const delay = time.Millisecond * 10
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := FilterAuto[int](all, func() func(v int) (bool, error) {
		return func(v int) (bool, error) {
			time.Sleep(delay)
			return v%2 == 0, nil
		}
	})
	expected, err := ToSlice(Generate[int](count/2, func(n int) (int, error) { return (n + 1) * 2, nil }))
	assert.NoError(t, err)

	start := time.Now()
	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	measuredTime := time.Now().Sub(start)
	worstTime := time.Duration(count) * delay
	expectedTime := time.Duration(itemsToMeasure)*delay + (count-itemsToMeasure)*delay/time.Duration(runtime.NumCPU())
	assert.EqualValues(t, expected, slice)

	fmt.Println("count:", count)
	fmt.Println("infinite cores:", time.Duration(itemsToMeasure+1)*delay)
	fmt.Println("single core:", worstTime)
	fmt.Println("expected", expectedTime)
	fmt.Println("measured:", measuredTime)

	assert.True(t, measuredTime < (expectedTime+worstTime)/2, "to slow")

	assert.EqualValues(t, expected, slice)
	toSlice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, toSlice)
}

func TestAutoFilterSerial(t *testing.T) {
	count := itemsToMeasure * 2
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := FilterAuto(all, func() func(v int) (bool, error) {
		return func(v int) (bool, error) {
			return v%2 == 0, nil
		}
	})
	expected, err := ToSlice(Generate[int](count/2, func(n int) (int, error) { return (n + 1) * 2, nil }))
	assert.NoError(t, err)

	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, slice)
	toSlice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, toSlice)
}

func TestAutoFilterShort(t *testing.T) {
	count := itemsToMeasure
	all := Generate[int](count, func(n int) (int, error) { return n + 1, nil })
	ints := FilterAuto(all, func() func(v int) (bool, error) {
		return func(v int) (bool, error) {
			time.Sleep(time.Microsecond * itemProcessingTimeMicroSec * 2)
			return v%2 == 0, nil
		}
	})
	expected, err := ToSlice(Generate[int](count/2, func(n int) (int, error) { return (n + 1) * 2, nil }))
	assert.NoError(t, err)

	slice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, slice)
	toSlice, err := ToSlice(ints)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, toSlice)
}

func TestCompact(t *testing.T) {
	type testCase struct {
		name string
		it   Producer[int]
		want []int
	}
	var empty []int
	tests := []testCase{
		{name: "no dup", it: Slice[int]([]int{1, 2, 3, 4, 5, 6, 7}), want: []int{1, 2, 3, 4, 5, 6, 7}},
		{name: "normal", it: Slice[int]([]int{1, 1, 2, 2, 3, 4, 4}), want: []int{1, 2, 3, 4}},
		{name: "normal", it: Slice[int]([]int{1, 1, 2, 2, 3, 4}), want: []int{1, 2, 3, 4}},
		{name: "all same", it: Slice[int]([]int{1, 1, 1, 1, 1, 1}), want: []int{1}},
		{name: "empty", it: Slice[int]([]int{}), want: empty},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			slice, err := ToSlice(Compact[int, int](test.it, func(i int) (int, error) {
				return i, nil
			}, equal[int]))
			assert.NoError(t, err)
			assert.EqualValues(t, test.want, slice)
		})
	}
}

func TestGroup(t *testing.T) {
	type testCase struct {
		name string
		it   Producer[int]
		want [][]int
	}
	var empty [][]int
	tests := []testCase{
		{name: "no dup", it: Slice[int]([]int{1, 2, 3, 4, 5, 6, 7}), want: [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}}},
		{name: "normal", it: Slice[int]([]int{1, 1, 2, 2, 3, 4, 4}), want: [][]int{{1, 1}, {2, 2}, {3}, {4, 4}}},
		{name: "normal", it: Slice[int]([]int{1, 1, 2, 2, 3, 4}), want: [][]int{{1, 1}, {2, 2}, {3}, {4}}},
		{name: "all same", it: Slice[int]([]int{1, 1, 1, 1, 1, 1}), want: [][]int{{1, 1, 1, 1, 1, 1}}},
		{name: "empty", it: Slice[int]([]int{}), want: empty},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := ToSlice(Group(test.it, equal[int]))
			assert.NoError(t, err)
			assert.Equal(t, len(test.want), len(got))
			for i, w := range test.want {
				g, err := ToSlice(got[i])
				assert.NoError(t, err)
				assert.EqualValues(t, w, g)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	const size = 100
	ints := Generate[int](size, func(i int) (int, error) { return i, nil })

	split, run, done := CopyProducer[int](10)

	for i := 0; i < 10; i++ {
		j := i
		go func() {
			got, err := ToSlice(split[j])
			assert.NoError(t, err)
			assert.Equal(t, size, len(got))
			for k, g := range got {
				assert.Equal(t, k, g)
			}
			done(err)
		}()
	}
	err := run(ints)
	assert.NoError(t, err)
}

func TestSplitTerminate(t *testing.T) {
	const size = 100
	ints := Generate[int](size, func(i int) (int, error) { return i, nil })

	split, run, done := CopyProducer[int](10)

	for i := 0; i < 10; i++ {
		j := i
		go func() {
			got, err := ToSlice(FirstN(split[j], j+1))
			assert.NoError(t, err)
			assert.Equal(t, j+1, len(got))
			for k, g := range got {
				assert.Equal(t, k, g)
			}
			done(err)
		}()
	}
	err := run(ints)
	assert.NoError(t, err)
}

func TestSplitError(t *testing.T) {
	const size = 100
	ints := Generate[int](size, func(i int) (int, error) {
		if i == 50 {
			return 0, errors.New("test")
		}
		return i, nil
	})

	split, run, done := CopyProducer[int](10)

	for i := 0; i < 10; i++ {
		j := i
		go func() {
			got, err := ToSlice(split[j])
			//assert.Error(t, err)
			assert.Equal(t, 50, len(got))
			for k, g := range got {
				assert.Equal(t, k, g)
			}
			done(err)
		}()
	}
	err := run(ints)
	assert.Error(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestSplitDontUse(t *testing.T) {
	const size = 100
	ints := Generate[int](size, func(i int) (int, error) { return i, nil })

	_, run, done := CopyProducer[int](10)

	for i := 0; i < 10; i++ {
		done(nil)
	}

	err := run(ints)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "timed out"))
}

func BenchmarkMap(b *testing.B) {
	ints := Generate[int](10000, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(Map[int, int](ints, square), mapAdd)
	}
}

func mapAdd(a, b int) (int, error) {
	return a + b, nil
}

func BenchmarkMapAuto(b *testing.B) {
	ints := Generate[int](10000, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(MapAuto(ints, func() func(int, int) (int, error) {
			return square
		}), mapAdd)
	}
}

func BenchmarkMapParallel(b *testing.B) {
	ints := Generate[int](10000, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(MapParallel(ints, func() func(int, int) (int, error) {
			return square
		}), add)
	}
}

func BenchmarkSMap(b *testing.B) {
	ints := Generate[int](100, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(Map(ints, squareSlow), add)
	}
}

func BenchmarkSMapAuto(b *testing.B) {
	ints := Generate[int](100, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(MapAuto(ints, func() func(int, int) (int, error) {
			return squareSlow
		}), add)
	}
}

func BenchmarkSMapParallel(b *testing.B) {
	ints := Generate[int](100, func(i int) (int, error) { return i, nil })
	for i := 0; i < b.N; i++ {
		Reduce(MapParallel(ints, func() func(int, int) (int, error) {
			return squareSlow
		}), add)
	}
}
