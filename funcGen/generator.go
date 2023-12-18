package funcGen

import (
	"bytes"
	"fmt"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/listMap"
	"log"
	"reflect"
	"sort"
	"unicode"
	"unicode/utf8"
)

type stackStorage[V any] struct {
	data []V
}

func (s *stackStorage[V]) set(n int, v V) {
	if n == len(s.data) {
		s.data = append(s.data, v)
	} else {
		s.data[n] = v
	}
}

func (s *stackStorage[V]) get(n int) V {
	return s.data[n]
}

type Stack[V any] struct {
	storage *stackStorage[V]
	offs    int
	size    int
}

func NewEmptyStack[V any]() Stack[V] {
	return Stack[V]{
		storage: &stackStorage[V]{data: make([]V, 0, 50)},
		offs:    0,
		size:    0,
	}
}

func NewStack[V any](v ...V) Stack[V] {
	return Stack[V]{
		storage: &stackStorage[V]{data: v},
		offs:    0,
		size:    len(v),
	}
}

func (s Stack[V]) ToSlice() []V {
	return s.storage.data[s.offs : s.offs+s.size]
}

func (s Stack[V]) Get(n int) V {
	return s.storage.get(s.offs + n)
}

func (s Stack[V]) Size() int {
	return s.size
}

func (s *Stack[V]) Push(v V) {
	s.storage.set(s.offs+s.size, v)
	s.size++
}

func (s *Stack[V]) CreateFrame(size int) Stack[V] {
	s.size -= size
	st := Stack[V]{
		storage: s.storage,
		offs:    s.offs + s.size,
		size:    size,
	}
	return st
}

func (s *Stack[V]) Init(v ...V) {
	s.offs = 0
	s.size = 0
	for _, a := range v {
		s.Push(a)
	}
}

// Operator defines a operator like +
type Operator[V any] struct {
	// Operator is the operator as a string like "+"
	Operator string
	// Impl is the implementation of the operation
	Impl func(st Stack[V], a, b V) (V, error)
	// IsPure is true if the result of the operation depends only on the operands.
	// This is usually the case, there are only special corner cases where it is not.
	// So IsPure is usually true.
	IsPure bool
	// IsCommutative is true if the operation is commutative
	IsCommutative bool
}

// UnaryOperator defines a operator like - or !
type UnaryOperator[V any] struct {
	// Operator is the operator as a string like "+"
	Operator string
	// Impl is the implementation of the operation
	Impl func(a V) (V, error)
}

// Func is the signature of the go closures created to build the
// generated function. The stack is used to store arguments and local
// variables created by let, and the closureStore is used to pass the
// accessed outer values to the function.
type Func[V any] func(stack Stack[V], closureStore []V) (V, error)

type FunctionDescription struct {
	Args        []string
	Description string
}

func (f *FunctionDescription) WriteTo(b *bytes.Buffer, name string) {
	b.WriteString(name)
	if f == nil {
		b.WriteRune('\n')
		return
	}

	b.WriteString("(")
	for i, a := range f.Args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(a)
	}
	b.WriteString(")\n\t")
	pos := 0
	var word bytes.Buffer
	appendWord := func() {
		if word.Len() > 0 {
			if pos+word.Len() > 70 {
				b.WriteString("\n\t")
				pos = 0
			} else {
				if pos > 0 {
					b.WriteRune(' ')
					pos++
				}
			}
			pos += word.Len()
			word.WriteTo(b)
		}
	}
	for _, c := range f.Description {
		if unicode.IsSpace(c) {
			appendWord()
		} else {
			word.WriteRune(c)
		}
	}
	appendWord()
}

// Function represents a function
type Function[V any] struct {
	// Func is the function itself
	Func Func[V]
	// Args gives the number of arguments required. It is used for checking
	// the number of arguments in the call. The value -1 means any number of
	// arguments is allowed
	Args int
	// IsPure is true if this is a pure function
	IsPure bool
	// Description is a description of the function
	Description *FunctionDescription
}

func (f Function[V]) SetMethodDescription(descr ...string) Function[V] {
	if f.Args > 0 && f.Args != len(descr) {
		panic(fmt.Errorf("wrong number of arguments in description: %d, expected %d", len(descr), f.Args))
	}
	f.Description = &FunctionDescription{
		Args:        descr[:len(descr)-1],
		Description: descr[len(descr)-1],
	}
	return f
}

