package value

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList(t *testing.T) {
	runTest(t, []testType{
		{exp: "[1,2,3]", res: NewList(Int(1), Int(2), Int(3))},
		{exp: "let a=2; [1,a,3]", res: NewList(Int(1), Int(2), Int(3))},
		{exp: "let a=2;[1,a]+[3,4]", res: NewList(Int(1), Int(2), Int(3), Int(4))},
		{exp: "[1,2,3][2]", res: Int(3)},
		{exp: "[1,2].replace(l->l[0]+l[1])", res: Int(3)},
		{exp: "[1,2,3].indexOf(2)", res: Int(1)},
		{exp: "[1,2,3].indexOf(7)", res: Int(-1)},
		{exp: "2 ~ [1,2,3]", res: Bool(true)},
		{exp: "[2,3] ~ [1,2,3]", res: Bool(true)},
		{exp: "[1,2] ~ [1,2,3]", res: Bool(true)},
		{exp: "[1,3] ~ [1,2,3]", res: Bool(true)},
		{exp: "[1,2,3] ~ [1,2,3]", res: Bool(true)},
		{exp: "[1,2,3,4] ~ [1,2,3]", res: Bool(false)},
		{exp: "7 ~ [1,2,3]", res: Bool(false)},
		{exp: "[2,7] ~ [1,2,3]", res: Bool(false)},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "[1,2,3]=[1,2,3]", res: Bool(true)},
		{exp: "[1,2,3]=[1,2,4]", res: Bool(false)},
		{exp: "[1,2,3]=[1,2]", res: Bool(false)},
		{exp: "[1,2,3].map(e->e*2)", res: NewList(Int(2), Int(4), Int(6))},
		{exp: "let a=2;[1,2,3].map(e->e*a)", res: NewList(Int(2), Int(4), Int(6))},
		{exp: "[1,2,3,4,5].reduce((a,b)->a+b)", res: Int(15)},
		{exp: "[1,2,3,4,5].sum()", res: Int(15)},
		{exp: "[1,2,3,4,5].mapReduce([-1,0], (s,i)->s.append(i)).string()", res: String("[-1, 0, 1, 2, 3, 4, 5]")},
		{exp: "[1,2,3].map(i->i*i)", res: NewList(Int(1), Int(4), Int(9))},
		{exp: "[1,2,3].accept(i->i>1)", res: NewList(Int(2), Int(3))},
		{exp: "[1,2,3].accept(i->i>1)", res: NewList(Int(2), Int(3))},
		{exp: "[1,2,3,3].reduce((a,b)->a+b)", res: Int(9)},
		{exp: "[1,3,2,4].orderLess((a,b)->a<b)", res: NewList(Int(1), Int(2), Int(3), Int(4))},
		{exp: "[1,3,2,4].orderLess((a,b)->a>b)", res: NewList(Int(4), Int(3), Int(2), Int(1))},
		{exp: "[1,3,2,4].order(n->n)", res: NewList(Int(1), Int(2), Int(3), Int(4))},
		{exp: "[1,3,2,4].orderRev(n->n)", res: NewList(Int(4), Int(3), Int(2), Int(1))},
		// Prefix Sum
		{exp: "[1,2,3,4,4].iir(i->i,(i,l)->i+l)", res: NewList(Int(1), Int(3), Int(6), Int(10), Int(14))},
		// Fibonacci Sequence
		{exp: "list(12).iir(i->[1,1],(i,l)->[l[1],l[0]+l[1]]).map(l->l[0])",
			res: NewList(Int(1), Int(1), Int(2), Int(3), Int(5), Int(8), Int(13), Int(21), Int(34), Int(55), Int(89), Int(144))},
		// Low-pass Filter
		{exp: "list(11).iir(i->0,(i,l)->(1024+l)>>1)",
			res: NewList(Int(0), Int(512), Int(768), Int(896), Int(960), Int(992), Int(1008), Int(1016), Int(1020), Int(1022), Int(1023))},
		// non-equidistant Low-pass Filter
		{exp: realIir, res: Float(0.707192)},
		{exp: realIirBuiltin, res: Float(0.707192)},
		{exp: realIirApply, res: Float(0.707192)},
		{exp: manualIirApply, res: Float(0.707192)},
		{exp: "list(6).combine((a,b)->a+b)", res: NewList(Int(1), Int(3), Int(5), Int(7), Int(9))},
		{exp: "list(6).combine3((a,b,c)->a+b+c)", res: NewList(Int(3), Int(6), Int(9), Int(12))},
		{exp: "list(6).combineN(3,l->l[0]+l[1]+l[2])", res: NewList(Int(3), Int(6), Int(9), Int(12))},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "list(20).visit([],(vis,val)->vis.append(val)).string()", res: String("[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]")},
		{exp: fsm, res: String("[{start:15, end:20}, {start:35, end:40}, {start:55, end:60}, {start:75, end:80}]")},
		{exp: visitAndCollect, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: accept, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: "list(12).groupByString(i->\"n\"+round(i/4)).order(a->a.key).string()",
			res: String("[{key:n0, values:[0, 1]}, {key:n1, values:[2, 3, 4, 5]}, {key:n2, values:[6, 7, 8, 9]}, {key:n3, values:[10, 11]}]")},
		{exp: "list(12).groupByInt(i->round(i/4)).order(a->a.key).string()",
			res: String("[{key:0, values:[0, 1]}, {key:1, values:[2, 3, 4, 5]}, {key:2, values:[6, 7, 8, 9]}, {key:3, values:[10, 11]}]")},
		{exp: "list(12).uniqueString(i->\"n\"+round(i/4)).order(a->a).string()", res: String("[n0, n1, n2, n3]")},
		{exp: "list(12).uniqueInt(i->round(i/4)).order(a->a).string()", res: String("[0, 1, 2, 3]")},
		{exp: "list(12).map(i->round(i/4)).compact(n->n).string()", res: String("[0, 1, 2, 3]")},
		{exp: "string(list(3).map(i->(i+1)*10).number((n,e)->\"\"+n+\"->\"+e))", res: String("[0->10, 1->20, 2->30]")},
		{exp: "[].reverse().string()", res: String("[]")},
		{exp: "[1].reverse().string()", res: String("[1]")},
		{exp: "[1,2].reverse().string()", res: String("[2, 1]")},
		{exp: "[1,2,3].reverse().string()", res: String("[3, 2, 1]")},
		{exp: "[1,2,3,4].reverse().string()", res: String("[4, 3, 2, 1]")},

		{exp: "[1,2,3].top(0).string()", res: String("[]")},
		{exp: "[1,2,3].top(1).string()", res: String("[1]")},
		{exp: "[1,2,3].top(2).string()", res: String("[1, 2]")},
		{exp: "[1,2,3].top(3).string()", res: String("[1, 2, 3]")},
		{exp: "[1,2,3].top(4).string()", res: String("[1, 2, 3]")},

		{exp: "[1,2,3].skip(0).string()", res: String("[1, 2, 3]")},
		{exp: "[1,2,3].skip(1).string()", res: String("[2, 3]")},
		{exp: "[1,2,3].skip(2).string()", res: String("[3]")},
		{exp: "[1,2,3].skip(3).string()", res: String("[]")},
		{exp: "[1,2,3].skip(4).string()", res: String("[]")},

		{exp: "[1,2,3,4].first()", res: Int(1)},
		{exp: "list(100).first()", res: Int(0)},
		{exp: "[1,2,3,4].last()", res: Int(4)},
		{exp: "list(10).last()", res: Int(9)},
		{exp: "list(1e9).present(n->n>100)", res: Bool(true)},
		{exp: "list(10).present(n->n>100)", res: Bool(false)},

		{exp: "[1,5,3,2,4].minMax(n->n).string()", res: String("{min:1, max:5, valid:true}")},
		{exp: "[].minMax(n->n).string()", res: String("{min:0, max:0, valid:false}")},

		{exp: "[1,2,3].cross([10,20,30],(a,b)->a+b).string()", res: String("[11, 21, 31, 12, 22, 32, 13, 23, 33]")},

		{exp: `let a=[{n:1, s:"eins"},{n:3, s:"drei"},{n:5, s:"fünf"}];
               let b=[{n:2, s:"zwei"},{n:4, s:"vier"},{n:6, s:"sechs"}];
               a.merge(b,(a,b)->a.n<b.n).string()`,
			res: String("[{n:1, s:eins}, {n:2, s:zwei}, {n:3, s:drei}, {n:4, s:vier}, {n:5, s:fünf}, {n:6, s:sechs}]")},

		{exp: movingWindow,
			res: String("[{t0:0, t1:0, len:1}, {t0:0, t1:0.1, len:2}, {t0:0, t1:0.2, len:3}, {t0:0, t1:0.3, len:4}, {t0:0, t1:0.4, len:5}, {t0:0, t1:0.5, len:6}, {t0:0, t1:0.6, len:7}, {t0:0, t1:0.7, len:8}, {t0:0, t1:0.8, len:9}, {t0:0, t1:0.9, len:10}, {t0:0, t1:1, len:11}, {t0:0.1, t1:1.1, len:11}, {t0:0.2, t1:1.2, len:11}, {t0:0.3, t1:1.3, len:11}, {t0:0.4, t1:1.4, len:11}, {t0:0.5, t1:1.5, len:11}, {t0:0.6, t1:1.6, len:11}, {t0:0.7, t1:1.7, len:11}, {t0:0.8, t1:1.8, len:11}, {t0:0.9, t1:1.9, len:11}]")},
	})
}

