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
	// IsCommutative is true if the operation is commutative
	IsCommutative bool
}

type ToBool[V any] func(c V) bool

// UnaryOperator defines a operator like - or !
type UnaryOperator[V any] struct {
	// Operator is the operator as a string like "+"
	Operator string
	// Impl is the implementation of the operation
	Impl func(a V) V
}

// MethodHandler is used to give access to methods.
type MethodHandler[V any] interface {
	// GetMethod is used to get a method on a value.
	// The value is the first argument at calling the function.
	GetMethod(value V, methodName string) (Function[V], bool)
}

type MethodHandlerFunc[V any] func(value V, methodName string) (Function[V], bool)

func (mh MethodHandlerFunc[V]) GetMethod(value V, methodName string) (Function[V], bool) {
	return mh(value, methodName)
}

// ClosureHandler is used to convert closures
type ClosureHandler[V any] interface {
	// FromClosure is used to convert a closure to a value
	FromClosure(c Function[V]) V
	// ToClosure is used to convert a value to a closure
	// It returns the closure and a bool which is true if the value was a closure
	ToClosure(c V) (Function[V], bool)
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
	stringHandler   StringConverter[V]
	methodHandler   MethodHandler[V]
	closureHandler  ClosureHandler[V]
	listHandler     ListHandler[V]
	mapHandler      MapHandler[V]
	optimizer       Optimizer
	toBool          ToBool[V]
	staticFunctions map[string]Function[V]
	constants       map[string]V
	opMap           map[string]Operator[V]
	uMap            map[string]UnaryOperator[V]
	customGenerator Generator[V]
}

// New creates a new FunctionGenerator
func New[V any]() *FunctionGenerator[V] {
	g := &FunctionGenerator[V]{
		staticFunctions: map[string]Function[V]{},
		constants:       map[string]V{},
		methodHandler:   MethodHandlerFunc[V](methodByReflection[V]),
	}
	g.optimizer = NewOptimizer(g)
	return g
}

