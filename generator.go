package parser2

import (
	"fmt"
)

type Operator[V any] struct {
	Operator string
	Impl     func(a, b V) V
}

type ClosureHandler[V any] interface {
	FromClosure(c Closure[V]) V
	ToClosure(fu V) (Closure[V], bool)
}

type ListHandler[V any] interface {
	FromList(items []V) V
	AccessList(list V, index V) (V, error)
}

type MapHandler[V any] interface {
	FromMap(items map[string]V) V
	AccessMap(m V, key string) (V, error)
}

type FunctionGenerator[V any] struct {
	parser          *Parser[V]
	operators       []Operator[V]
	numberParser    NumberParser[V]
	stringHandler   StringHandler[V]
	closureHandler  ClosureHandler[V]
	listHandler     ListHandler[V]
	mapHandler      MapHandler[V]
	optimizer       Optimizer
	toFunction      func(f V) (Function[V], bool)
	staticFuncs     map[string]Function[V]
	constants       map[string]V
	opMap           map[string]Operator[V]
	customGenerator Generator[V]
}

func New[V any]() *FunctionGenerator[V] {
	g := &FunctionGenerator[V]{
		staticFuncs: map[string]Function[V]{},
		constants:   map[string]V{},
	}
	g.optimizer = NewOptimizer(g)
	return g
}

func (g *FunctionGenerator[V]) Parser() *Parser[V] {
	if g.parser == nil {
		if g.numberParser == nil {
			panic("no number parser set")
		}

		parser := NewParser[V]().
			SetNumberParser(g.numberParser).
			SetStringHandler(g.stringHandler)

		opMap := map[string]Operator[V]{}
		for _, o := range g.operators {
			parser.Op(o.Operator)
			opMap[o.Operator] = o
		}
		g.parser = parser
		g.opMap = opMap
	}
	return g.parser
}

func (g *FunctionGenerator[V]) SetNumberParser(numberParser NumberParser[V]) *FunctionGenerator[V] {
	g.numberParser = numberParser
	return g
}

func (g *FunctionGenerator[V]) SetClosureHandler(closureHandler ClosureHandler[V]) *FunctionGenerator[V] {
	g.closureHandler = closureHandler
	return g
}

func (g *FunctionGenerator[V]) SetListHandler(listHandler ListHandler[V]) *FunctionGenerator[V] {
	g.listHandler = listHandler
	return g
}

func (g *FunctionGenerator[V]) SetMapHandler(mapHandler MapHandler[V]) *FunctionGenerator[V] {
	g.mapHandler = mapHandler
	return g
}

func (g *FunctionGenerator[V]) SetStringHandler(stringHandler StringHandler[V]) *FunctionGenerator[V] {
	g.stringHandler = stringHandler
	return g
}

func (g *FunctionGenerator[V]) SetOptimizer(optimizer Optimizer) *FunctionGenerator[V] {
	g.optimizer = optimizer
	return g
}

func (g *FunctionGenerator[V]) SetCustomGenerator(generator Generator[V]) *FunctionGenerator[V] {
	g.customGenerator = generator
	return g
}

func (g *FunctionGenerator[V]) ToFunction(toFunc func(V) (Function[V], bool)) *FunctionGenerator[V] {
	g.toFunction = toFunc
	return g
}

func (g *FunctionGenerator[V]) AddStaticFunction(n string, f Function[V]) *FunctionGenerator[V] {
	g.staticFuncs[n] = f
	return g
}

func (g *FunctionGenerator[V]) AddConstant(n string, c V) *FunctionGenerator[V] {
	g.constants[n] = c
	return g
}

func (g *FunctionGenerator[V]) AddOp(operator string, impl func(a V, b V) V) *FunctionGenerator[V] {
	g.operators = append(g.operators, Operator[V]{
		Operator: operator,
		Impl:     impl,
	})
	return g
}

type Vars[V any] interface {
	Get(string) V
}

type Variables[V any] map[string]V

func (v Variables[V]) Get(k string) V {
	if va, ok := v[k]; ok {
		return va
	}
	panic(fmt.Sprintf("variable not found: %v", k))
}

type addVar[V any] struct {
	name   string
	val    V
	parent Vars[V]
}

func (a addVar[V]) Get(s string) V {
	if s == a.name {
		return a.val
	}
	return a.parent.Get(s)
}

type AddVars[V any] struct {
	Vars   map[string]V
	Parent Vars[V]
}

func (a AddVars[V]) Get(s string) V {
	if v, ok := a.Vars[s]; ok {
		return v
	}
	return a.Parent.Get(s)
}

type Code[V any] func(v Vars[V]) V

type Generator[V any] interface {
	Generate(AST, *FunctionGenerator[V]) Code[V]
}

func (g *FunctionGenerator[V]) Generate(exp string) (c func(Vars[V]) (V, error), errE error) {
	defer func() {
		rec := recover()
		if rec != nil {
			errE = fmt.Errorf("error generating code: %v", rec)
			c = nil
		}
	}()

	ast, err := g.CreateAst(exp)
	if err != nil {
		return nil, err
	}

	code := g.Gen(ast)
	return func(v Vars[V]) (res V, e error) {
		defer func() {
			rec := recover()
			if rec != nil {
				e = fmt.Errorf("error evaluating expression '%v': %v", exp, rec)
			}
		}()
		return code(v), nil
	}, nil
}

