package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"time"
)

const startTimeout = 2 // seconds

// MultiUse takes a map of closures and the list is passed to the closures. The
// return values of the closures are returned in a map. The keys in the result
// map are the same keys used to pass the closures. MultiUse is useful if you
// have to use the same list multiple times and the list is both expensive to
// create and expensive to store. This is because MultiUse allows you to use the
// list multiple times without having to store the list elements for later reuse.
func (l *List) MultiUse(st funcGen.Stack[Value]) Map {
	if m, ok := st.Get(1).ToMap(); ok {
		var muList multiUseList
		m.Iter(func(key string, value Value) bool {
			if f, ok := value.ToClosure(); ok {
				if f.Args == 1 {
					muList = append(muList, &multiUseEntry{name: key, fu: f.Func})
				} else {
					panic("map in multiUse needs to contain closures with one argument")
				}
			} else {
				panic("map in multiUse need to contain closures")
			}
			return true
		})

		muList.runConsumerClosures()

		errChan := muList.runProducer(l.Iterator())

		return muList.createResult(errChan)
	} else {
		panic("first argument in multiUse needs to be a map")
	}
}

type multiUseEntry struct {
	name         string
	fu           funcGen.Func[Value]
	writer       chan<- Value
	requestClose <-chan struct{}
	result       <-chan multiUseResult
}

type multiUseList []*multiUseEntry

// createIterable creates an iterable based on the writer chanel. If the yield
// function returns false, the requestClose channel is closed which will cause
// the producer to stop sending values to this iterable.
func (mu *multiUseEntry) createIterable(started chan<- struct{}) iterator.Iterable[Value] {
	return func() iterator.Iterator[Value] {
		if mu.writer != nil {
			panic(fmt.Errorf("list passed to multiUse closure %s can only be used once", mu.name))
		}
		r := make(chan Value)
		c := make(chan struct{})
		mu.writer = r
		mu.requestClose = c
		started <- struct{}{}
		return func(yield func(Value) bool) bool {
			for v := range r {
				if !yield(v) {
					close(c)
					return false
				}
			}
			close(c)
			return true
		}
	}
}

type multiUseResult struct {
	result Value
	err    error
}

// runConsumer calls the closure and sends the result to the result channel. If
// the closure panics, the panic is recovered and also sent to the result
// channel. if the closure returns a list, the list is evaluated before it is
// sent to the result channel.
func (mu *multiUseEntry) runConsumer(started chan struct{}) {
	r := make(chan multiUseResult)
	mu.result = r
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				r <- multiUseResult{result: nil, err: parser2.AnyToError(rec)}
			}
			close(r)
		}()
		st := funcGen.NewEmptyStack[Value]()
		st.Push(NewListFromIterable(mu.createIterable(started)))
		value := mu.fu(st, nil)
		if list, ok := value.(*List); ok {
			// Force evaluation of lists. Lazy evaluation is not possible here because it is
			// not possible to iterate over a list at any time. Iteration is only possible
			// synchronously with all iterators at the same time.
			list.Eval()
		}
		r <- multiUseResult{result: value, err: nil}
	}()
}

// runProducer runs the iterator of the source list and sends the values to all
// the destination lists. If a destination lists yield returns false it closes
// its requestClose channel which will cause this method to stop sending values to this
// destination list. If all destination lists yield functions have returned
// false, also the source iterator returns false, which stops the iteration.
func (ml multiUseList) runProducer(i iterator.Iterator[Value]) <-chan error {
	errChan := make(chan error)
	go func() {
		defer func() {
			// recover a panic and send it to the error channel
			if rec := recover(); rec != nil {
				errChan <- parser2.AnyToError(rec)
			}
			// If a panic occurs, the writers are not closed. This ensures that the consumers
			// do not stop waiting for data and therefore do not send a result. This ensures
			// that the error sent above is actually received.
		}()
		running := len(ml)
		i(func(v Value) bool {
			for _, mu := range ml {
				if mu.writer != nil {
					select {
					case mu.writer <- v:
					case <-mu.requestClose:
						running--
						close(mu.writer)
						mu.writer = nil
					}
				}
			}
			return running > 0
		})
		for _, mu := range ml {
			if mu.writer != nil {
				close(mu.writer)
			}
		}
	}()
	return errChan
}

// runConsumerClosures runs all the consumer closures and waits for them to be
// started. Started means that the closure has requested an iterator to iterate
// over the list.
func (ml multiUseList) runConsumerClosures() {
	started := make(chan struct{})
	for _, mu := range ml {
		mu.runConsumer(started)
	}

	// wait for all consumers to be started
	for i := 0; i < len(ml); i++ {
		select {
		case <-time.After(startTimeout * time.Second):
			panic(ml.timeOutError())
		case <-started:
		}
	}
}

// createResult creates the result map. If the source iterator panics, the panic
// is rethrown. If one of the closures panics, the panic is recovered and also
// rethrown. This method needs to be called in the main thread. It is the only
// part that does not run in its own goroutine.
func (ml multiUseList) createResult(errChan <-chan error) Map {
	resultMap := listMap.New[Value](len(ml))
	for _, mu := range ml {
		select {
		case result := <-mu.result:
			if result.err != nil {
				panic(result.err)
			}
			resultMap = resultMap.Append(mu.name, result.result)
		case err := <-errChan:
			panic(err)
		}
	}
	return NewMap(resultMap)
}

// timeOutError creates the error message if one of the closures does not request
// its iterator.
func (ml multiUseList) timeOutError() error {
	var buffer bytes.Buffer
	buffer.WriteString("list passed to closure is not used; affected closure(s): ")
	first := true
	for _, mu := range ml {
		if mu.writer == nil {
			if first {
				first = false
			} else {
				buffer.WriteString(", ")
			}
			buffer.WriteString(mu.name)
		}
	}
	return errors.New(buffer.String())
}
