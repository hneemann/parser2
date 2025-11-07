package iterator

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

type Consumer[V any] func(V, error) bool

type Producer[V any] func(Consumer[V])

func Empty[V any]() Producer[V] {
	return func(yield Consumer[V]) {
	}
}

func Single[V any](v V) Producer[V] {
	return func(yield Consumer[V]) {
		yield(v, nil)
	}
}

// Slice create an Iterable from a slice
func Slice[V any](items []V) Producer[V] {
	return func(yield Consumer[V]) {
		for _, i := range items {
			if !yield(i, nil) {
				return
			}
		}
	}
}

func First[V any](it Producer[V]) (V, error) {
	for v, err := range it {
		return v, err
	}
	var v V
	return v, errors.New("empty iterator")
}

// ToSlice reads all items from the Producer and stores them in a slice.
func ToSlice[V any](it Producer[V]) ([]V, error) {
	var sl []V
	for v, err := range it {
		if err != nil {
			return sl, err
		}
		sl = append(sl, v)
	}
	return sl, nil
}

type container[V any] struct {
	num int
	val V
	err error
}

// ToChan writes elements to a channel
func ToChan[V any](it Producer[V]) (<-chan container[V], chan struct{}) {
	c := make(chan container[V])
	done := make(chan struct{})
	go func() {
		i := 0
		for v, err := range it {
			select {
			case c <- container[V]{num: i, val: v, err: err}:
				i++
			case <-done:
				break
			}
		}
		close(c)
	}()
	return c, done
}

// Equals checks if the two Iterators are equal.
func Equals[V any](i1, i2 Producer[V], equals func(V, V) (bool, error)) (bool, error) {
	ch1, done1 := ToChan(i1)
	ch2, done2 := ToChan(i2)
	defer func() {
		close(done1)
		close(done2)
	}()

	for {
		c1, ok1 := <-ch1
		c2, ok2 := <-ch2
		if ok1 && ok2 {
			if c1.err != nil {
				return false, c1.err
			}
			if c2.err != nil {
				return false, c2.err
			}
			eq, err := equals(c1.val, c2.val)
			if err != nil {
				return false, err
			}
			if !eq {
				return false, nil
			}
		} else {
			return ok1 == ok2, nil
		}
	}
}

func Filter[V any](p Producer[V], accept func(v V) (bool, error)) Producer[V] {
	return func(yield Consumer[V]) {
		for i, err := range p {
			acc := true
			if err == nil {
				acc, err = accept(i)
			}
			if acc || err != nil {
				if !yield(i, err) {
					return
				}
			}
		}
	}
}
func FilterParallel[V any](p Producer[V], acceptFac func() func(v V) (bool, error)) Producer[V] {
	if runtime.NumCPU() == 1 {
		return Filter(p, acceptFac())
	}

	type filterContainer[V any] struct {
		val    V
		accept bool
	}

	return func(yield Consumer[V]) {
		m := MapParallel(p, func() func(i int, val V) (filterContainer[V], error) {
			accept := acceptFac()
			return func(i int, val V) (filterContainer[V], error) {
				b, err := accept(val)
				if err != nil {
					return filterContainer[V]{}, err
				}
				return filterContainer[V]{val, b}, nil
			}
		})
		for fc, err := range m {
			if fc.accept || err != nil {
				if !yield(fc.val, err) {
					return
				}
			}
		}
	}
}

func FilterAuto[V any](p Producer[V], acceptFac func() func(v V) (bool, error)) Producer[V] {
	if runtime.NumCPU() == 1 {
		return Filter(p, acceptFac())
	}
	return FilterParallel(p, acceptFac)
}

const (
	itemProcessingTimeMicroSec = 200
	itemsToMeasure             = 11
)

// Map maps the elements to new element created by the given mapFunc function
func Map[I, O any](p Producer[I], mapper func(i int, v I) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		i := 0
		for item, err := range p {
			var o O
			if err == nil {
				o, err = mapper(i, item)
			}
			if !yield(o, err) {
				return
			}
			i++
		}
	}
}

