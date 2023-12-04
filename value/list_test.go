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
		{exp: "list(6).combine((a,b)->a+b)", res: NewList(Int(1), Int(3), Int(5), Int(7), Int(9))},
		{exp: "list(6).combine3((a,b,c)->a+b+c)", res: NewList(Int(3), Int(6), Int(9), Int(12))},
		{exp: "list(6).combineN(3,(a,l)->l[0]+l[1]+l[2])", res: NewList(Int(3), Int(6), Int(9), Int(12))},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "list(20).visit([],(vis,val)->vis.append(val)).string()", res: String("[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]")},
		{exp: visitAndCollect, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: accept, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: "list(12).groupByString(i->\"n\"+round(i/4)).order(a->a.key).string()",
			res: String("[{key:n0, value:[0, 1]}, {key:n1, value:[2, 3, 4, 5]}, {key:n2, value:[6, 7, 8, 9]}, {key:n3, value:[10, 11]}]")},
		{exp: "list(12).groupByInt(i->round(i/4)).order(a->a.key).string()",
			res: String("[{key:0, value:[0, 1]}, {key:1, value:[2, 3, 4, 5]}, {key:2, value:[6, 7, 8, 9]}, {key:3, value:[10, 11]}]")},
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
		{exp: "list(1e9).present(n->n>100)", res: Bool(true)},
		{exp: "list(10).present(n->n>100)", res: Bool(false)},

		{exp: "[1,5,3,2,4].minMax(n->n).string()", res: String("{min:1, max:5, valid:true}")},
		{exp: "[].minMax(n->n).string()", res: String("{min:0, max:0, valid:false}")},
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
			got := NewListConvert(tt.conv, tt.items...).ToSlice()
			assert.Equalf(t, tt.want, got, "NewListConvert, %v vs. %v", tt.want, got)
		})
	}
}
