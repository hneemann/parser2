package value

import (
	"errors"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/iterator"
	"github.com/hneemann/parser2/listMap"
)

// MultiUse takes a map of closures and the list is passed to the closures. The
// return values of the closures are returned in a map. The keys in the result
// map are the same keys used to pass the closures. MultiUse is useful if you
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
		var innerErr error
		m.Iter(func(key string, value Value) bool {
			if f, ok := value.(Closure); ok {
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
			return EmptyMap, innerErr
		}

		prList, run, done := iterator.CopyProducer(l.iterable, len(muList))
		for i, mu := range muList {
			pr := prList[i]
			mu := mu
			go func() {
				done(mu.runConsumer(pr))
			}()
		}
		err := run(st)

		if err != nil {
			return EmptyMap, err
		}

		return muList.createResult(), nil
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
func (mu *multiUseEntry) runConsumer(itera iterator.Producer[Value, funcGen.Stack[Value]]) error {
	st := funcGen.NewEmptyStack[Value]()
	st.Push(NewListFromIterable(itera))
	value, err := mu.fu(st, nil)
	if err != nil {
		return err
	}

	// Force evaluation of lists. Lazy evaluation is not possible here because it is
	// not possible to iterate over a list at any time. Iteration is only possible
	// synchronously with all iterators at the same time.
	err = deepEvalLists(st, value)
	if err != nil {
		return err
	}

	mu.result = value
	return nil
}

// createResult creates the result map. If the source iterator panics, the panic
// is rethrown. If one of the closures panics, the panic is recovered and also
// rethrown. This method needs to be called in the main thread. It is the only
// part that does not run in its own goroutine.
func (ml multiUseList) createResult() Map {
	resultMap := listMap.New[Value](len(ml))
	for _, mu := range ml {
		resultMap = resultMap.Append(mu.name, mu.result)
	}
	return NewMap(resultMap)
}