func (g *FunctionGenerator[V]) getParser() *Parser[V] {
	if g.parser == nil {
		parser := NewParser[V]().
			SetNumberParser(g.numberParser).
			SetStringConverter(g.stringHandler)

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

func (g *FunctionGenerator[V]) SetMethodHandler(methodHandler MethodHandler[V]) *FunctionGenerator[V] {
	g.methodHandler = methodHandler
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

func (g *FunctionGenerator[V]) SetStringConverter(stringConverter StringConverter[V]) *FunctionGenerator[V] {
	if g.parser != nil {
		panic("parser already created")
	}
	g.stringHandler = stringConverter
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

func (g *FunctionGenerator[V]) SetToBool(toBool ToBool[V]) *FunctionGenerator[V] {
	g.toBool = toBool
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
	g.operators = append(g.operators, Operator[V]{
		Operator:      operator,
		Impl:          impl,
		IsPure:        isPure,
		IsCommutative: isCommutative,
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

// Variables interface is used to access named variables
type Variables[V any] interface {
	// Get returns the value of the given name
	// False is returned if the value does not exist.
	Get(string) (V, bool)
}

type VarMap[V any] map[string]V

func (v VarMap[V]) Get(k string) (V, bool) {
	if va, ok := v[k]; ok {
		return va, true
	} else {
		return va, false
	}
}

type addVar[V any] struct {
	name   string
	val    V
	parent Variables[V]
}

func (a addVar[V]) Get(s string) (V, bool) {
	if s == a.name {
		return a.val, true
	}
	return a.parent.Get(s)
}

// AddVars is used to add values to a Variables[V] instance.
type AddVars[V any] struct {
	Vars   map[string]V
	Parent Variables[V]
}

func (a AddVars[V]) Get(s string) (V, bool) {
	if v, ok := a.Vars[s]; ok {
		return v, true
	}
	return a.Parent.Get(s)
}

// Func is a function which represents a node of the AST
type Func[V any] func(v Variables[V]) V

// Generator is used to define a customized generation of functions
type Generator[V any] interface {
	Generate(AST, *FunctionGenerator[V]) (Func[V], error)
}

// Generate takes a string and returns the function representing the expression given in
// the string. This is the main entry point of the parser.
func (g *FunctionGenerator[V]) Generate(exp string) (c func(Variables[V]) (V, error), errE error) {
	defer func() {
		rec := recover()
		if rec != nil {
			errE = EnhanceErrorf(rec, "error generating code")
			c = nil
		}
	}()

	ast, err := g.CreateAst(exp)
	if err != nil {
		return nil, err
	}

	expFunc, err := g.GenerateFunc(ast)
	if err != nil {
		return nil, EnhanceErrorf(err, "error generating code")
	}
	return func(v Variables[V]) (res V, e error) {
		defer func() {
			rec := recover()
			if rec != nil {
				e = EnhanceErrorf(rec, "error evaluating expression")
			}
		}()
		return expFunc(v), nil
	}, nil
}

// CreateAst uses the parser to create the abstract syntax tree.
// This method is public manly to inspect the AST in tests that live outside
// this package.
func (g *FunctionGenerator[V]) CreateAst(exp string) (AST, error) {
	ast, err := g.getParser().Parse(exp)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}

	if g.optimizer != nil {
		ast, err = Optimize(ast, g.optimizer)
		if err != nil {
			return nil, err
		}
	}
	return ast, nil
}

// Function represents a function
type Function[V any] struct {
	// Func is the function itself
	Func func(a []V) V
	// Args gives the number of arguments required. It is used for checking
	// the number of arguments in the call. The value -1 means any number of
	// arguments is allowed
	Args int
	// IsPure is true if this is a pure function
	IsPure bool
}

// Eval makes calling of functions a bit easier.
func (f *Function[V]) Eval(a ...V) V {
	return f.Func(a)
}

// GenerateFunc creates a Func[V] function.
// This method is public to allow its usage in the custom code generator.
func (g *FunctionGenerator[V]) GenerateFunc(ast AST) (Func[V], error) {
	if g.customGenerator != nil {
		c, err := g.customGenerator.Generate(ast, g)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return c, nil
		}
	}
	switch a := ast.(type) {
	case *Ident:
		return func(v Variables[V]) V {
			if va, ok := v.Get(a.Name); ok {
				return va
			} else {
				panic(a.Errorf("variable '%v' not found", a.Name))
			}
		}, nil
	case *Const[V]:
		n := a.Value
		return func(v Variables[V]) V {
			return n
		}, nil
	case *Let:
		if c, ok := a.Value.(*ClosureLiteral); ok {
			// if "let" is used to store a closure, allow recursion
			closureFunc, err := g.GenerateFunc(c.Func)
			if err != nil {
				return nil, err
			}
			valFunc := func(v Variables[V]) V {
				funcAdded := addVar[V]{name: a.Name, parent: v}
				theClosure := g.closureHandler.FromClosure(Function[V]{
					Func: func(args []V) V {
						vm := map[string]V{}
						for i, n := range c.Names {
							vm[n] = args[i]
						}
						return closureFunc(AddVars[V]{
							Vars:   vm,
							Parent: funcAdded,
						})
					},
					Args: len(c.Names),
				})
				funcAdded.val = theClosure
				return theClosure
			}
			mainFunc, err := g.GenerateFunc(a.Inner)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				va := valFunc(v)
				return mainFunc(addVar[V]{name: a.Name, val: va, parent: v})
			}, nil
		} else {
			// simple non closure let
			valFunc, err := g.GenerateFunc(a.Value)
			if err != nil {
				return nil, err
			}
			mainFunc, err := g.GenerateFunc(a.Inner)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				va := valFunc(v)
				return mainFunc(addVar[V]{name: a.Name, val: va, parent: v})
			}, nil
		}
	case *If:
		condFunc, err := g.GenerateFunc(a.Cond)
		if err != nil {
			return nil, err
		}
		thenFunc, err := g.GenerateFunc(a.Then)
		if err != nil {
			return nil, err
		}
		elseFunc, err := g.GenerateFunc(a.Else)
		if err != nil {
			return nil, err
		}
		if g.toBool != nil {
			return func(v Variables[V]) V {
				if g.toBool(condFunc(v)) {
					return thenFunc(v)
				} else {
					return elseFunc(v)
				}
			}, nil
		}
	case *Unary:
		valFunc, err := g.GenerateFunc(a.Value)
		if err != nil {
			return nil, err
		}
		op := g.uMap[a.Operator].Impl
		return func(v Variables[V]) V {
			return op(valFunc(v))
		}, nil
	case *Operate:
		aFunc, err := g.GenerateFunc(a.A)
		if err != nil {
			return nil, err
		}
		bFunc, err := g.GenerateFunc(a.B)
		if err != nil {
			return nil, err
		}
		op := g.opMap[a.Operator].Impl
		return func(v Variables[V]) V {
			return op(aFunc(v), bFunc(v))
		}, nil
	case *ClosureLiteral:
		closureFunc, err := g.GenerateFunc(a.Func)
		if err != nil {
			return nil, err
		}
		return func(v Variables[V]) V {
			return g.closureHandler.FromClosure(Function[V]{
				Func: func(args []V) V {
					vm := map[string]V{}
					for i, n := range a.Names {
						vm[n] = args[i]
					}
					return closureFunc(AddVars[V]{
						Vars:   vm,
						Parent: v,
					})
				},
				Args: len(a.Names),
			})
		}, nil
	case *ListLiteral:
		if g.listHandler != nil {
			itemFuncs, err := g.genFuncList(a.List)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				itemValues := make([]V, len(itemFuncs))
				for i, itemFunc := range itemFuncs {
					itemValues[i] = itemFunc(v)
				}
				return g.listHandler.FromList(itemValues)
			}, nil
		}
	case *ListAccess:
		if g.listHandler != nil {
			indexFunc, err := g.GenerateFunc(a.Index)
			if err != nil {
				return nil, err
			}
			listFunc, err := g.GenerateFunc(a.List)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				i := indexFunc(v)
				l := listFunc(v)
				if v, err := g.listHandler.AccessList(l, i); err == nil {
					return v
				} else {
					panic(a.EnhanceErrorf(err, "List error"))
				}
			}, nil
		}
	case *MapLiteral:
		if g.mapHandler != nil {
			itemsCode, err := g.genCodeMap(a.Map)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				mapValues := map[string]V{}
				for i, arg := range itemsCode {
					mapValues[i] = arg(v)
				}
				return g.mapHandler.FromMap(mapValues)
			}, nil
		}
	case *MapAccess:
		if g.mapHandler != nil {
			mapFunc, err := g.GenerateFunc(a.MapValue)
			if err != nil {
				return nil, err
			}
			return func(v Variables[V]) V {
				l := mapFunc(v)
				if v, err := g.mapHandler.AccessMap(l, a.Key); err == nil {
					return v
				} else {
					panic(a.EnhanceErrorf(err, "Map error"))
				}
			}, nil
		}
	case *FunctionCall:
		if id, ok := a.Func.(*Ident); ok {
			if fun, ok := g.staticFunctions[id.Name]; ok {
				if fun.Args >= 0 && fun.Args != len(a.Args) {
					return nil, id.Errorf("wrong number of arguments at call of %s, required %d, found %d", id.Name, fun.Args, len(a.Args))
				}
				argsFuncList, err := g.genFuncList(a.Args)
				if err != nil {
					return nil, err
				}
				return func(v Variables[V]) V {
					return fun.Func(evalList(argsFuncList, v))
				}, nil
			}
		}
		funcFunc, err := g.GenerateFunc(a.Func)
		if err != nil {
			return nil, err
		}
		argsFuncList, err := g.genFuncList(a.Args)
		if err != nil {
			return nil, err
		}
		return func(v Variables[V]) V {
			theFunc, ok := g.extractFunction(funcFunc(v))
			if !ok {
				panic(a.Errorf("not a function: %v", a.Func))
			}
			if theFunc.Args >= 0 && theFunc.Args != len(a.Args) {
				panic(a.Errorf("wrong number of arguments at call of %v, required %d, found %d", a.Func, theFunc.Args, len(a.Args)))
			}
			return theFunc.Func(evalList(argsFuncList, v))
		}, nil
	case *MethodCall:
		valFunc, err := g.GenerateFunc(a.Value)
		if err != nil {
			return nil, err
		}
		name := a.Name
		argsFuncList, err := g.genFuncList(a.Args)
		if err != nil {
			return nil, err
		}
		return func(v Variables[V]) V {
			value := valFunc(v)
			// name could be a method, but it could also be the name of a field which stores a closure
			// If it is a closure field, this should be a map access!
			if g.mapHandler != nil && g.mapHandler.IsMap(value) {
				if va, err := g.mapHandler.AccessMap(value, name); err == nil {
					if theFunc, ok := g.extractFunction(va); ok {
						return theFunc.Func(evalList(argsFuncList, v))
					}
				}
			}

			if g.methodHandler != nil {
				if me, ok := g.methodHandler.GetMethod(value, name); ok {
					if me.Args != len(argsFuncList)+1 {
						panic(a.Errorf("wrong number of arguments at call of %s, required %d, found %d", name, me.Args-1, len(argsFuncList)))
					}
					argsValues := make([]V, len(argsFuncList)+1)
					argsValues[0] = value
					for i, arg := range argsFuncList {
						argsValues[i+1] = arg(v)
					}
					return me.Func(argsValues)
				}
			}
			panic(a.Errorf("method %s not found", name))
		}, nil
	}
	return nil, ast.GetLine().Errorf("not supported: %v", ast)
}