func (g *FunctionGenerator[V]) CreateAst(exp string) (AST, error) {
	ast, err := g.Parser().Parse(exp)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}

	if g.optimizer != nil {
		ast = Optimize(ast, g.optimizer)
	}
	return ast, nil
}

type Function[V any] struct {
	Func   func(a []V) V
	Args   int
	IsPure bool
}

type Closure[V any] struct {
	Names   []string
	Func    Code[V]
	Context Vars[V]
}

func (g *FunctionGenerator[V]) Gen(ast AST) Code[V] {
	if g.customGenerator != nil {
		c := g.customGenerator.Generate(ast, g)
		if c != nil {
			return c
		}
	}
	switch a := ast.(type) {
	case Ident:
		return func(v Vars[V]) V {
			return v.Get(string(a))
		}
	case Const[V]:
		n := a.Value
		return func(v Vars[V]) V {
			return n
		}
	case *Let:
		cVal := g.Gen(a.Value)
		inner := g.Gen(a.Inner)
		return func(v Vars[V]) V {
			va := cVal(v)
			return inner(addVar[V]{name: a.Name, val: va, parent: v})
		}
	case *Operate:
		ca := g.Gen(a.A)
		cb := g.Gen(a.B)
		op := g.opMap[a.Operator].Impl
		return func(v Vars[V]) V {
			return op(ca(v), cb(v))
		}
	case *ClosureLiteral:
		fu := g.Gen(a.Func)
		return func(v Vars[V]) V {
			return g.closureHandler.FromClosure(Closure[V]{Names: a.Names, Func: fu, Context: v})
		}
	case ListLiteral:
		if g.listHandler != nil {
			items := g.genList(a)
			return func(v Vars[V]) V {
				it := make([]V, len(items))
				for i, item := range items {
					it[i] = item(v)
				}
				return g.listHandler.FromList(it)
			}
		}
	case *ArrayAccess:
		if g.listHandler != nil {
			index := g.Gen(a.Index)
			list := g.Gen(a.List)
			return func(v Vars[V]) V {
				i := index(v)
				l := list(v)
				if v, err := g.listHandler.AccessList(l, i); err == nil {
					return v
				} else {
					panic(fmt.Sprint(err))
				}
			}
		}
	case MapLiteral:
		if g.mapHandler != nil {
			itemsCode := g.genMap(a)
			return func(v Vars[V]) V {
				return g.mapHandler.FromMap(evalMap(itemsCode, v))
			}
		}
	case *MapAccess:
		if g.mapHandler != nil {
			ma := g.Gen(a.MapValue)
			return func(v Vars[V]) V {
				l := ma(v)
				if v, err := g.mapHandler.AccessMap(l, a.Key); err == nil {
					return v
				} else {
					panic(fmt.Sprint(err))
				}
			}
		}
	case *FunctionCall:
		if id, ok := a.Func.(Ident); ok {
			if fun, ok := g.staticFuncs[string(id)]; ok {
				if fun.Args != len(a.Args) {
					panic(fmt.Errorf("wrong number of arguments in call to %v", id))
				}
				argsCode := g.genList(a.Args)
				return func(v Vars[V]) V {
					return fun.Func(evalList(argsCode, v))
				}
			}
		}
		fu := g.Gen(a.Func)
		argsCode := g.genList(a.Args)
		return func(v Vars[V]) V {
			theFunc, ok := g.extractFunction(fu(v))
			if !ok {
				panic(fmt.Errorf("not a function: %v", a.Func))
			}
			if theFunc.Args != len(a.Args) {
				panic(fmt.Errorf("wrong number of arguments in call to %v", a.Func))
			}
			return theFunc.Func(evalList(argsCode, v))
		}
	}
	panic(fmt.Sprintf("not supported: %v", ast))
}

func (g *FunctionGenerator[V]) genList(a []AST) []Code[V] {
	args := make([]Code[V], len(a))
	for i, arg := range a {
		args[i] = g.Gen(arg)
	}
	return args
}

func (g *FunctionGenerator[V]) genMap(a map[string]AST) map[string]Code[V] {
	args := map[string]Code[V]{}
	for i, arg := range a {
		args[i] = g.Gen(arg)
	}
	return args
}

func evalList[V any](argsCode []Code[V], v Vars[V]) []V {
	argsValues := make([]V, len(argsCode))
	for i, arg := range argsCode {
		argsValues[i] = arg(v)
	}
	return argsValues
}

func evalMap[V any](argsCode map[string]Code[V], v Vars[V]) map[string]V {
	argsValues := map[string]V{}
	for i, arg := range argsCode {
		argsValues[i] = arg(v)
	}
	return argsValues
}

func (g *FunctionGenerator[V]) extractFunction(fu V) (Function[V], bool) {
	if g.closureHandler != nil {
		if c, ok := g.closureHandler.ToClosure(fu); ok {
			return Function[V]{
				Func: func(args []V) V {
					vm := map[string]V{}
					for i, n := range c.Names {
						vm[n] = args[i]
					}
					return c.Func(AddVars[V]{
						Vars:   vm,
						Parent: c.Context,
					})
				},
				Args: len(c.Names),
			}, true
		}
	}
	if g.toFunction != nil {
		return g.toFunction(fu)
	}
	return Function[V]{}, false
}