const realIir = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

func lowPass(tau)
	(p0,p1,y)->
		let dt = p1.t - p0.t;
		let a = exp(-dt / tau);
		{t:p1.t,f:y.f*a + p1.s*(1-a)};

let filtered=data.iirCombine(p->{t:p.t,f:0},lowPass(1/(2*pi)));

let minMax=filtered.skip(100).minMax(p->p.f);
(minMax.max-minMax.min)/2
`

const realIirBuiltin = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

let lp=createLowPass("f",p->p.t,p->p.s,1/(2*pi));

let filtered=data.iirCombine(lp.initial,lp.filter);

let minMax=filtered.skip(100).minMax(p->p.f);
(minMax.max-minMax.min)/2
`

const realIirApply = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

let filtered=data.iirApply(createLowPass("f",p->p.t,p->p.s,1/(2*pi)));

let minMax=filtered.skip(100).minMax(p->p.f);
(minMax.max-minMax.min)/2
`

const manualIirApply = `
let data=list(1000).map(i->
	let t=i/50;
	{t:t, s:sin(2*pi*t)});

func CLP(name, getT, getX, tau)
   {initial:
      p->p.put(name,getX(p)),
    filter: 
 	  (p0,p1,y)->
		let dt = getT(p1) - getT(p0);
		let a = exp(-dt / tau);
		p1.put(name, y.get(name)*a + getX(p1)*(1-a))};

