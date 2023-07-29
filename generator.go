package parser2

import (
	"bytes"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

// Operator defines a operator like +
type Operator[V any] struct {
	// Operator is the operator as a string like "+"
	Operator string
	// Impl is the implementation of the operation
	Impl func(a, b V) V
	// IsPure is true if the result of the operation depends only on the operands.
	// This is usually the case, there are only special corner cases where it is not.
	// So IsPure is usually true.
	IsPure bool
}

// UnaryOperator defines a operator like - or !
type UnaryOperator[V any] struct {
	// Operator is the operator as a string like "+"
	Operator string
	// Impl is the implementation of the operation
	Impl func(a V) V
}

// ClosureHandler is used to convert closures
type ClosureHandler[V any] interface {
	// FromClosure is used to convert a closure to a value
	FromClosure(c Closure[V]) V
	// ToClosure is used to convert a value to a closure
	// It returns the closure and a bool which is true if the value was a closure
	ToClosure(c V) (Closure[V], bool)
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

// FunctionGenerator is used to create a closure based implementation of
// the given expression. The type parameter gives the type the parser works on.
type FunctionGenerator[V any] struct {
	parser          *Parser[V]
	operators       []Operator[V]
	unary           []UnaryOperator[V]
	numberParser    NumberParser[V]
	stringHandler   StringHandler[V]
	closureHandler  ClosureHandler[V]
	listHandler     ListHandler[V]
	mapHandler      MapHandler[V]
	optimizer       Optimizer
	staticFunctions map[string]Function[V]
	constants       map[string]V
	opMap           map[string]Operator[V]
	uMap            map[string]UnaryOperator[V]
	customGenerator Generator[V]
	typeOfValue     reflect.Type
}

// New creates a new FunctionGenerator
func New[V any]() *FunctionGenerator[V] {
	var v *V
	g := &FunctionGenerator[V]{
		staticFunctions: map[string]Function[V]{},
		constants:       map[string]V{},
		typeOfValue:     reflect.TypeOf(v).Elem(),
	}
	g.optimizer = NewOptimizer(g)
	return g
}

func (g *FunctionGenerator[V]) getParser() *Parser[V] {
	if g.parser == nil {
		parser := NewParser[V]().
			SetNumberParser(g.numberParser).
			SetStringHandler(g.stringHandler)

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

func (g *FunctionGenerator[V]) SetNumberParser(numberParser NumberParser[V]) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
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
	if g.parser != nil {
		panic("parser already created")
	}
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

func (g *FunctionGenerator[V]) AddSimpleFunction(name string, f func(V) V) *FunctionGenerator[V] {
	return g.AddStaticFunction(name, Function[V]{
		Func:   func(a []V) V { return f(a[0]) },
		Args:   1,
		IsPure: true,
	})
}

func (g *FunctionGenerator[V]) AddStaticFunction(n string, f Function[V]) *FunctionGenerator[V] {
	g.staticFunctions[n] = f
	return g
}

func (g *FunctionGenerator[V]) AddConstant(n string, c V) *FunctionGenerator[V] {
	g.constants[n] = c
	return g
}

// AddOp adds an operation to the generator.
// The Operation needs to be pure.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOp(operator string, impl func(a V, b V) V) *FunctionGenerator[V] {
	return g.AddOpPure(operator, impl, true)
}

// AddOpPure adds an operation to the generator.
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (g *FunctionGenerator[V]) AddOpPure(operator string, impl func(a V, b V) V, isPure bool) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.operators = append(g.operators, Operator[V]{
		Operator: operator,
		Impl:     impl,
		IsPure:   isPure,
	})
	return g
}

func (g *FunctionGenerator[V]) AddUnary(operator string, impl func(a V) V) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.unary = append(g.unary, UnaryOperator[V]{
		Operator: operator,
		Impl:     impl,
	})
	return g
}

func (g *FunctionGenerator[V]) ModifyParser(modify func(a *Parser[V])) *FunctionGenerator[V] {
	modify(g.getParser())
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
	ast, err := g.getParser().Parse(exp)
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

func (f *Function[V]) Eval(a ...V) V {
	return f.Func(a)
}

type Closure[V any] struct {
	Names   []string
	Func    Code[V]
	Context Vars[V]
}

func (c *Closure[V]) CreateFunction() Function[V] {
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
	}
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
	case *Unary:
		c := g.Gen(a.Value)
		op := g.uMap[a.Operator].Impl
		return func(v Vars[V]) V {
			return op(c(v))
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
			if fun, ok := g.staticFunctions[string(id)]; ok {
				if fun.Args >= 0 && fun.Args != len(a.Args) {
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
			if theFunc.Args >= 0 && theFunc.Args != len(a.Args) {
				panic(fmt.Errorf("wrong number of arguments in call to %v", a.Func))
			}
			return theFunc.Func(evalList(argsCode, v))
		}
	case *MethodCall:
		valueCode := g.Gen(a.Value)
		name := a.Name
		argsCode := g.genList(a.Args)
		tov := g.typeOfValue
		return func(v Vars[V]) V {
			value := valueCode(v)
			// name could be a method, but it could also be the name of a field which stores a closure
			// If it is a closure field, this should be a map access!
			if g.mapHandler != nil && g.mapHandler.IsMap(value) {
				if va, err := g.mapHandler.AccessMap(value, name); err == nil {
					if theFunc, ok := g.extractFunction(va); ok {
						return theFunc.Func(evalList(argsCode, v))
					}
				}
			}
			argsValues := make([]reflect.Value, len(argsCode)+1)
			argsValues[0] = reflect.ValueOf(value)
			for i, arg := range argsCode {
				argsValues[i+1] = reflect.ValueOf(arg(v))
			}
			return callMethod(value, name, argsValues, tov)
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
			return c.CreateFunction(), true
		}
	}
	return Function[V]{}, false
}

func callMethod[V any](value V, name string, args []reflect.Value, typeOfValue reflect.Type) V {
	name = firstRuneUpper(name)
	typeOf := reflect.TypeOf(value)
	if m, ok := typeOf.MethodByName(name); ok {
		res := m.Func.Call(args)
		if len(res) == 1 {
			if v, ok := res[0].Interface().(V); ok {
				return v
			} else {
				panic(fmt.Errorf("result of method %v is not a value. It is: %v", name, res[0]))
			}
		} else {
			panic(fmt.Errorf("method %v does not return a single value: %v", name, len(res)))
		}
	} else {
		var buf bytes.Buffer
	outer:
		for i := 0; i < typeOf.NumMethod(); i++ {
			m := typeOf.Method(i)
			mt := m.Type
			if mt.NumOut() == 1 {
				if mt.Out(0).Implements(typeOfValue) {
					for i := 0; i < mt.NumIn(); i++ {
						if !mt.In(i).Implements(typeOfValue) {
							continue outer
						}
					}
					if buf.Len() > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(m.Name)
					buf.WriteString("(")
					for i := 1; i < mt.NumIn(); i++ {
						if i > 1 {
							buf.WriteString(", ")
						}
						buf.WriteString(mt.In(i).Name())
					}
					buf.WriteString(")")
				}
			}
		}
		panic(fmt.Errorf("method %v not found, available are: "+buf.String(), name))
	}
}

func firstRuneUpper(name string) string {
	r, l := utf8.DecodeRune([]byte(name))
	if unicode.IsUpper(r) {
		return name
	}
	return string(unicode.ToUpper(r)) + name[l:]
}