func MapParallel[I, O any](p Producer[I], mapperFac func() func(i int, v I) (O, error)) Producer[O] {
	if runtime.NumCPU() == 1 {
		return Map(p, mapperFac())
	}

	return func(yield Consumer[O]) {
		c, done := ToChan(p)
		doneOpen := true
		result := make(chan container[O])
		wg := sync.WaitGroup{}
		for range runtime.NumCPU() {
			wg.Add(1)
			mf := mapperFac()
			go func() {
				for item := range c {
					var o O
					var err error
					if item.err != nil {
						err = item.err
					} else {
						o, err = mf(item.num, item.val)
					}
					result <- container[O]{num: item.num, val: o, err: err}
				}
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(result)
		}()

		var err error
		nextOut := 0
		buffer := make(map[int]container[O])
		for r := range result {
			if r.err != nil && err == nil {
				if doneOpen {
					doneOpen = false
					close(done)
				}
				err = r.err
			}
			if r.num == nextOut {
				if !yield(r.val, err) {
					if doneOpen {
						doneOpen = false
						close(done)
					}
					return
				}
				nextOut++
				for {
					if b, ok := buffer[nextOut]; ok {
						if !yield(b.val, b.err) {
							if doneOpen {
								doneOpen = false
								close(done)
							}
							return
						}
						delete(buffer, nextOut)
						nextOut++
					} else {
						break
					}
				}
			} else {
				buffer[r.num] = r
			}
		}
	}
}

func MapAuto[I, O any](p Producer[I], mapperFac func() func(i int, v I) (O, error)) Producer[O] {
	if runtime.NumCPU() == 1 {
		return Map(p, mapperFac())
	}

	return MapParallel(p, mapperFac)
}

func Cross[I1, I2, O any](i1 Producer[I1], i2 Producer[I2], crossFunc func(i1 I1, i2 I2) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		var o O
		for i1v, err := range i1 {
			if err != nil {
				if !yield(o, err) {
					return
				}
			}
			for i2v, err := range i2 {
				if err == nil {
					o, err = crossFunc(i1v, i2v)
				}
				if !yield(o, err) {
					return
				}
			}
		}
	}
}

// Compact returns an iterable which contains no consecutive duplicates.
func Compact[V, M any](items Producer[V], convert func(V) (M, error), equal func(M, M) (bool, error)) Producer[V] {
	return func(yield Consumer[V]) {
		isLast := false
		var last M
		for v, err := range items {
			var val M
			if err == nil {
				val, err = convert(v)
			}
			if isLast && err == nil {
				eq := false
				eq, err = equal(last, val)
				if !eq {
					if !yield(v, err) {
						return
					}
				}
			} else {
				isLast = true
				if !yield(v, err) {
					return
				}
			}
			last = val
		}
	}
}

// Group returns an iterable which contains iterables of equal values
func Group[V any](items Producer[V], equal func(V, V) (bool, error)) Producer[Producer[V]] {
	return func(yield Consumer[Producer[V]]) {
		var list []V
		for v, err := range items {
			if len(list) > 0 || err != nil {
				eq := false
				if err == nil {
					eq, err = equal(list[len(list)-1], v)
				}
				if eq {
					list = append(list, v)
				} else {
					if !yield(Slice[V](list), err) {
						return
					}
					list = []V{v}
				}
			} else {
				list = []V{v}
			}
		}
		if len(list) > 0 {
			yield(Slice[V](list), nil)
		}
	}
}

// Thinning returns an iterable which skips a certain amount of elements
// from the parent iterable. If skip is set to 1, every second element is skipped.
// The first and the last item are always returned.
func Thinning[V any](items Producer[V], n int) Producer[V] {
	return func(yield Consumer[V]) {
		i := 0
		var skipped V
		for v, err := range items {
			if i == 0 || err != nil {
				i = n
				if !yield(v, err) {
					return
				}
			} else {
				skipped = v
				i--
			}
		}
		if i < n {
			yield(skipped, nil)
		}
	}
}

func MergeElements[I1, I2, O any](it1 Producer[I1], it2 Producer[I2], combine func(i1 I1, i2 I2) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		aMain, aStop := ToChan(it1)
		bMain, bStop := ToChan(it2)
		defer func() {
			close(aStop)
			close(bStop)
		}()
		for {
			a, aOk := <-aMain
			b, bOk := <-bMain

			if aOk && bOk {
				var err error
				if a.err != nil {
					err = a.err
				} else if b.err != nil {
					err = b.err
				}
				var o O
				if err == nil {
					o, err = combine(a.val, b.val)
				}
				if !yield(o, err) {
					return
				}
			} else if aOk || bOk {
				var o O
				yield(o, errors.New("iterables in mergeElements dont have the same size"))
				return
			} else {
				return
			}
		}
	}
}