func (f Function[V]) SetDescription(descr ...string) Function[V] {
	if f.Args >= 0 && f.Args+1 != len(descr) {
		panic(fmt.Errorf("wrong number of arguments in description: %d, expected %d", len(descr), f.Args+1))
	}
	f.Description = &FunctionDescription{
		Args:        descr[:len(descr)-1],
		Description: descr[len(descr)-1],
	}
	return f
}

// Eval is used to evaluate a function with one argument
// The stack [st] is used to pass the given argument [a] to the function.
// The pushed value is removed after the function is called.
func (f Function[V]) Eval(st Stack[V], a V) (V, error) {
	st.Push(a)
	return f.Func(st.CreateFrame(1), nil)
}

// EvalSt is used to evaluate a function with multiple arguments
// The stack [st] is used to pass the given arguments to the function.
// The pushed values are removed after the function is called.
func (f Function[V]) EvalSt(st Stack[V], a ...V) (V, error) {
	for _, e := range a {
		st.Push(e)
	}
	return f.Func(st.CreateFrame(len(a)), nil)
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
	FromMap(items listMap.ListMap[V]) V
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

// MethodHandler is used to give access to methods.
type MethodHandler[V any] interface {
	// GetMethod is used to get a method on a value.
	// The value is the first argument at calling the function.
	GetMethod(value V, methodName string) (Function[V], error)
}

type MethodHandlerFunc[V any] func(value V, methodName string) (Function[V], error)

func (mh MethodHandlerFunc[V]) GetMethod(value V, methodName string) (Function[V], error) {
	return mh(value, methodName)
}

// Generator is used to define a customized generation of functions
type Generator[V any] interface {
	Generate(parser2.AST, GeneratorContext, *FunctionGenerator[V]) (Func[V], error)
}

type ToBool[V any] func(c V) (bool, bool)

type IsEqual[V any] func(st Stack[V], a, b V) (bool, error)

type constMap[V any] map[string]V

func (c constMap[V]) GetConst(name string) (V, bool) {
	v, ok := c[name]
	return v, ok
}

type FunctionGenerator[V any] struct {
	parser          *parser2.Parser[V]
	operators       []Operator[V]
	unary           []UnaryOperator[V]
	numberParser    parser2.NumberParser[V]
	stringHandler   parser2.StringConverter[V]
	listHandler     ListHandler[V]
	mapHandler      MapHandler[V]
	closureHandler  ClosureHandler[V]
	methodHandler   MethodHandler[V]
	optimizer       parser2.Optimizer
	constants       constMap[V]
	toBool          ToBool[V]
	isEqual         IsEqual[V]
	staticFunctions map[string]Function[V]
	opMap           map[string]Operator[V]
	uMap            map[string]UnaryOperator[V]
	customGenerator Generator[V]
}

// New creates a new FunctionGenerator
func New[V any]() *FunctionGenerator[V] {
	g := &FunctionGenerator[V]{
		constants:       constMap[V]{},
		staticFunctions: make(map[string]Function[V]),
		methodHandler:   MethodHandlerFunc[V](methodByReflection[V]),
	}
	g.optimizer = NewOptimizer(NewEmptyStack[V](), g)
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

func (g *FunctionGenerator[V]) SetMethodHandler(methodHandler MethodHandler[V]) *FunctionGenerator[V] {
	g.methodHandler = methodHandler
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

func (g *FunctionGenerator[V]) AddUnary(operator string, impl func(a V) (V, error)) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	uni := UnaryOperator[V]{
		Operator: operator,
		Impl:     impl,
	}
	for i, u := range g.unary {
		if u.Operator == operator {
			g.unary[i] = uni
			return g
		}
	}
	g.unary = append(g.unary, uni)
	return g
}

// AddSimpleOp adds an operation to the generator.
// The Operation needs to be pure.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddSimpleOp(operator string, isCommutative bool, impl func(a V, b V) (V, error)) *FunctionGenerator[V] {
	return g.AddOpPure(operator, isCommutative, func(st Stack[V], a V, b V) (V, error) {
		return impl(a, b)
	}, true)
}

// AddOp adds an operation to the generator.
// The Operation needs to be pure.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOp(operator string, isCommutative bool, impl func(st Stack[V], a V, b V) (V, error)) *FunctionGenerator[V] {
	return g.AddOpPure(operator, isCommutative, impl, true)
}

