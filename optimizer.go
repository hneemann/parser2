package parser2

import "fmt"

type optimizer[V any] struct {
	g *FunctionGenerator[V]
}

func NewOptimizer[V any](g *FunctionGenerator[V]) Optimizer {
	return optimizer[V]{g}
}

func (o optimizer[V]) Optimize(ast AST) AST {
	// evaluate constants
	if i, ok := ast.(Ident); ok {
		if c, ok := o.g.constants[string(i)]; ok {
			return Const[V]{c}
		}
	}
	// evaluate const operations like 1+2
	if oper, ok := ast.(*Operate); ok {
		if operator, ok := o.g.opMap[oper.Operator]; ok {
			if ac, ok := o.isConst(oper.A); ok {
				if bc, ok := o.isConst(oper.B); ok {
					return Const[V]{operator.Impl(ac, bc)}
				}
			}
		}
	}
	// evaluate const list literals like [1,2,3]
	if o.g.listHandler != nil {
		if list, ok := ast.(ListLiteral); ok {
			if l, ok := o.allConst(list); ok {
				return Const[V]{o.g.listHandler.FromList(l)}
			}
		}
	}
	// evaluate const map literals like {a:1,b:2}
	if o.g.mapHandler != nil {
		if m, ok := ast.(MapLiteral); ok {
			cm := map[string]V{}
			for k, e := range m {
				if v, ok := o.isConst(e); ok {
					cm[k] = v
				} else {
					cm = nil
					break
				}
			}
			if cm != nil {
				return Const[V]{o.g.mapHandler.FromMap(cm)}
			}
		}
	}
	// evaluate const static function calls like sqrt(2)
	if fc, ok := ast.(*FunctionCall); ok {
		if name, ok := fc.Func.(Ident); ok {
			if fu, ok := o.g.staticFuncs[string(name)]; ok && fu.IsPure {
				if fu.Args != len(fc.Args) {
					panic(fmt.Sprintf("number of args wrong in: %v", fc))
				}
				if c, ok := o.allConst(fc.Args); ok {
					return Const[V]{fu.Func(c)}
				}
			}
		}
	}
	return nil
}

func (o optimizer[V]) allConst(asts []AST) ([]V, bool) {
	con := make([]V, len(asts))
	for i, ast := range asts {
		if c, ok := o.isConst(ast); ok {
			con[i] = c
		} else {
			return nil, false
		}
	}
	return con, true
}

func (o optimizer[V]) isConst(ast AST) (V, bool) {
	if n, ok := ast.(Const[V]); ok {
		return n.Value, true
	}
	var zero V
	return zero, false
}
