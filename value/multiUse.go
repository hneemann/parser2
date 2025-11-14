package value

import (
	"errors"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
)

// MultiUse takes a map of functions, and the list is passed to the functions. The
// return values of the functions are returned in a map. The keys in the result
// map are the same keys used to pass the functions. MultiUse is useful if you
// have to use the same list multiple times and the list is both expensive to
// create and expensive to store. This is because MultiUse allows you to use the
// list multiple times without having to store the list elements for later reuse.
// An example could be reading a csv file. Reading such a file is expensive, so
// you want to do it only once. But if the file is a large file, you also do not
// want to store the content of the file in memory to be able to use it multiple
// times. MultiUse allows you to read the file once and then use the list
// multiple times without having to store the content of the file.
func (l *List) MultiUse(st funcGen.Stack[Value]) (Map, error) {
	if m, ok := st.Get(1).ToMap(); ok {
		var muList multiUseList
		for key, value := range m.Iter {
			if f, ok := value.(Closure); ok {
				if f.Args == 1 {
					muList = append(muList, &multiUseEntry{name: key, fu: f.Func})
				} else {
					return Map{}, errors.New("map in multiUse needs to contain functions with one argument")
				}
			} else {
				return Map{}, errors.New("map in multiUse need to contain functions")
			}
		}
		if len(muList) < 1 {
			return Map{}, errors.New("map in multiUse needs to contain at leat two functions")
		}

		prList, run, done := iterator.CopyProducer[Value](len(muList))
		for i, mu := range muList {
			pr := prList[i]
			go mu.runConsumer(pr, done)
		}
		err := run(l.iterable(st))

		if err != nil {
			return EmptyMap, err
		}

		return muList.createResult()
	} else {
		return EmptyMap, errors.New("first argument in multiUse needs to be a map")
	}
}

type multiUseEntry struct {
	name   string
	fu     funcGen.ParserFunc[Value]
	result Value
}

type multiUseList []*multiUseEntry

// runConsumer calls the closure and sends the result to the result channel. If
// the closure panics, the panic is recovered and also sent to the result
// channel. if the closure returns a list, the list is evaluated before it is
// sent to the result channel.
func (mu *multiUseEntry) runConsumer(itera iterator.Producer[Value], done func(error)) {
	st := funcGen.NewEmptyStack[Value]()
	used := false
	var innerErr error
	st.Push(NewListFromIterable(func(st funcGen.Stack[Value]) iterator.Producer[Value] {
		if used {
			innerErr = errors.New("copied iterator a can only be used once")
			return iterator.Empty[Value]()
		}
		used = true
		return itera
	}))
	value, err := mu.fu(st, nil)
	if innerErr != nil {
		done(innerErr)
		return
	}
	if err != nil {
		done(err)
		return
	}

	// Force evaluation of lists. Lazy evaluation is not possible here because it is
	// not possible to iterate over a list at any time. Iteration is only possible
	// synchronously with all iterators at the same time.
	err = deepEvalLists(st, value)
	mu.result = value
	done(err)
}

// createResult creates the result map. If the source iterator panics, the panic
// is rethrown. If one of the closures panics, the panic is recovered and also
// rethrown. This method needs to be called in the main thread. It is the only
// part that does not run in its own goroutine.
func (ml multiUseList) createResult() (Map, error) {
	resultMap := listMap.New[Value](len(ml))
	for _, mu := range ml {
		if mu.result == nil {
			return Map{}, errors.New("internal multiuse error: nil result")
		}
		resultMap = resultMap.Append(mu.name, mu.result)
	}
	return NewMap(resultMap), nil
}