// AddOpPure adds an operation to the generator.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOpPure(operator string, isCommutative bool, impl func(st Stack[V], a V, b V) (V, error), isPure bool) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}

	opItem := Operator[V]{
		Operator:      operator,
		Impl:          impl,
		IsPure:        isPure,
		IsCommutative: isCommutative,
	}

	for i, op := range g.operators {
		if op.Operator == operator {
			g.operators[i] = opItem
			return g
		}
	}
	g.operators = append(g.operators, opItem)
	return g
}

func (g *FunctionGenerator[V]) AddConstant(n string, c V) *FunctionGenerator[V] {
	g.constants[n] = c
	return g
}

func (g *FunctionGenerator[V]) AddSimpleFunction(name string, f func(V) V) *FunctionGenerator[V] {
	return g.AddStaticFunction(name, Function[V]{
		Func:   func(st Stack[V], cs []V) (V, error) { return f(st.Get(0)), nil },
		Args:   1,
		IsPure: true,
	})
}

func (g *FunctionGenerator[V]) AddGoFunction(name string, args int, f func(a ...V) (V, error)) *FunctionGenerator[V] {
	return g.AddStaticFunction(name, Function[V]{
		Func:   func(st Stack[V], cs []V) (V, error) { return f(st.ToSlice()...) },
		Args:   args,
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

func (g *FunctionGenerator[V]) ModifyParser(modify func(a *parser2.Parser[V])) *FunctionGenerator[V] {
	modify(g.getParser())
	return g
}

func (g *FunctionGenerator[V]) getParser() *parser2.Parser[V] {
	if g.parser == nil {
		parser := parser2.NewParser[V]().
			SetNumberParser(g.numberParser).
			SetStringConverter(g.stringHandler).
			SetConstants(g.constants).
			SetOptimizer(g.optimizer)

		opMap := map[string]Operator[V]{}
		for _, o := range g.operators {
			parser.Op(o.Operator)
			opMap[o.Operator] = o
		}
		uMap := map[string]UnaryOperator[V]{}
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

type argsMap map[string]int

func (am argsMap) add(name string) error {
	if name == "" {
		return fmt.Errorf("empty names are not allowed")
	}
	if _, ok := am[name]; ok {
		return fmt.Errorf("variable redeclared: %s", name)
	}
	am[name] = len(am)
	return nil
}

func (am argsMap) copyAndAdd(name string) (argsMap, error) {
	n := argsMap{}
	for k, v := range am {
		n[k] = v
	}
	err := n.add(name)
	if err != nil {
		return nil, err
	}
	return n, nil
}

type GeneratorContext struct {
	am       argsMap
	cm       argsMap
	ThisName string
}

func (c GeneratorContext) addLocalVar(name string) (GeneratorContext, error) {
	newAm, err := c.am.copyAndAdd(name)
	if err != nil {
		return GeneratorContext{}, err
	}
	return GeneratorContext{am: newAm, cm: c.cm, ThisName: c.ThisName}, nil
}

func (g *FunctionGenerator[V]) Generate(exp string, args ...string) (func(Stack[V]) (V, error), error) {
	return g.generateIntern(args, exp, "")
}

func (g *FunctionGenerator[V]) GenerateWithMap(exp string, mapName string) (func(Stack[V]) (V, error), error) {
	return g.generateIntern([]string{mapName}, exp, mapName)
}

func (g *FunctionGenerator[V]) generateIntern(args []string, exp string, ThisName string) (func(Stack[V]) (V, error), error) {
	ast, err := g.CreateAst(exp)
	if err != nil {
		return nil, err
	}

	am := argsMap{}
	if args != nil {
		for _, a := range args {
			err = am.add(a)
			if err != nil {
				return nil, err
			}
		}
	}

	gc := GeneratorContext{am: am, cm: nil, ThisName: ThisName}

	f, err := g.GenerateFunc(ast, gc)
	if err != nil {
		return nil, err
	}
	return func(st Stack[V]) (val V, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Print("panic in function: ", rec)
				var zero V
				val = zero
				err = parser2.AnyToError(rec)
			}
		}()
		return f(st, nil)
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

func (g *FunctionGenerator[V]) GenerateFunc(ast parser2.AST, gc GeneratorContext) (Func[V], error) {
	var zero V
	if g.customGenerator != nil {
		c, err := g.customGenerator.Generate(ast, gc, g)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return c, nil
		}
	}
	switch a := ast.(type) {
	case *parser2.Const[V]:
		return func(st Stack[V], cs []V) (V, error) {
			return a.Value, nil
		}, nil
	case *parser2.Ident:
		if index, ok := gc.am[a.Name]; ok {
			return func(st Stack[V], cs []V) (V, error) {
				return st.Get(index), nil
			}, nil
		} else {
			if index, ok := gc.cm[a.Name]; ok {
				return func(st Stack[V], cs []V) (V, error) {
					return cs[index], nil
				}, nil
			} else {
				if gc.ThisName != "" && g.mapHandler != nil {
					if index, ok := gc.am[gc.ThisName]; ok {
						return func(st Stack[V], cs []V) (V, error) {
							this := st.Get(index)
							if v, err := g.mapHandler.AccessMap(this, a.Name); err == nil {
								return v, nil
							} else {
								var zero V
								return zero, a.EnhanceErrorf(err, "Map error")
							}
						}, nil
					}
				}
				return nil, a.Errorf("not found: %s", a.Name)
			}
		}
	case *parser2.Let:
		var err error
		var valFunc Func[V]
		if c, ok := a.Value.(*parser2.ClosureLiteral); ok {
			uses := g.checkIfClosure(a, argsMap{})
			if _, ok := uses[a.Name]; ok {
				funcArgs := argsMap{}
				for _, arg := range c.Names {
					err := funcArgs.add(arg)
					if err != nil {
						return nil, err
					}
				}
				usedVars := g.checkIfClosure(c.Func, funcArgs)
				valFunc, err = g.createClosureLiteralFunc(c, GeneratorContext{am: funcArgs, cm: usedVars}, gc, a.Name)
				if err != nil {
					return nil, err
				}
			}
		}
		if valFunc == nil {
			valFunc, err = g.GenerateFunc(a.Value, gc)
			if err != nil {
				return nil, err
			}
		}
		newGc, err := gc.addLocalVar(a.Name)
		if err != nil {
			return nil, a.EnhanceErrorf(err, "error in let")
		}
		mainFunc, err := g.GenerateFunc(a.Inner, newGc)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) (V, error) {
			va, err := valFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in let")
			}
			st.Push(va)
			return mainFunc(st, cs)
		}, nil
	case *parser2.If:
		if g.toBool != nil {
			condFunc, err := g.GenerateFunc(a.Cond, gc)
			if err != nil {
				return nil, err
			}
			thenFunc, err := g.GenerateFunc(a.Then, gc)
			if err != nil {
				return nil, err
			}
			elseFunc, err := g.GenerateFunc(a.Else, gc)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				condVal, err := condFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in if")
				}
				if cond, ok := g.toBool(condVal); ok {
					if cond {
						return thenFunc(st, cs)
					} else {
						return elseFunc(st, cs)
					}
				} else {
					return zero, a.Errorf("if condition is not a bool")
				}
			}, nil
		}
	case *parser2.Switch[V]:
		if g.isEqual != nil {
			switchValueFunc, err := g.GenerateFunc(a.SwitchValue, gc)
			if err != nil {
				return nil, err
			}
			defaultFunc, err := g.GenerateFunc(a.Default, gc)
			if err != nil {
				return nil, err
			}

			type caseFunc struct {
				constFunc  Func[V]
				resultFunc Func[V]
			}
			var cases []caseFunc
			for _, c := range a.Cases {
				constFunc, err := g.GenerateFunc(c.CaseConst, gc)
				if err != nil {
					return nil, err
				}
				resultFunc, err := g.GenerateFunc(c.Value, gc)
				if err != nil {
					return nil, err
				}
				cases = append(cases, caseFunc{
					constFunc:  constFunc,
					resultFunc: resultFunc,
				})
			}
			return func(st Stack[V], cs []V) (V, error) {
				val, err := switchValueFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in switch")
				}
				for _, c := range cases {
					constval, err := c.constFunc(st, cs)
					if err != nil {
						return zero, a.EnhanceErrorf(err, "error in switch-case")
					}
					equal, err := g.isEqual(st, val, constval)
					if err != nil {
						return zero, a.EnhanceErrorf(err, "error in switch-case")
					}
					if equal {
						return c.resultFunc(st, cs)
					}
				}
				return defaultFunc(st, cs)
			}, nil
		}
	case *parser2.TryCatch:
		tryFunc, err := g.GenerateFunc(a.Try, gc)
		if err != nil {
			return nil, err
		}
		catchFunc, err := g.GenerateFunc(a.Catch, gc)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) (V, error) {
			v, err := tryFunc(st, cs)
			if err == nil {
				return v, nil
			}
			v2, err := catchFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in catch")
			}
			return v2, err
		}, nil
	case *parser2.Unary:
		valFunc, err := g.GenerateFunc(a.Value, gc)
		if err != nil {
			return nil, err
		}
		op := g.uMap[a.Operator].Impl
		return func(st Stack[V], cs []V) (V, error) {
			v, err := valFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in unary %v", a.Operator)
			}
			return op(v)
		}, nil
	case *parser2.Operate:
		aFunc, err := g.GenerateFunc(a.A, gc)
		if err != nil {
			return nil, err
		}
		bFunc, err := g.GenerateFunc(a.B, gc)
		if err != nil {
			return nil, err
		}
		op := g.opMap[a.Operator].Impl
		return func(st Stack[V], cs []V) (V, error) {
			aVal, err := aFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in operation %v", a.Operator)
			}
			bVal, err := bFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in operation %v", a.Operator)
			}
			return op(st, aVal, bVal)
		}, nil
	case *parser2.ClosureLiteral:
		funcArgs := argsMap{}
		for _, arg := range a.Names {
			err := funcArgs.add(arg)
			if err != nil {
				return nil, err
			}
		}
		usedVars := g.checkIfClosure(a.Func, funcArgs)
		if len(usedVars) == 0 {
			// not a closure, just a function
			closureFunc, err := g.GenerateFunc(a.Func, GeneratorContext{am: funcArgs})
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				return g.closureHandler.FromClosure(Function[V]{
					Func: closureFunc,
					Args: len(a.Names),
				}), nil
			}, nil
		} else {
			// is a real closure
			return g.createClosureLiteralFunc(a, GeneratorContext{
				am: funcArgs,
				cm: usedVars,
			}, gc, "")
		}
	case *parser2.ListLiteral:
		if g.listHandler != nil {
			itemFuncs, err := g.genFuncList(a.List, gc)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				itemValues := make([]V, len(itemFuncs))
				for i, itemFunc := range itemFuncs {
					v, err := itemFunc(st, cs)
					if err != nil {
						return zero, a.EnhanceErrorf(err, "List literal error")
					}
					itemValues[i] = v
				}
				return g.listHandler.FromList(itemValues), nil
			}, nil
		}
	case *parser2.ListAccess:
		if g.listHandler != nil {
			indexFunc, err := g.GenerateFunc(a.Index, gc)
			if err != nil {
				return nil, err
			}
			listFunc, err := g.GenerateFunc(a.List, gc)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				i, err := indexFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in list index")
				}
				l, err := listFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in getting list")
				}
				return g.listHandler.AccessList(l, i)
			}, nil
		}
	case *parser2.MapLiteral:
		if g.mapHandler != nil {
			itemsCode, err := g.genCodeMap(a.Map, gc)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				mapValues := listMap.New[V](len(itemsCode))
				var innerError error
				itemsCode.Iter(func(key string, value Func[V]) bool {
					var v V
					v, innerError = value(st, cs)
					if innerError != nil {
						return false
					}
					mapValues = mapValues.Append(key, v)
					return true
				})
				if innerError != nil {
					return zero, a.EnhanceErrorf(innerError, "Map literal error")
				}
				return g.mapHandler.FromMap(mapValues), nil
			}, nil
		}
	case *parser2.MapAccess:
		if g.mapHandler != nil {
			mapFunc, err := g.GenerateFunc(a.MapValue, gc)
			if err != nil {
				return nil, err
			}
			return func(st Stack[V], cs []V) (V, error) {
				l, err := mapFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in getting map")
				}
				return g.mapHandler.AccessMap(l, a.Key)
			}, nil
		}
	case *parser2.FunctionCall:
		if id, ok := a.Func.(*parser2.Ident); ok {
			if fun, ok := g.staticFunctions[id.Name]; ok {
				if fun.Args >= 0 && fun.Args != len(a.Args) {
					return nil, id.Errorf("wrong number of arguments at call of %s, required %d, found %d", id.Name, fun.Args, len(a.Args))
				}
				argsFuncList, err := g.genFuncList(a.Args, gc)
				if err != nil {
					return nil, err
				}
				return func(st Stack[V], cs []V) (V, error) {
					for _, argFunc := range argsFuncList {
						v, err := argFunc(st, cs)
						if err != nil {
							return zero, a.EnhanceErrorf(err, "error in function call to %s", id.Name)
						}
						st.Push(v)
					}
					return fun.Func(st.CreateFrame(len(argsFuncList)), nil)
				}, nil
			}
		}
		funcFunc, err := g.GenerateFunc(a.Func, gc)
		if err != nil {
			return nil, g.generateStaticFunctionDocu(err)
		}
		argsFuncList, err := g.genFuncList(a.Args, gc)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) (V, error) {
			funcVal, err := funcFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in getting function")
			}
			theFunc, ok := g.ExtractFunction(funcVal)
			if !ok {
				return zero, a.Errorf("not a function: %v", a.Func)
			}
			if theFunc.Args >= 0 && theFunc.Args != len(a.Args) {
				return zero, a.Errorf("wrong number of arguments at call of %v, required %d, found %d", a.Func, theFunc.Args, len(a.Args))
			}
			for _, argFunc := range argsFuncList {
				v, err := argFunc(st, cs)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error in arguments in function call to %v", a.Func)
				}
				st.Push(v)
			}
			return theFunc.Func(st.CreateFrame(len(argsFuncList)), cs)
		}, nil
	case *parser2.MethodCall:
		valFunc, err := g.GenerateFunc(a.Value, gc)
		if err != nil {
			return nil, err
		}
		name := a.Name
		argsFuncList, err := g.genFuncList(a.Args, gc)
		if err != nil {
			return nil, err
		}
		return func(st Stack[V], cs []V) (V, error) {
			value, err := valFunc(st, cs)
			if err != nil {
				return zero, a.EnhanceErrorf(err, "error in method call to %s", name)
			}
			// name could be a method, but it could also be the name of a field which stores a closure
			// If it is a closure field, this should be a map access!
			if g.mapHandler != nil && g.mapHandler.IsMap(value) {
				if va, err := g.mapHandler.AccessMap(value, name); err == nil {
					if theFunc, ok := g.ExtractFunction(va); ok {
						for _, argFunc := range argsFuncList {
							v, err := argFunc(st, cs)
							if err != nil {
								return zero, a.EnhanceErrorf(err, "error in arguments in method call to %s", name)
							}
							st.Push(v)
						}
						return theFunc.Func(st.CreateFrame(len(argsFuncList)), cs)
					}
				}
			}
			if g.methodHandler != nil {
				me, err := g.methodHandler.GetMethod(value, name)
				if err != nil {
					return zero, a.EnhanceErrorf(err, "error accessing method %s", name)
				}
				if me.Args > 0 && me.Args != len(argsFuncList)+1 {
					return zero, a.Errorf("wrong number of arguments at call of %s, required %d, found %d", name, me.Args-1, len(argsFuncList))
				}
				st.Push(value)
				for _, arg := range argsFuncList {
					v, err := arg(st, cs)
					if err != nil {
						return zero, a.EnhanceErrorf(err, "error in arguments in method call to %s", name)
					}
					st.Push(v)
				}
				return me.Func(st.CreateFrame(len(argsFuncList)+1), nil)
			}
			return zero, a.Errorf("method %s not found", name)
		}, nil
	}
	return nil, ast.GetLine().Errorf("not supported: %v", ast)
}