func (g *FunctionGenerator[V]) genFuncList(a []AST) ([]Func[V], error) {
	args := make([]Func[V], len(a))
	for i, arg := range a {
		var err error
		args[i], err = g.GenerateFunc(arg)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

func (g *FunctionGenerator[V]) genCodeMap(a map[string]AST) (map[string]Func[V], error) {
	args := map[string]Func[V]{}
	for i, arg := range a {
		var err error
		args[i], err = g.GenerateFunc(arg)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

func evalList[V any](argsCode []Func[V], v Variables[V]) []V {
	argsValues := make([]V, len(argsCode))
	for i, arg := range argsCode {
		argsValues[i] = arg(v)
	}
	return argsValues
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

func methodByReflection[V any](value V, name string) (Function[V], bool) {
	name = firstRuneUpper(name)
	typeOf := reflect.TypeOf(value)
	if m, ok := typeOf.MethodByName(name); ok {
		return Function[V]{
			Func: func(args []V) V {
				argsValues := make([]reflect.Value, len(args))
				argsValues[0] = reflect.ValueOf(value)
				for i, arg := range args {
					argsValues[i] = reflect.ValueOf(arg)
				}

				res := m.Func.Call(argsValues)
				if len(res) == 1 {
					if v, ok := res[0].Interface().(V); ok {
						return v
					} else {
						panic(fmt.Errorf("result of method %s is not a value. It is: %v", name, res[0]))
					}
				} else {
					panic(fmt.Errorf("method %s does not return a single value but %v values", name, len(res)))
				}
			},
			Args:   m.Type.NumIn(),
			IsPure: false,
		}, true
	} else {
		var buf bytes.Buffer
		var v *V
		typeOfValue := reflect.TypeOf(v).Elem()
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
		panic(fmt.Errorf("method %s not found on %v, available are: %v", name, typeOf, buf.String()))
	}
}

func firstRuneUpper(name string) string {
	r, l := utf8.DecodeRune([]byte(name))
	if unicode.IsUpper(r) {
		return name
	}
	return string(unicode.ToUpper(r)) + name[l:]
}
