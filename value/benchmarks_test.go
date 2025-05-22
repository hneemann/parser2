package value

import (
	"github.com/hneemann/parser2/funcGen"
	"testing"
)

const bench1 = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

data.iirApply(createLowPass("f",p->p.t,p->p.s,1/(2*pi)))
`

const bench2 = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

func CLP(name, getT, getX, tau)
   {initial:
      p->p.put(name,getX(p)),
    filter: 
 	  (p0,p1,y)->
		let a = exp((getT(p0) - getT(p1)) / tau);
		p1.put(name, y.get(name)*a + getX(p1)*(1-a))};

data.iirApply(CLP("f",p->p.t,p->p.s,1/(2*pi)))
`

func getList(bench string) *List {
	valueParser := New(nil)
	f, err := valueParser.Generate(bench)
	if err != nil {
		panic(err)
	}
	st := funcGen.NewEmptyStack[Value]()
	l, err := f(st)
	if err != nil {
		panic(err)
	}
	list, ok := l.ToList()
	if !ok {
		panic("not a list")
	}
	return list
}

var list1 = getList(bench1)

var list2 = getList(bench2)

func Benchmark_filter1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list1.Iterate(funcGen.NewEmptyStack[Value](), func(v Value) error {
			return nil
		})
	}
}

func Benchmark_filter2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list2.Iterate(funcGen.NewEmptyStack[Value](), func(v Value) error {
			return nil
		})
	}
}
