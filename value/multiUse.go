package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"log"
	"sync"
	"time"
)

const startTimeout = 2 // seconds

// MultiUse takes a map of closures and the list is passed to the closures. The
// return values of the closures are returned in a map. The keys in the result
// map are the same keys used to pass the closures. MultiUse is useful if you
// have to use the same list multiple times and the list is both expensive to
// create and expensive to store. This is because MultiUse allows you to use the
// list multiple times without having to store the list elements for later reuse.
func (l *List) MultiUse(st funcGen.Stack[Value]) (Map, error) {
	if m, ok := st.Get(1).ToMap(); ok {
		var muList multiUseList
		var innerErr error
		m.Iter(func(key string, value Value) bool {
			if f, ok := value.ToClosure(); ok {
				if f.Args == 1 {
					muList = append(muList, &multiUseEntry{name: key, fu: f.Func})
				} else {
					innerErr = errors.New("map in multiUse needs to contain functions with one argument")
					return false
				}
			} else {
				innerErr = errors.New("map in multiUse need to contain functions")
				return false
			}
			return true
		})
		if innerErr != nil {
			return Map{}, innerErr
		}

		err := muList.runConsumerClosures()
		if err != nil {
			return Map{}, err
		}

		errChan := muList.runProducer(l.Iterator(st))

		return muList.createResult(errChan)
	} else {
		return Map{}, errors.New("first argument in multiUse needs to be a map")
	}
}

type multiUseEntry struct {
	name            string
	fu              funcGen.Func[Value]
	writerLock      sync.Mutex
	writer          chan<- Value
	requestClose    chan struct{}
	requestIsClosed bool
	result          <-chan multiUseResult
}

type multiUseList []*multiUseEntry

// createIterable creates an iterable based on the writer chanel. If the yield
// function returns false, the requestClose channel is closed which will cause
// the producer to stop sending values to this iterable.
func (mu *multiUseEntry) createIterable(started chan<- string) iterator.Iterable[Value, funcGen.Stack[Value]] {
	return func(st funcGen.Stack[Value]) iterator.Iterator[Value] {
		if mu.writer != nil {
			return func(yield func(Value) bool) (bool, error) {
				return false, fmt.Errorf("list passed to multiUse function %s can only be used once", mu.name)
			}
		}
		r := make(chan Value)
		mu.writer = r
		mu.requestClose = make(chan struct{})
		started <- mu.name
		return func(yield func(Value) bool) (bool, error) {
			for v := range r {
				if !yield(v) {
					mu.stopWriter()
					return false, nil
				}
			}
			mu.stopWriter()
			return true, nil
		}
	}
}

func (mu *multiUseEntry) stopWriter() {
	if !mu.requestIsClosed && mu.requestClose != nil {
		close(mu.requestClose)
		mu.requestIsClosed = true
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
func (mu *multiUseEntry) runConsumer(started chan string) {
	r := make(chan multiUseResult)
	mu.result = r
	go func() {
		defer func() {
			mu.stopWriter()
			if rec := recover(); rec != nil {
				log.Print("panic in multiUse consumer: ", rec)
				// If start is not reported yet, do now. This happens if evaluation of function
				// fails before the iterator has even started.
				if mu.writer == nil {
					started <- mu.name
				}
				// send error message to the result channel
				r <- multiUseResult{err: parser2.AnyToError(rec)}
			}
			close(r)
		}()
		st := funcGen.NewEmptyStack[Value]()
		st.Push(NewListFromIterable(mu.createIterable(started)))
		value, err := mu.fu(st, nil)
		if err != nil {
			mu.writerLock.Lock()
			if mu.writer == nil {
				started <- mu.name
			}
			mu.writerLock.Unlock()
			r <- multiUseResult{err: err}
			return
		}
		// Force evaluation of lists. Lazy evaluation is not possible here because it is
		// not possible to iterate over a list at any time. Iteration is only possible
		// synchronously with all iterators at the same time.
		err = deepEvalLists(st, value)
		if err != nil {
			r <- multiUseResult{err: err}
			return
		}
		r <- multiUseResult{result: value}
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
				log.Print("panic in multiUse producer: ", rec)
				errChan <- parser2.AnyToError(rec)
			}
			// If a panic occurs, the writers are not closed. This ensures that the consumers
			// do not stop waiting for data and therefore do not send a result. This ensures
			// that the error sent above is actually received.
		}()
		running := len(ml)
		_, err := i(func(v Value) bool {
			for _, mu := range ml {
				if mu.writer != nil {
					select {
					case mu.writer <- v:
					case <-mu.requestClose:
						running--
						mu.writerLock.Lock()
						close(mu.writer)
						mu.writer = nil
						mu.writerLock.Unlock()
					}
				}
			}
			return running > 0
		})
		if err != nil {
			errChan <- err
		} else {
			for _, mu := range ml {
				if mu.writer != nil {
					close(mu.writer)
				}
			}
		}
	}()
	return errChan
}

// runConsumerClosures runs all the consumer closures and waits for them to be
// started. Started means that the closure has requested an iterator to iterate
// over the list.
func (ml multiUseList) runConsumerClosures() error {
	started := make(chan string)
	for _, mu := range ml {
		mu.runConsumer(started)
	}

	// wait for all consumers to be started
	for i := 0; i < len(ml); i++ {
		select {
		case <-time.After(startTimeout * time.Second):
			return ml.timeOutError()
		case <-started:
		}
	}
	return nil
}

// createResult creates the result map. If the source iterator panics, the panic
// is rethrown. If one of the closures panics, the panic is recovered and also
// rethrown. This method needs to be called in the main thread. It is the only
// part that does not run in its own goroutine.
func (ml multiUseList) createResult(errChan <-chan error) (Map, error) {
	resultMap := listMap.New[Value](len(ml))
	for _, mu := range ml {
		select {
		case result := <-mu.result:
			if result.err != nil {
				return Map{}, result.err
			} else {
				resultMap = resultMap.Append(mu.name, result.result)
			}
		case err := <-errChan:
			return Map{}, err
		}
	}
	return NewMap(resultMap), nil
}

// timeOutError creates the error message if one of the closures does not request
// its iterator.
func (ml multiUseList) timeOutError() error {
	var buffer bytes.Buffer
	buffer.WriteString("list passed to function is not used; affected function(s): ")
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
