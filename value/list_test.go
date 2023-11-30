package value

import (
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
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
		// Prefix Sum
		{exp: "[1,2,3,4,4].iir(i->i,(i,l)->i+l)", res: NewList(Int(1), Int(3), Int(6), Int(10), Int(14))},
		// Fibonacci Sequence
		{exp: "list(12).iir(i->[1,1],(i,l)->[l[1],l[0]+l[1]]).map(l->l[0])",
			res: NewList(Int(1), Int(1), Int(2), Int(3), Int(5), Int(8), Int(13), Int(21), Int(34), Int(55), Int(89), Int(144))},
		// Low-pass Filter
		{exp: "list(11).iir(i->0,(i,l)->(1024+l)>>1)",
			res: NewList(Int(0), Int(512), Int(768), Int(896), Int(960), Int(992), Int(1008), Int(1016), Int(1020), Int(1022), Int(1023))},
		{exp: "list(6).combine((a,b)->a+b)", res: NewList(Int(1), Int(3), Int(5), Int(7), Int(9))},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "\"\"+list(20).visit([],(vis,val)->vis.append(val))", res: String("[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]")},
		{exp: visitAndCollect, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: accept, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
		{exp: "\"\"+list(12).groupByString(i->\"n\"+round(i/4)).order((a,b)->a.key<b.key)",
			res: String("[{key:n0, value:[0, 1]}, {key:n1, value:[2, 3, 4, 5]}, {key:n2, value:[6, 7, 8, 9]}, {key:n3, value:[10, 11]}]")},
		{exp: "\"\"+list(12).groupByInt(i->round(i/4)).order((a,b)->a.key<b.key)",
			res: String("[{key:0, value:[0, 1]}, {key:1, value:[2, 3, 4, 5]}, {key:2, value:[6, 7, 8, 9]}, {key:3, value:[10, 11]}]")},
		{exp: "string(list(3).map(i->(i+1)*10).number((n,e)->\"\"+n+\"->\"+e))", res: String("[0->10, 1->20, 2->30]")},
		{exp: "[].reverse().string()", res: String("[]")},
		{exp: "[1].reverse().string()", res: String("[1]")},
		{exp: "[1,2].reverse().string()", res: String("[2, 1]")},
		{exp: "[1,2,3].reverse().string()", res: String("[3, 2, 1]")},
		{exp: "[1,2,3,4].reverse().string()", res: String("[4, 3, 2, 1]")},
	})
}

const visitAndCollect = `
  let data=list(100).map(i->if i%10=9 then i else 0);
  
  let events=data
       .visit([],(vis,i)->if i!=0 
                          then vis.append(i)
                          else vis);
  ""+events
`

const accept = `
  let data=list(100).map(i->if i%10=9 then i else 0);
  
  let events=data.accept(i->i!=0);

  ""+events
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

func TestList_Top(t *testing.T) {
	tests := []struct {
		name    string
		origLen int
		topLen  int
		wantLen int
	}{
		{name: "normal", origLen: 100, topLen: 4, wantLen: 4},
		{name: "match", origLen: 4, topLen: 4, wantLen: 4},
		{name: "short", origLen: 4, topLen: 10, wantLen: 4},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			g := NewListFromIterable(iterator.Generate(test.origLen, func(i int) Value { return Int(i) }))
			l := g.Top(funcGen.NewStack[Value](g, Int(test.topLen)))
			sl := l.ToSlice()
			assert.Equal(t, test.wantLen, len(sl))
			for i, item := range sl {
				li, ok := item.ToInt()
				assert.True(t, ok)
				assert.Equal(t, i, li)
			}
		})
	}
}
