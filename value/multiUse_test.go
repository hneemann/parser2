package value

import (
	"testing"
)

func TestMultiUse(t *testing.T) {
	runTest(t, []testType{
		{exp: `[1,2,3,4].multiUse({
                  add:  l->l.map(e->e+1),
                  mul:  l->l.map(e->e*2)
                 }).string()`, res: String("{add:[2, 3, 4, 5], mul:[2, 4, 6, 8]}")},
		{exp: `[1,2,3,4].multiUse({
		          sum:   l->l.reduce((a,b)->a+b),
		          prod:  l->l.reduce((a,b)->a*b),
		          first: l->l.first(),
		       }).string()`, res: String("{sum:10, prod:24, first:1}")},
		{exp: `numbers(1000000000).multiUse({
		           a:  l->l.first(),
		           b:  l->l.first()*2,
		        }).string()`, res: String("{a:0, b:0}")},
		{exp: `numbers(1000000000).map(n->n+10).multiUse({
		           a:  l->l.first(),
		           b:  l->l.first()*2,
		        }).string()`, res: String("{a:10, b:20}")},
		{exp: `numbers(1000000000).multiUse({
		           a:  l->l.present(n->n>10),
		           b:  l->l.present(n->n>100),
		        }).string()`, res: String("{a:true, b:true}")},
	})
}