// Merge is used to merge two iterables.
// The less function determines which element to take first
// Makes sens only if the provided iterables are ordered.
func Merge[V any](ai, bi Producer[V], less func(V, V) (bool, error)) Producer[V] {
	return func(yield Consumer[V]) {
		aMain, aStop := ToChan(ai)
		bMain, bStop := ToChan(bi)
		defer func() {
			close(aStop)
			close(bStop)
		}()
		isA := false
		var a container[V]
		isB := false
		var b container[V]
		for {
			if !isA {
				a, isA = <-aMain
				if !isA {
					if isB {
						if !yield(b.val, b.err) {
							return
						}
					}
					copyValues(bMain, yield)
					return
				}
			}
			if !isB {
				b, isB = <-bMain
				if !isB {
					if isA {
						if !yield(a.val, a.err) {
							return
						}
					}
					copyValues(aMain, yield)
					return
				}
			}
			var err error
			if a.err != nil {
				err = a.err
			} else if b.err != nil {
				err = b.err
			}
			var lessA bool
			if err == nil {
				lessA, err = less(a.val, b.val)
			}
			if lessA {
				if !yield(a.val, err) {
					return
				}
				isA = false
			} else {
				if !yield(b.val, err) {
					return
				}
				isB = false
			}
		}
	}
}

func copyValues[V any](main <-chan container[V], yield Consumer[V]) {
	for c := range main {
		if !yield(c.val, c.err) {
			return
		}
	}
}

func Combine[I, O any](p Producer[I], combine func(I, I) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		isValue := false
		var last I
		var o O
		for i, err := range p {
			if isValue {
				if err == nil {
					o, err = combine(last, i)
				}
				if !yield(o, err) {
					return
				}
			} else {
				if err != nil {
					if !yield(o, err) {
						return
					}
				}
				isValue = true
			}
			last = i
		}
	}
}

func Combine3[I, O any](p Producer[I], combine func(I, I, I) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		valuesPresent := 0
		var last I
		var lastLast I
		var o O
		for i, err := range p {
			if err != nil {
				if !yield(o, err) {
					return
				}
			}
			if valuesPresent == 0 {
				valuesPresent++
				lastLast = i
			} else if valuesPresent == 1 {
				valuesPresent++
				last = i
			} else {
				if !yield(combine(lastLast, last, i)) {
					return
				}
				lastLast = last
				last = i
			}
		}
	}
}

// CombineN maps N consecutive elements to a new element.
// The generated iterable has (N-1) elements less than the original iterable.
func CombineN[I, O any](in Producer[I], n int, combine func(int, []I) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		valuesPresent := 0
		pos := 0
		vals := make([]I, n, n)
		for i, err := range in {
			if err != nil {
				var o O
				if !yield(o, err) {
					return
				}
			}
			vals[pos] = i
			pos++
			if pos == n {
				pos = 0
			}
			if valuesPresent < n {
				valuesPresent++
			}
			if valuesPresent == n {
				if !yield(combine(pos, vals)) {
					return
				}
			}
		}
	}
}

// IirMap maps a value, the last value and the last created value to a new element.
// Can be used to implement iir filters like a low-pass. The last item is provided
// to allow handling of non-equidistant values.
func IirMap[I, R any](items Producer[I], initial func(item I) (R, error), iir func(item I, lastItem I, last R) (R, error)) Producer[R] {
	return func(yield Consumer[R]) {
		isLast := false
		var lastItem I
		var last R
		for i, err := range items {
			if err == nil {
				if isLast {
					last, err = iir(i, lastItem, last)
				} else {
					last, err = initial(i)
					isLast = true
				}
			}
			lastItem = i
			if !yield(last, err) {
				return
			}
		}
	}
}

// FirstN returns the first n elements of an Iterable
func FirstN[V any](items Producer[V], n int) Producer[V] {
	return func(yield Consumer[V]) {
		i := 0
		for v, err := range items {
			if i == n {
				return
			}
			if !yield(v, err) {
				return
			}
			i++
		}
	}
}