func (g *FunctionGenerator[V]) createClosureLiteralFunc(a *parser2.ClosureLiteral, innerContext GeneratorContext, gc GeneratorContext, recursiveName string) (Func[V], error) {
	closureFunc, err := g.GenerateFunc(a.Func, innerContext)
	if err != nil {
		return nil, err
	}

	type copyMode int
	const (
		stack copyMode = iota
		closure
		this
	)

	type copyAction struct {
		index int
		mode  copyMode
	}
	copyActions := make([]copyAction, len(innerContext.cm))
	for n, ci := range innerContext.cm {
		if i, ok := gc.am[n]; ok {
			copyActions[ci] = copyAction{
				index: i,
				mode:  stack,
			}
		} else {
			if i, ok := gc.cm[n]; ok {
				copyActions[ci] = copyAction{
					index: i,
					mode:  closure,
				}
			} else {
				if n == recursiveName {
					copyActions[ci] = copyAction{
						mode: this,
					}
				} else {
					return nil, a.Errorf("not found: %s", n)
				}
			}
		}
	}
	return func(st Stack[V], cs []V) (V, error) {
		closureStore := make([]V, len(copyActions))
		cl := g.closureHandler.FromClosure(Function[V]{
			Func: func(st Stack[V], cs []V) (V, error) {
				return closureFunc(st, closureStore)
			},
			Args: len(a.Names),
		})
		for i, ca := range copyActions {
			switch ca.mode {
			case stack:
				closureStore[i] = st.Get(ca.index)
			case closure:
				closureStore[i] = cs[ca.index]
			case this:
				closureStore[i] = cl
			}
		}
		return cl, nil
	}, nil
}

