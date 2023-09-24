package stack

import (
	"fmt"
	"github.com/hneemann/parser2"
)

type Stack[V any] interface {
	Get(n int) V
	Push(V)
	Pop() V
	Size() int
	CreateFrame() Stack[V]
}

type SimpleStack[V any] []V

func (s SimpleStack[V]) Get(n int) V {
	return s[n]
}

func (s SimpleStack[V]) Size() int {
	return len(s)
}

func (s *SimpleStack[V]) Push(v V) {
	*s = append(*s, v)
}

func (s *SimpleStack[V]) Pop() V {
	last := len(*s) - 1
	v := (*s)[last]
	*s = (*s)[0:last]
	return v
}

func (s SimpleStack[V]) CreateFrame() Stack[V] {
	return &StackFrame[V]{
		offset: len(s),
		size:   0,
		parent: &s,
	}
}

type StackFrame[V any] struct {
	offset int
	size   int
	parent Stack[V]
}

func (s *StackFrame[V]) Get(n int) V {
	return s.parent.Get(s.offset + n)
}

func (s *StackFrame[V]) Size() int {
	return s.size
}

func (s *StackFrame[V]) Push(v V) {
	s.size++
	s.parent.Push(v)
}

func (s *StackFrame[V]) Pop() V {
	s.size--
	return s.parent.Pop()
}

func (s *StackFrame[V]) CreateFrame() Stack[V] {
	return &StackFrame[V]{
		offset: s.offset + s.size,
		size:   0,
		parent: s.parent,
	}
}

// Function represents a function
type Function[V any] struct {
	// Func is the function itself
	Func func(st Stack[V], closureStore []V) V
	// Args gives the number of arguments required. It is used for checking
	// the number of arguments in the call. The value -1 means any number of
	// arguments is allowed
	Args int
	// IsPure is true if this is a pure function
	IsPure bool
}

func (f Function[V]) Eval(a ...V) V {
	st := SimpleStack[V](a)
	return f.Func(&st, nil)
}

// ListHandler is used to create and access lists or arrays
type ListHandler[V any] interface {
	// FromList is used to convert a list to a value
	FromList(items []V) V
	// AccessList is used to get a value from a list
	AccessList(list V, index V) (V, error)
}

// MapHandler is used to create and access maps
type MapHandler[V any] interface {
	// FromMap creates a map
	FromMap(items map[string]V) V
	// AccessMap is used to get a value from a map
	AccessMap(m V, key string) (V, error)
	// IsMap is used to check if the given value is a map
	IsMap(value V) bool
}

// ClosureHandler is used to convert closures
type ClosureHandler[V any] interface {
	// FromClosure is used to convert a closure to a value
	FromClosure(c Function[V]) V
	// ToClosure is used to convert a value to a closure
	// It returns the closure and a bool which is true if the value was a closure
	ToClosure(c V) (Function[V], bool)
}

// Generator is used to define a customized generation of functions
type Generator[V any] interface {
	Generate(parser2.AST, ArgsMap, ArgsMap, *FunctionGenerator[V]) (Func[V], error)
}

type ToBool[V any] func(c V) bool

type IsEqual[V any] func(a, b V) bool

type constMap[V any] map[string]V

func (c constMap[V]) GetConst(name string) (V, bool) {
	v, ok := c[name]
	return v, ok
}

type FunctionGenerator[V any] struct {
	parser          *parser2.Parser[V]
	operators       []parser2.Operator[V]
	unary           []parser2.UnaryOperator[V]
	numberParser    parser2.NumberParser[V]
	stringHandler   parser2.StringConverter[V]
	listHandler     ListHandler[V]
	mapHandler      MapHandler[V]
	closureHandler  ClosureHandler[V]
	optimizer       parser2.Optimizer
	constants       constMap[V]
	toBool          ToBool[V]
	isEqual         IsEqual[V]
	staticFunctions map[string]Function[V]
	opMap           map[string]parser2.Operator[V]
	uMap            map[string]parser2.UnaryOperator[V]
	customGenerator Generator[V]
}

