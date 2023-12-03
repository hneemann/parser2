package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
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

func (mu *multiUseEntry) runConsumer(started chan struct{}) {
	r := make(chan multiUseResult)
	mu.result = r
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				if e, ok := rec.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", rec)
				}
				r <- multiUseResult{result: nil, err: err}
			}
			close(r)
		}()
		st := funcGen.NewEmptyStack[Value]()
		st.Push(NewListFromIterable(mu.createIterable(started)))
		value := mu.fu(st, nil)
		if list, ok := value.(*List); ok {
			// force evaluation of lists
			list.Eval()
		}
		r <- multiUseResult{result: value, err: nil}
	}()
}

func (ml multiUseList) runProducer(i iterator.Iterator[Value]) <-chan error {
	errChan := make(chan error)
	go func() {
		defer func() {
			for _, mu := range ml {
				if mu.writer != nil {
					close(mu.writer)
				}
			}
			if rec := recover(); rec != nil {
				if e, ok := rec.(error); ok {
					errChan <- e
				} else {
					errChan <- fmt.Errorf("%v", rec)
				}
			}
		}()
		i(func(v Value) bool {
			for _, mu := range ml {
				if mu.writer != nil {
					select {
					case mu.writer <- v:
					case <-mu.requestClose:
						close(mu.writer)
						mu.writer = nil
					}
				}
			}
			return true
		})
	}()
	return errChan
}

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