func (g *FunctionGenerator[V]) genFuncList(a []parser2.AST, gc GeneratorContext) ([]Func[V], error) {
	args := make([]Func[V], len(a))
	for i, arg := range a {
		var err error
		args[i], err = g.GenerateFunc(arg, gc)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

// ExtractFunction is used to extract a function from a value
// Up to now only closures are supported.
func (g *FunctionGenerator[V]) ExtractFunction(fu V) (Function[V], bool) {
	if g.closureHandler != nil {
		if c, ok := g.closureHandler.ToClosure(fu); ok {
			return c, true
		}
	}
	return Function[V]{}, false
}

func (g *FunctionGenerator[V]) checkIfClosure(ast parser2.AST, args argsMap) argsMap {
	found := argsMap{}
	fna := findNonArgAccess[V]{args: args, found: &found, staticFunc: g.staticFunctions}
	ast.Traverse(&fna)
	return found
}

type findNonArgAccess[V any] struct {
	args       argsMap
	found      *argsMap
	staticFunc map[string]Function[V]
}

func (f *findNonArgAccess[V]) inner(args argsMap) *findNonArgAccess[V] {
	return &findNonArgAccess[V]{args: args, found: f.found, staticFunc: f.staticFunc}
}

func (f *findNonArgAccess[V]) Visit(ast parser2.AST) bool {
	switch a := ast.(type) {
	case *parser2.Ident:
		if _, ok := f.args[a.Name]; !ok {
			if _, ok := (*f.found)[a.Name]; !ok {
				(*f.found)[a.Name] = len(*f.found)
			}
		}
		return false
	case *parser2.ClosureLiteral:
		innerArgs := argsMap{}
		for k, v := range f.args {
			innerArgs[k] = v
		}
		for _, n := range a.Names {
			innerArgs[n] = len(innerArgs)
		}
		a.Func.Traverse(f.inner(innerArgs))
		return false
	case *parser2.FunctionCall:
		if id, ok := a.Func.(*parser2.Ident); ok {
			if _, ok := f.staticFunc[id.Name]; !ok {
				a.Func.Traverse(f)
			}
		} else {
			a.Func.Traverse(f)
		}
		for _, ar := range a.Args {
			ar.Traverse(f)
		}
		return false
	case *parser2.Let:
		a.Value.Traverse(f)
		innerArgs := argsMap{}
		for k, v := range f.args {
			innerArgs[k] = v
		}
		innerArgs.add(a.Name)
		a.Inner.Traverse(f.inner(innerArgs))
		return false
	}
	return true
}

func (g *FunctionGenerator[V]) genCodeMap(a listMap.ListMap[parser2.AST], gc GeneratorContext) (args listMap.ListMap[Func[V]], err error) {
	args = listMap.New[Func[V]](a.Size())
	a.Iter(func(key string, value parser2.AST) bool {
		var f Func[V]
		f, err = g.GenerateFunc(value, gc)
		if err != nil {
			args = nil
			return false
		}
		args = args.Append(key, f)
		return true
	})
	return
}

func (g *FunctionGenerator[V]) generateStaticFunctionDocu(err error) error {
	type sf struct {
		name string
		f    Function[V]
	}
	var list []sf
	for n, f := range g.staticFunctions {
		list = append(list, sf{name: n, f: f})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].name < list[j].name
	})

	var b bytes.Buffer
	for _, f := range list {
		b.WriteRune('\n')
		f.f.Description.WriteTo(&b, f.name)
	}
	return fmt.Errorf("%w\n\nAvailable functions are:%s", err, b.String())
}