// New creates a new FunctionGenerator
func New[V any]() *FunctionGenerator[V] {
	g := &FunctionGenerator[V]{
		constants:       constMap[V]{},
		staticFunctions: make(map[string]Function[V]),
	}
	//g.optimizer = parser2.NewOptimizer(g)
	return g
}

func (g *FunctionGenerator[V]) SetNumberParser(numberParser parser2.NumberParser[V]) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.numberParser = numberParser
	return g
}

func (g *FunctionGenerator[V]) SetStringConverter(stringConverter parser2.StringConverter[V]) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.stringHandler = stringConverter
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

func (g *FunctionGenerator[V]) SetClosureHandler(closureHandler ClosureHandler[V]) *FunctionGenerator[V] {
	g.closureHandler = closureHandler
	return g
}

func (g *FunctionGenerator[V]) SetToBool(toBool ToBool[V]) *FunctionGenerator[V] {
	g.toBool = toBool
	return g
}

func (g *FunctionGenerator[V]) SetIsEqual(isEqual IsEqual[V]) *FunctionGenerator[V] {
	g.isEqual = isEqual
	return g
}

func (g *FunctionGenerator[V]) AddUnary(operator string, impl func(a V) V) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.unary = append(g.unary, parser2.UnaryOperator[V]{
		Operator: operator,
		Impl:     impl,
	})
	return g
}

// AddOp adds an operation to the generator.
// The Operation needs to be pure.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOp(operator string, isCommutative bool, impl func(a V, b V) V) *FunctionGenerator[V] {
	return g.AddOpPure(operator, isCommutative, impl, true)
}

// AddOpPure adds an operation to the generator.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOpPure(operator string, isCommutative bool, impl func(a V, b V) V, isPure bool) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.operators = append(g.operators, parser2.Operator[V]{
		Operator:      operator,
		Impl:          impl,
		IsPure:        isPure,
		IsCommutative: isCommutative,
	})
	return g
}

func (g *FunctionGenerator[V]) AddConstant(n string, c V) *FunctionGenerator[V] {
	g.constants[n] = c
	return g
}

func (g *FunctionGenerator[V]) AddSimpleFunction(name string, f func(V) V) *FunctionGenerator[V] {
	return g.AddStaticFunction(name, Function[V]{
		Func:   func(st Stack[V], cs []V) V { return f(st.Get(0)) },
		Args:   1,
		IsPure: true,
	})
}

func (g *FunctionGenerator[V]) AddStaticFunction(n string, f Function[V]) *FunctionGenerator[V] {
	g.staticFunctions[n] = f
	return g
}

func (g *FunctionGenerator[V]) SetOptimizer(optimizer parser2.Optimizer) *FunctionGenerator[V] {
	g.optimizer = optimizer
	return g
}

func (g *FunctionGenerator[V]) SetCustomGenerator(generator Generator[V]) *FunctionGenerator[V] {
	g.customGenerator = generator
	return g
}

func (g *FunctionGenerator[V]) getParser() *parser2.Parser[V] {
	if g.parser == nil {
		parser := parser2.NewParser[V]().
			SetNumberParser(g.numberParser).
			SetStringConverter(g.stringHandler).
			SetConstants(g.constants).
			SetOptimizer(g.optimizer)

		opMap := map[string]parser2.Operator[V]{}
		for _, o := range g.operators {
			parser.Op(o.Operator)
			opMap[o.Operator] = o
		}
		uMap := map[string]parser2.UnaryOperator[V]{}
		for _, u := range g.unary {
			parser.Unary(u.Operator)
			uMap[u.Operator] = u
		}

		g.parser = parser
		g.opMap = opMap
		g.uMap = uMap
	}
	return g.parser
}

type ArgsMap map[string]int

func (am ArgsMap) add(name string) ArgsMap {
	am[name] = len(am)
	return am
}

