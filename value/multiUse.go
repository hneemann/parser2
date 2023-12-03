package value

import (
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
)

type multiUseEntry struct {
	name string
	fu   funcGen.Func[Value]
}

func (l *List) MultiUse(st funcGen.Stack[Value]) Map {
	if m, ok := st.Get(1).ToMap(); ok {
		var muList []multiUseEntry
		m.Iter(func(key string, value Value) bool {
			if f, ok := value.ToClosure(); ok {
				if f.Args == 1 {
					muList = append(muList, multiUseEntry{name: key, fu: f.Func})
				} else {
					panic("map in multiUse needs to contain closures with one argument")
				}
			} else {
				panic("map in multiUse need to contain closures")
			}
			return true
		})

		resultMap := listMap.New[Value](len(muList))
		for _, mu := range muList {
			st.Push(l)
			r := mu.fu(st.CreateFrame(1), nil)
			resultMap = resultMap.Append(mu.name, r)
		}
		return NewMap(resultMap)
	} else {
		panic("first argument in multiUse needs to be a map")
	}
}
