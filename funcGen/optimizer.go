package funcGen

import (
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/listMap"
)

type optimizer[V any] struct {
	st Stack[V]
	g  *FunctionGenerator[V]
}

func NewOptimizer[V any](st Stack[V], g *FunctionGenerator[V]) parser2.Optimizer {
	return optimizer[V]{st: st, g: g}
}

func (o optimizer[V]) Optimize(ast parser2.AST) parser2.AST {
	// evaluate const operations like 1+2
	if oper, ok := ast.(*parser2.Operate); ok {
		if operator, ok := o.g.opMap[oper.Operator]; ok {
			if bc, ok := o.isConst(oper.B); ok {
				if operator.IsPure {
					if ac, ok := o.isConst(oper.A); ok {
						co, err := operator.Impl.Calc(o.st, ac, bc)
						if err != nil {
							return ast
						}
						return &parser2.Const[V]{co, oper.Line}
					}
				}
				if operator.IsCommutative {
					if aOp, ok := oper.A.(*parser2.Operate); ok && aOp.Operator == oper.Operator {
						if iac, ok := o.isConst(aOp.A); ok {
							co, err := operator.Impl.Calc(o.st, iac, bc)
							if err != nil {
								return ast
							}
							return &parser2.Operate{
								Operator: oper.Operator,
								Priority: oper.Priority,
								A:        &parser2.Const[V]{co, oper.Line},
								B:        aOp.B,
							}
						}
						if ibc, ok := o.isConst(aOp.B); ok {
							co, err := operator.Impl.Calc(o.st, ibc, bc)
							if err != nil {
								return ast
							}
							return &parser2.Operate{
								Operator: oper.Operator,
								Priority: oper.Priority,
								A:        aOp.A,
								B:        &parser2.Const[V]{co, oper.Line},
							}
						}
					}
				}
			}
		}
	}

	// evaluate const unary operations like -1
	if oper, ok := ast.(*parser2.Unary); ok {
		if operator, ok := o.g.uMap[oper.Operator]; ok {
			if c, ok := o.isConst(oper.Value); ok {
				co, err := operator.Impl.Calc(c)
				if err != nil {
					return ast
				}
				return &parser2.Const[V]{co, oper.Line}
			}
		}
	}
	// evaluate const if operation
	if ifNode, ok := ast.(*parser2.If); ok && o.g.toBool != nil {
		if c, ok := o.isConst(ifNode.Cond); ok {
			if cond, ok := o.g.toBool(c); ok {
				if cond {
					return ifNode.Then
				} else {
					return ifNode.Else
				}
			} else {
				return ast
			}
		}
	}
	// evaluate const list literals like [1,2,3]
	if o.g.listHandler != nil {
		if list, ok := ast.(*parser2.ListLiteral); ok {
			if l, ok := o.allConst(list.List); ok {
				return &parser2.Const[V]{o.g.listHandler.FromList(l), list.Line}
			}
		}
	}
	// evaluate const map literals like {a:1,b:2}
	if o.g.mapHandler != nil {
		if m, ok := ast.(*parser2.MapLiteral); ok {
			cm := listMap.New[V](len(m.Map))
			for key, value := range m.Map.Iter {
				if v, ok := o.isConst(value); ok {
					cm = cm.Append(key, v)
				} else {
					cm = nil
					break
				}
			}
			if cm != nil {
				return &parser2.Const[V]{o.g.mapHandler.FromMap(cm), m.Line}
			}
		}
	}
	// evaluate const static function calls like sqrt(2)
	if fc, ok := ast.(*parser2.FunctionCall); ok {
		if ident, ok := fc.Func.(*parser2.Ident); ok {
			if fu, ok := o.g.staticFunctions[ident.Name]; ok && fu.IsPure {
				if fu.argsNumberNotMatching(len(fc.Args)) {
					return ast
				}
				if c, ok := o.allConst(fc.Args); ok {
					v, err := fu.Func(NewStack[V](c...), nil)
					if err != nil {
						return ast
					}
					return &parser2.Const[V]{v, ident.Line}
				}
			}
		}
		if con, ok := fc.Func.(*parser2.Const[V]); ok {
			if o.g.closureHandler != nil {
				if closure, ok := o.g.closureHandler.ToClosure(con.Value); ok {
					if closure.IsPure {
						if closure.Args != -1 && closure.Args != len(fc.Args) {
							return ast
						}
						if c, ok := o.allConst(fc.Args); ok {
							v, err := closure.Func(NewStack[V](c...), nil)
							if err != nil {
								return ast
							}
							return &parser2.Const[V]{v, con.Line}
						}
					}
				}
			}
		}
	}

	// evaluate const method calls like c.conj()
	if mc, ok := ast.(*parser2.MethodCall); ok {
		if con, ok := mc.Value.(*parser2.Const[V]); ok {
			if c, ok := o.allConst(mc.Args); ok {
				if o.g.methodHandler != nil {
					fu, err := o.g.methodHandler.GetMethod(con.Value, mc.Name)
					if err != nil {
						return ast
					}
					if fu.IsPure {
						if fu.Args != -1 && len(c)+1 != fu.Args {
							return ast
						}
						args := make([]V, len(c)+1)
						args[0] = con.Value
						copy(args[1:], c)
						v, err := fu.Func(NewStack[V](args...), nil)
						if err != nil {
							return ast
						}
						return &parser2.Const[V]{v, con.Line}
					}
				}
			}
		}
	}

	if o.g.closureHandler != nil {
		if cl, ok := ast.(*parser2.ClosureLiteral); ok && len(cl.OuterIdents) == 0 && !cl.Recursive {
			closureFunc, err := o.g.GenerateFunc(cl.Func, GeneratorContext{am: cl.Names})
			if err != nil {
				return ast
			}
			v := o.g.closureHandler.FromClosure(Function[V]{
				Func: closureFunc,
				Args: len(cl.Names),
			})
			return &parser2.Const[V]{v, cl.Line}
		}
	}

	return ast
}

func (o optimizer[V]) allConst(asts []parser2.AST) ([]V, bool) {
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

func (o optimizer[V]) isConst(ast parser2.AST) (V, bool) {
	if n, ok := ast.(*parser2.Const[V]); ok {
		return n.Value, true
	}
	var zero V
	return zero, false
}