type Func[V any] func(stack Stack[V], closureStore []V) V

func (g *FunctionGenerator[V]) Generate(args []string, exp string) (func([]V) (V, error), error) {
	ast, err := g.CreateAst(exp)
	if err != nil {
		return nil, err
	}

	am := ArgsMap{}
	for _, a := range args {
		am.add(a)
	}

	f, err := g.GenerateFunc(ast, am, nil)
	if err != nil {
		return nil, err
	}
	return func(v []V) (val V, err error) {
		defer func() {
			rec := recover()
			if rec != nil {
				var zero V
				val = zero
				if e, ok := rec.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", rec)
				}
			}
		}()
		stack := SimpleStack[V](v)
		return f(&stack, nil), nil
	}, nil
}

// CreateAst uses the parser to create the abstract syntax tree.
// This method is public manly to inspect the AST in tests that live outside
// this package.
func (g *FunctionGenerator[V]) CreateAst(exp string) (parser2.AST, error) {
	ast, err := g.getParser().Parse(exp)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}

	if g.optimizer != nil {
		ast, err = parser2.Optimize(ast, g.optimizer)
		if err != nil {
			return nil, err
		}
	}
	return ast, nil
}

func (g *FunctionGenerator[V]) GenerateFunc(ast parser2.AST, am, cm ArgsMap) (Func[V], error) {
	if g.customGenerator != nil {
		c, err := g.customGenerator.Generate(ast, am, cm, g)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return c, nil
		}
	}
	switch a := ast.(type) {
	case *parser2.Const[V]:
		return func(st Stack[V], cs []V) V {
			return a.Value
		}, nil
	case *parser2.Ident:
		if index, ok := am[a.Name]; ok {
			return func(st Stack[V], cs []V) V {
				return st.Get(index)
			}, nil
		} else {
			if index, ok := cm[a.Name]; ok {
				return func(st Stack[V], cs []V) V {
					return cs[index]
				}, nil
			} else {
				return nil, a.Errorf("not found:%s", a.Name)
			}
		}
	case *parser2.Let:
		valFunc, err := g.GenerateFunc(a.Value, am, cm)
		if err != nil {
			return nil, err
		}
		mainFunc, err := g.GenerateFunc(a.Inner, am.add(a.Name), cm)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) V {
			va := valFunc(st, cs)
			st.Push(va)
			return mainFunc(st, cs)
		}, nil
	case *parser2.If:
		if g.toBool != nil {
			condFunc, err := g.GenerateFunc(a.Cond, am, cm)
			if err != nil {
				return nil, err
			}
			thenFunc, err := g.GenerateFunc(a.Then, am, cm)
			if err != nil {
				return nil, err
			}
			elseFunc, err := g.GenerateFunc(a.Else, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				if g.toBool(condFunc(st, cs)) {
					return thenFunc(st, cs)
				} else {
					return elseFunc(st, cs)
				}
			}, nil
		}
	case *parser2.Switch[V]:
		if g.isEqual != nil {
			switchValueFunc, err := g.GenerateFunc(a.SwitchValue, am, cm)
			if err != nil {
				return nil, err
			}
			defaultFunc, err := g.GenerateFunc(a.Default, am, cm)
			if err != nil {
				return nil, err
			}

			type caseFunc struct {
				constFunc  Func[V]
				resultFunc Func[V]
			}
			var cases []caseFunc
			for _, c := range a.Cases {
				constFunc, err := g.GenerateFunc(c.CaseConst, am, cm)
				if err != nil {
					return nil, err
				}
				resultFunc, err := g.GenerateFunc(c.Value, am, cm)
				if err != nil {
					return nil, err
				}
				cases = append(cases, caseFunc{
					constFunc:  constFunc,
					resultFunc: resultFunc,
				})
			}
			return func(st Stack[V], cs []V) V {
				val := switchValueFunc(st, cs)
				for _, c := range cases {
					if g.isEqual(val, c.constFunc(st, cs)) {
						return c.resultFunc(st, cs)
					}
				}
				return defaultFunc(st, cs)
			}, nil
		}
	case *parser2.Unary:
		valFunc, err := g.GenerateFunc(a.Value, am, cm)
		if err != nil {
			return nil, err
		}
		op := g.uMap[a.Operator].Impl
		return func(st Stack[V], cs []V) V {
			return op(valFunc(st, cs))
		}, nil

	case *parser2.Operate:
		aFunc, err := g.GenerateFunc(a.A, am, cm)
		if err != nil {
			return nil, err
		}
		bFunc, err := g.GenerateFunc(a.B, am, cm)
		if err != nil {
			return nil, err
		}
		op := g.opMap[a.Operator].Impl
		return func(st Stack[V], cs []V) V {
			return op(aFunc(st, cs), bFunc(st, cs))
		}, nil
	case *parser2.ClosureLiteral:
		funcArgs := ArgsMap{}
		for _, arg := range a.Names {
			funcArgs.add(arg)
		}
		usedVars := g.checkIfClosure(a.Func, funcArgs)
		if len(usedVars) == 0 {
			// not a closure, just a lambda
			closureFunc, err := g.GenerateFunc(a.Func, funcArgs, nil)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				return g.closureHandler.FromClosure(Function[V]{
					Func: closureFunc,
					Args: len(a.Names),
				})
			}, nil
		} else {
			// is a real closure
			closureFunc, err := g.GenerateFunc(a.Func, funcArgs, usedVars)
			if err != nil {
				return nil, err
			}

			type copyAction struct {
				index     int
				fromStack bool
			}
			copyActions := make([]copyAction, len(usedVars))
			for n, ci := range usedVars {
				if i, ok := am[n]; ok {
					copyActions[ci] = copyAction{
						index:     i,
						fromStack: true,
					}
				} else {
					if i, ok := cm[n]; ok {
						copyActions[ci] = copyAction{
							index:     i,
							fromStack: false,
						}
					} else {
						return nil, a.Errorf("not found: %s", n)
					}
				}
			}
			return func(st Stack[V], cs []V) V {
				closureStore := make([]V, len(copyActions))
				for i, ca := range copyActions {
					if ca.fromStack {
						closureStore[i] = st.Get(ca.index)
					} else {
						closureStore[i] = cs[ca.index]
					}
				}
				return g.closureHandler.FromClosure(Function[V]{
					Func: func(st Stack[V], cs []V) V {
						return closureFunc(st, closureStore)
					},
					Args: len(a.Names),
				})
			}, nil
		}
	case *parser2.ListLiteral:
		if g.listHandler != nil {
			itemFuncs, err := g.genFuncList(a.List, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				itemValues := make([]V, len(itemFuncs))
				for i, itemFunc := range itemFuncs {
					itemValues[i] = itemFunc(st, cs)
				}
				return g.listHandler.FromList(itemValues)
			}, nil
		}
	case *parser2.ListAccess:
		if g.listHandler != nil {
			indexFunc, err := g.GenerateFunc(a.Index, am, cm)
			if err != nil {
				return nil, err
			}
			listFunc, err := g.GenerateFunc(a.List, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				i := indexFunc(st, cs)
				l := listFunc(st, cs)
				if v, err := g.listHandler.AccessList(l, i); err == nil {
					return v
				} else {
					panic(a.EnhanceErrorf(err, "List error"))
				}
			}, nil
		}
	case *parser2.MapLiteral:
		if g.mapHandler != nil {
			itemsCode, err := g.genCodeMap(a.Map, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				mapValues := map[string]V{}
				for i, arg := range itemsCode {
					mapValues[i] = arg(st, cs)
				}
				return g.mapHandler.FromMap(mapValues)
			}, nil
		}
	case *parser2.MapAccess:
		if g.mapHandler != nil {
			mapFunc, err := g.GenerateFunc(a.MapValue, am, cm)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) V {
				l := mapFunc(st, cs)
				if v, err := g.mapHandler.AccessMap(l, a.Key); err == nil {
					return v
				} else {
					panic(a.EnhanceErrorf(err, "Map error"))
				}
			}, nil
		}
	case *parser2.FunctionCall:
		if id, ok := a.Func.(*parser2.Ident); ok {
			if fun, ok := g.staticFunctions[id.Name]; ok {
				if fun.Args >= 0 && fun.Args != len(a.Args) {
					return nil, id.Errorf("wrong number of arguments at call of %s, required %d, found %d", id.Name, fun.Args, len(a.Args))
				}
				argsFuncList, err := g.genFuncList(a.Args, am, cm)
				if err != nil {
					return nil, err
				}
				return func(st Stack[V], cs []V) V {
					sf := st.CreateFrame()
					for _, argFunc := range argsFuncList {
						sf.Push(argFunc(st, cs))
					}
					return fun.Func(sf, nil)
				}, nil
			}
		}
		funcFunc, err := g.GenerateFunc(a.Func, am, cm)
		if err != nil {
			return nil, err
		}
		argsFuncList, err := g.genFuncList(a.Args, am, cm)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) V {
			funcVal := funcFunc(st, cs)
			theFunc, ok := g.extractFunction(funcVal)
			if !ok {
				panic(a.Errorf("not a function: %v", a.Func))
			}
			if theFunc.Args >= 0 && theFunc.Args != len(a.Args) {
				panic(a.Errorf("wrong number of arguments at call of %v, required %d, found %d", a.Func, theFunc.Args, len(a.Args)))
			}
			sf := st.CreateFrame()
			for _, argFunc := range argsFuncList {
				sf.Push(argFunc(st, cs))
			}
			return theFunc.Func(sf, cs)
		}, nil
	}
	return nil, ast.GetLine().Errorf("not supported: %v", ast)
}