func methodByReflection[V any](value V, name string) (Function[V], error) {
	name = firstRuneUpper(name)
	typeOf := reflect.TypeOf(value)
	if m, ok := typeOf.MethodByName(name); ok {
		err := matches[V](m)
		if err != nil {
			return Function[V]{}, err
		}

		return Function[V]{
			Func: func(st Stack[V], cs []V) (V, error) {
				argsValues := make([]reflect.Value, st.Size())
				for i := 0; i < st.Size(); i++ {
					argsValues[i] = reflect.ValueOf(st.Get(i))
				}

				res := m.Func.Call(argsValues)
				if v, ok := res[0].Interface().(V); ok {
					return v, nil
				} else {
					var zero V
					return zero, fmt.Errorf("result of method %s is not a value. It is: %v", name, res[0])
				}
			},
			Args:   m.Type.NumIn(),
			IsPure: false,
		}, nil
	} else {
		var buf bytes.Buffer
		for i := 0; i < typeOf.NumMethod(); i++ {
			m := typeOf.Method(i)
			if matches[V](m) == nil {
				if buf.Len() > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(m.Name)
				buf.WriteString("(")
				mt := m.Func.Type()
				for i := 1; i < mt.NumIn(); i++ {
					if i > 1 {
						buf.WriteString(", ")
					}
					buf.WriteString(mt.In(i).Name())
				}
				buf.WriteString(")")
			}
		}
		return Function[V]{}, fmt.Errorf("method %s not found on %v, available are: %v", name, typeOf, buf.String())
	}
}

