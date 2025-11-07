package iterator

import (
	"errors"
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
				return
			}
		}
		close(c)
	}()
	return c, done
}

func FilterAuto[V any](p Producer[V], acceptFac func() func(v V) (bool, error)) Producer[V] {
	return func(yield Consumer[V]) {
		accFunc := acceptFac()
		for i, err := range p {
			acc := true
			if err == nil {
				acc, err = accFunc(i)
			}
			if acc || err != nil {
				if !yield(i, err) {
					return
				}
			}
		}
	}
}

func MapAuto[I, O any](p Producer[I], mapperFac func() func(i int, v I) (O, error)) Producer[O] {
	return func(yield Consumer[O]) {
		mapper := mapperFac()
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
		if yield(c.val, c.err) {
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
	ack := make(chan error)
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