func (g *FunctionGenerator[V]) genFuncList(a []parser2.AST, am, cm ArgsMap) ([]Func[V], error) {
	args := make([]Func[V], len(a))
	for i, arg := range a {
		var err error
		args[i], err = g.GenerateFunc(arg, am, cm)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

// extractFunction is used to extract a function from a value
// Up to now only closures are supported.
func (g *FunctionGenerator[V]) extractFunction(fu V) (Function[V], bool) {
	if g.closureHandler != nil {
		if c, ok := g.closureHandler.ToClosure(fu); ok {
			return c, true
		}
	}
	return Function[V]{}, false
}

func (g *FunctionGenerator[V]) checkIfClosure(ast parser2.AST, args ArgsMap) ArgsMap {
	found := ArgsMap{}
	fna := findNonArgAccess{args: args, found: &found}
	ast.Traverse(&fna)
	return found
}

type findNonArgAccess struct {
	args  ArgsMap
	found *ArgsMap
}

func (f *findNonArgAccess) Visit(ast parser2.AST) bool {
	switch a := ast.(type) {
	case *parser2.Ident:
		if _, ok := f.args[a.Name]; !ok {
			(*f.found)[a.Name] = len(*f.found)
		}
		return false
	case *parser2.ClosureLiteral:
		args := ArgsMap{}
		for k, v := range f.args {
			args[k] = v
		}
		for _, n := range a.Names {
			args[n] = len(args)
		}
		a.Func.Traverse(&findNonArgAccess{args: args, found: f.found})
		return false
	case *parser2.Let:
		a.Value.Traverse(f)
		inner := ArgsMap{}
		for k, v := range f.args {
			inner[k] = v
		}
		inner.add(a.Name)
		a.Inner.Traverse(&findNonArgAccess{args: inner, found: f.found})
		return false
	}
	return true
}

func (g *FunctionGenerator[V]) genCodeMap(a map[string]parser2.AST, am, cm ArgsMap) (map[string]Func[V], error) {
	args := map[string]Func[V]{}
	for i, arg := range a {
		var err error
		args[i], err = g.GenerateFunc(arg, am, cm)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}