func firstRuneUpper(name string) string {
	r, l := utf8.DecodeRune([]byte(name))
	if unicode.IsUpper(r) {
		return name
	}
	return string(unicode.ToUpper(r)) + name[l:]
}

func matches[V any](m reflect.Method) error {
	typeOfV := reflect.TypeOf((*V)(nil)).Elem()
	mt := m.Func.Type()
	for i := 1; i < mt.NumIn(); i++ {
		if !typeOfV.AssignableTo(mt.In(i)) {
			return fmt.Errorf("type %v does not match %v", mt.In(i), typeOfV)
		}
	}
	if mt.NumOut() != 1 {
		return fmt.Errorf("wrong number of return values: found %d, want 1", mt.NumOut())
	}
	if !mt.Out(0).AssignableTo(typeOfV) {
		return fmt.Errorf("first return value needs to be assignable to %v", typeOfV)
	}
	return nil
}

/*
func firstRuneLower(name string) string {
	r, l := utf8.DecodeRune([]byte(name))
	if unicode.IsLower(r) {
		return name
	}
	return string(unicode.ToLower(r)) + name[l:]
}

func PrintMatchingCode[V any](v V) {
	t := reflect.TypeOf((*V)(nil)).Elem()
	typeName := t.Name()
	typeOfV := reflect.TypeOf(v)

	typeOfVName := typeOfV.Name()
	mapName := typeOfVName + "MethodMap"
	if typeOfV.Kind() == reflect.Pointer {
		typeOfVName = "*" + typeOfV.Elem().Name()
		mapName = typeOfV.Elem().Name() + "MethodMap"
	}

	fmt.Printf("\n\nvar %s=map[string]funcGen.Function[%s]{\n", firstRuneLower(mapName), typeName)
	for i := 0; i < typeOfV.NumMethod(); i++ {
		m := typeOfV.Method(i)
		if matches[V](m) == nil {
			methodName := firstRuneLower(m.Name)
			mt := m.Func.Type()

			fmt.Printf("  \"%s\": {\n", methodName)
			fmt.Printf("    Func:func (st funcGen.Stack[%[1]v], cs []%[1]v) %[1]v {\n", typeName)
			fmt.Printf("      return (st.Get(0).(%v)).%s(", typeOfVName, m.Name)
			for j := 1; j < mt.NumIn(); j++ {
				if j > 1 {
					fmt.Print(", ")
				}
				fmt.Printf("st.Get(%d)", j)
			}
			fmt.Println(")")
			fmt.Println("    },")
			fmt.Printf("    Args: %d,\n", mt.NumIn())
			fmt.Println("    IsPure:true,")
			fmt.Println("  },")
		}
	}
	fmt.Print("}\n\n\n")
}
*/