let filtered=data.iirApply(CLP("f",p->p.t,p->p.s,1/(2*pi)));

let minMax=filtered.skip(100).minMax(p->p.f);
(minMax.max-minMax.min)/2
`

const visitAndCollect = `
  let data=list(100).map(i->if i%10=9 then i else 0);
  
  let events=data
       .visit([],(vis,i)->if i!=0 
                          then vis.append(i)
                          else vis);
  events.string()
`

const accept = `
  let data=list(100).map(i->if i%10=9 then i else 0);
  
  let events=data.accept(i->i!=0);

  events.string()
`

const fsm = `
	let data=list(100).map(i->{t:i,v:i%20});

    const search=0;
    const inEvent=1;
	func fsm(vis, p)
		if vis.state=search
        then
			if p.v<15 then vis
			else {state:inEvent, start:p.t, events:vis.events}
		else
			if p.v>=15 then vis
			else {state:search, events:vis.events.append({start:vis.start, end:p.t})};

  let events=data.visit({state:search, events:[]},fsm).events;

  events.string()
`

const movingWindow = `
	let data=list(20).map(i->{t:i/10,v:i});
	
	data.movingWindow(p->p.t).map(l->{t0:l[0].t,t1:l[l.size()-1].t,len:l.size()}).string()
`

func TestNewListCreate(t *testing.T) {
	type testCase[I any] struct {
		name  string
		conv  func(I) Value
		items []I
		want  []Value
	}
	tests := []testCase[int]{
		{
			name: "empty",
			conv: func(i int) Value {
				return Float(i)
			},
			items: []int{},
			want:  nil,
		},
		{
			name: "some",
			conv: func(i int) Value {
				return Float(i)
			},
			items: []int{1, 2, 3, 4},
			want:  []Value{Float(1), Float(2), Float(3), Float(4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewListConvert(tt.conv, tt.items...).ToSlice()
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "NewListConvert, %v vs. %v", tt.want, got)
		})
	}
}