// Skip skips the first elements.
// The number of elements to skip is given in skip.
func Skip[V any](items Producer[V], n int) Producer[V] {
	return func(yield Consumer[V]) {
		i := 0
		for v, err := range items {
			if i < n {
				i++
				if err != nil {
					if !yield(v, err) {
						return
					}
				}
			} else {
				if !yield(v, err) {
					return
				}
			}
		}
	}
}

// Reduce reduces the items of the iterable to a single value by calling the reduce function.
func Reduce[V any](it Producer[V], reduceFunc func(V, V) (V, error)) (V, error) {
	var sum V
	isValue := false
	for v, err := range it {
		if err != nil {
			return sum, err
		}
		if isValue {
			sum, err = reduceFunc(sum, v)
			if err != nil {
				return sum, err
			}
		} else {
			sum = v
			isValue = true
		}
	}
	if !isValue {
		return sum, errors.New("reduce on empty iterable")
	}
	return sum, nil
}

// ReduceParallel reduces the items of the iterable to a single value by calling the reduce function.
func ReduceParallel[V any](it Producer[V], reduceFac func() func(V, V) (V, error)) (V, error) {
	return Reduce(it, reduceFac())
}

// MapReduce combines a map and reduce step in one go.
// Avoids generating intermediate map results.
// Instead of map(n->n^2).reduce((a,b)->a+b) one
// can write  mapReduce(0, (s,n)->s+n^2)
// Useful if map and reduce are both low cost operations.
func MapReduce[S, V any](it Producer[V], initial S, reduceFunc func(S, V) (S, error)) (S, error) {
	for v, err := range it {
		if err != nil {
			return initial, err
		}
		var err error
		initial, err = reduceFunc(initial, v)
		if err != nil {
			return initial, err
		}
	}
	return initial, nil
}

func Append[V any](it1 Producer[V], it2 Producer[V]) Producer[V] {
	return func(yield Consumer[V]) {
		for v, err := range it1 {
			if !yield(v, err) {
				return
			}
		}
		for v, err := range it2 {
			if !yield(v, err) {
				return
			}
		}
	}
}

func Generate[V any](n int, gen func(i int) (V, error)) Producer[V] {
	return func(yield Consumer[V]) {
		for i := 0; i < n; i++ {
			if !yield(gen(i)) {
				return
			}
		}
	}
}

// CopyProducer copies the initial producer into num identical producers.
// The returned producers can be used in parallel to process the input in.
//
// The return values are:
//
//	The prodList contains the copied producers.
//	The run function needs to be called with the consumer context to start reading the given producer.
func CopyProducer[V any](num int) ([]Producer[V], func(in Producer[V]) error, func(error)) {

	type data struct {
		v   V
		err error
	}

	type holder struct {
		c    chan data
		stop chan struct{}
		done bool
	}

	holders := make([]*holder, num)
	for i := 0; i < num; i++ {
		holders[i] = &holder{
			c:    make(chan data),
			stop: make(chan struct{}),
		}
	}

	prodList := make([]Producer[V], num)
	for i := 0; i < num; i++ {
		ho := holders[i]
		prodList[i] = func(yield Consumer[V]) {
			defer close(ho.stop)
			for d := range ho.c {
				if !yield(d.v, d.err) {
					return
				}
			}
		}
	}
	errorTerm := make(chan struct{})
	errorTermOpen := true
	ack := make(chan error, num)
	run := func(in Producer[V]) error {
	outer:
		for v, err := range in {
			d := data{v: v, err: err}
			wasDone := false
			for _, h := range holders {
				select {
				case <-errorTerm:
					break outer
				case <-time.After(time.Second * 5):
					for _, htc := range holders {
						close(htc.c)
					}
					return errors.New("iterator timed out")
				case h.c <- d:
				case <-h.stop:
					h.done = true
					wasDone = true
				}
			}
			if wasDone {
				nh := make([]*holder, 0, len(holders))
				for _, h := range holders {
					if h.done {
						close(h.c)
					} else {
						nh = append(nh, h)
					}
				}
				holders = nh
				if len(holders) == 0 {
					break
				}
			}
		}
		for _, h := range holders {
			close(h.c)
		}
		var err error
		for range num {
			aErr := <-ack
			if aErr != nil {
				err = aErr
			}
		}
		return err
	}
	done := func(err error) {
		if err != nil {
			if errorTermOpen {
				close(errorTerm)
				errorTermOpen = false
			}
		}
		ack <- err
	}

	return prodList, run, done
}
