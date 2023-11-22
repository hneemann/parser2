package value

import "testing"

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
		{exp: "7 ~ [1,2,3]", res: Bool(false)},
		{exp: "[1,2,3].size()", res: Int(3)},
		{exp: "[1,2,3]=[1,2,3]", res: Bool(true)},
		{exp: "[1,2,3]=[1,2,4]", res: Bool(false)},
		{exp: "[1,2,3]=[1,2]", res: Bool(false)},
		{exp: "[1,2,3].map(e->e*2)", res: NewList(Int(2), Int(4), Int(6))},
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
		{exp: "\"\"+list(12).group(i->\"n\"+round(i/4),i->i).list().order((a,b)->a.key<b.key)",
			res: String("[{key:n0, value:[0, 1]}, {key:n1, value:[2, 3, 4, 5]}, {key:n2, value:[6, 7, 8, 9]}, {key:n3, value:[10, 11]}]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "let a=[1,2].append(3);\"\"+[a.append(4), a.append(5)]", res: String("[[1, 2, 3, 4], [1, 2, 3, 5]]")},
		{exp: "\"\"+list(20).visit(val->[val],(vis,val)->vis.append(val))", res: String("[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]")},
		{exp: stateMachine, res: String("[9, 19, 29, 39, 49, 59, 69, 79, 89, 99]")},
	})
}

const stateMachine = `
  let data=list(100).map(i->if i%10=9 then i else 0);
  
  let events=data
       .visit(
            i->{last:i, found:[]},
            (vis,i)->if vis.last<i 
                     then {last:i, found:vis.found.append(i)}
                     else {last:i, found:vis.found}
             )
       .found;

  ""+events
`
