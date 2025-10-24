package value

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/listMap"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Type int

const maxTypeId = 30

var (
	nilTypeId     Type
	IntTypeId     Type
	FloatTypeId   Type
	StringTypeId  Type
	BoolTypeId    Type
	ListTypeId    Type
	MapTypeId     Type
	ClosureTypeId Type
	FormatTypeId  Type
	LinkTypeId    Type
	FileTypeId    Type
)

type Value interface {
	ToList() (*List, bool)
	ToMap() (Map, bool)
	ToFloat() (float64, bool)
	ToString(st funcGen.Stack[Value]) (string, error)
	GetType() Type
}

func MethodAtType[V Value](args int, method func(obj V, stack funcGen.Stack[Value]) (Value, error)) funcGen.Function[Value] {
	return funcGen.Function[Value]{Func: func(stack funcGen.Stack[Value], closureStore []Value) (Value, error) {
		if obj, ok := stack.Get(0).(V); ok {
			return method(obj, stack)
		}
		return nil, fmt.Errorf("internal error: call of method on wrong type")
	}, Args: args + 1, IsPure: true}
}

type MethodMap map[string]funcGen.Function[Value]

func (mm MethodMap) Get(name string) (funcGen.Function[Value], error) {
	if m, ok := mm[name]; ok {
		return m, nil
	}

	type fes struct {
		name string
		fu   funcGen.Function[Value]
	}
	var l []fes
	for k, f := range mm {
		l = append(l, fes{name: k, fu: f})
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].name < l[j].name
	})
	var b bytes.Buffer
	for _, fe := range l {
		b.WriteRune('\n')
		fe.fu.Description.WriteTo(&b, fe.name)
	}
	documentation := b.String()

	return funcGen.Function[Value]{}, parser2.NewNotFoundError(name, fmt.Errorf("method '%s' not found; available are:\n%s", name, documentation))
}

func (mm MethodMap) add(more MethodMap) {
	for k, m := range more {
		mm[k] = m
	}
}

type Closure funcGen.Function[Value]

func (c Closure) ToList() (*List, bool) {
	return nil, false
}

func (c Closure) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (c Closure) ToFloat() (float64, bool) {
	return 0, false
}

func (c Closure) ToString(funcGen.Stack[Value]) (string, error) {
	return "<function>", nil
}

func (c Closure) Eval(st funcGen.Stack[Value], a Value) (Value, error) {
	return funcGen.Function[Value](c).Eval(st, a)
}

func (c Closure) EvalSt(st funcGen.Stack[Value], a ...Value) (Value, error) {
	return funcGen.Function[Value](c).EvalSt(st, a...)
}

func createClosureMethods() MethodMap {
	return MethodMap{
		"args": MethodAtType(0, func(c Closure, stack funcGen.Stack[Value]) (Value, error) { return Int(c.Args), nil }).
			SetMethodDescription("Returns the number of arguments the function takes."),
		"invoke": MethodAtType(1, func(c Closure, stack funcGen.Stack[Value]) (Value, error) {
			if l, ok := stack.Get(1).ToList(); ok {
				args, err := l.ToSlice(stack)
				if err != nil {
					return nil, err
				}
				if len(args) != c.Args {
					return nil, fmt.Errorf("wrong number of arguments in invoke: %d instead of %d", len(args), c.Args)
				}
				for _, arg := range args {
					stack.Push(arg)
				}
				return c.Func(stack.CreateFrame(len(args)), nil)
			} else {
				return nil, fmt.Errorf("argument of invike needs to be a list, not: %s", TypeName(stack.Get(1)))
			}
		}).
			SetMethodDescription("arg_list", "Invokes the function. The values of the given list are passed to the function as arguments."),
	}
}

func (c Closure) GetType() Type {
	return ClosureTypeId
}

type Bool bool

func (b Bool) ToList() (*List, bool) {
	return nil, false
}

func (b Bool) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (b Bool) ToFloat() (float64, bool) {
	return 0, false
}

func (b Bool) ToString(funcGen.Stack[Value]) (string, error) {
	if b {
		return "true", nil
	}
	return "false", nil
}

func createBoolMethods() MethodMap {
	return MethodMap{
		"string": MethodAtType(0, func(b Bool, stack funcGen.Stack[Value]) (Value, error) {
			s, err := b.ToString(stack)
			return String(s), err
		}).
			SetMethodDescription("Returns the string 'true' or 'false'."),
	}
}

func (b Bool) GetType() Type {
	return BoolTypeId
}

type Float float64

func (f Float) ToList() (*List, bool) {
	return nil, false
}

func (f Float) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (f Float) ToString(funcGen.Stack[Value]) (string, error) {
	return strconv.FormatFloat(float64(f), 'g', -1, 64), nil
}

func createFloatMethods() MethodMap {
	return MethodMap{
		"string": MethodAtType(0, func(f Float, stack funcGen.Stack[Value]) (Value, error) {
			s, err := f.ToString(stack)
			return String(s), err
		}).
			SetMethodDescription("Returns a string representation of the float."),
	}
}

func (f Float) GetType() Type {
	return FloatTypeId
}

func (f Float) ToFloat() (float64, bool) {
	return float64(f), true
}

type Int int

func (i Int) ToList() (*List, bool) {
	return nil, false
}

func (i Int) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (i Int) ToString(funcGen.Stack[Value]) (string, error) {
	return strconv.Itoa(int(i)), nil
}

func createIntMethods() MethodMap {
	return MethodMap{
		"string": MethodAtType(0, func(i Int, stack funcGen.Stack[Value]) (Value, error) {
			s, err := i.ToString(stack)
			return String(s), err
		}).
			SetMethodDescription("Returns a string representation of the int."),
	}
}

func (i Int) GetType() Type {
	return IntTypeId
}

func (i Int) ToFloat() (float64, bool) {
	return float64(i), true
}

type SimpleUnary struct {
	list []funcGen.UnaryOperatorFunc[Value]
	fg   *FunctionGenerator
	op   string
}

func NewUnaryOperationList(fg *FunctionGenerator, op string) *SimpleUnary {
	return &SimpleUnary{fg: fg, op: op}
}

func (su *SimpleUnary) Calc(v Value) (Value, error) {
	aType := v.GetType()
	if aType < Type(len(su.list)) {
		un := su.list[aType]
		if un != nil {
			return un(v)
		}
	}
	return nil, errors.New("unary operation '" + su.op + "' not defined on " + su.fg.typeNames[aType])
}

func (su *SimpleUnary) Register(a Type, op funcGen.UnaryOperatorFunc[Value]) {
	for a >= Type(len(su.list)) {
		su.list = append(su.list, nil)
	}
	if su.list[a] != nil {
		panic("unary operation '" + su.op + "' is already registered on this types")
	}
	su.list[a] = op
}

type OperationMatrix interface {
	funcGen.OperatorImpl[Value]
	Register(a, b Type, op funcGen.OperatorFunc[Value])
}

type operationMatrixSimple struct {
	matrix [][]funcGen.OperatorImpl[Value]
	fg     *FunctionGenerator
	op     string
}

func NewOperationMatrix(fg *FunctionGenerator, op string) OperationMatrix {
	return &operationMatrixSimple{fg: fg, op: op}
}

func (o *operationMatrixSimple) Calc(st funcGen.Stack[Value], a, b Value) (Value, error) {
	aType := a.GetType()
	bType := b.GetType()
	if aType < Type(len(o.matrix)) {
		line := o.matrix[aType]
		if bType < Type(len(line)) {
			oi := line[bType]
			if oi != nil {
				return oi.Calc(st, a, b)
			}
		}
	}
	return nil, errors.New("operation '" + o.op + "' not defined on " + o.fg.typeNames[aType] + ", " + o.fg.typeNames[bType])
}

func (o *operationMatrixSimple) Register(a, b Type, op funcGen.OperatorFunc[Value]) {
	for a >= Type(len(o.matrix)) {
		o.matrix = append(o.matrix, []funcGen.OperatorImpl[Value]{})
	}
	for b >= Type(len(o.matrix[a])) {
		o.matrix[a] = append(o.matrix[a], nil)
	}
	if o.matrix[a][b] != nil {
		panic("operation '" + o.op + "' is already registered on this types")
	}
	o.matrix[a][b] = op
}
func (fg *FunctionGenerator) ParseNumber(n string) (Value, error) {
	i, err := strconv.Atoi(n)
	if err == nil {
		return Int(i), nil
	}
	fl, err := strconv.ParseFloat(n, 64)
	if err == nil {
		return Float(fl), nil
	}
	return nil, err
}

func (fg *FunctionGenerator) FromString(s string) Value {
	return String(s)
}

func (fg *FunctionGenerator) FromClosure(c funcGen.Function[Value]) Value {
	return Closure(c)
}

func (fg *FunctionGenerator) ToClosure(value Value) (funcGen.Function[Value], bool) {
	if cl, ok := value.(Closure); ok {
		return funcGen.Function[Value](cl), true
	}
	return funcGen.Function[Value]{}, false
}

func (fg *FunctionGenerator) FromMap(items listMap.ListMap[Value]) Value {
	return Map{m: items}
}

func (fg *FunctionGenerator) AccessMap(mapValue Value, key string) (Value, error) {
	if m, ok := mapValue.ToMap(); ok {
		if v, ok := m.Get(key); ok {
			return v, nil
		} else {
			return nil, parser2.NewNotFoundError(key, fmt.Errorf("key '%s' not found in map; available are: %s", key, m.keyListDescription()))
		}
	} else {
		return nil, fmt.Errorf("'.%s' not possible; %s is not a map", key, TypeName(mapValue))
	}
}

func TypeName(v Value) string {
	tName := reflect.TypeOf(v).String()
	pos := strings.LastIndex(tName, ".")
	if pos >= 0 {
		tName = tName[pos+1:]
	}
	return tName
}

func (fg *FunctionGenerator) IsMap(mapValue Value) bool {
	_, ok := mapValue.ToMap()
	return ok
}

func (fg *FunctionGenerator) FromList(items []Value) Value {
	return NewList(items...)
}

func (fg *FunctionGenerator) AccessList(list Value, index Value) (Value, error) {
	if l, ok := list.ToList(); ok {
		if i, ok := index.(Int); ok {
			if i < 0 {
				return nil, fmt.Errorf("negative list index")
			} else {
				size, err := l.Size(funcGen.NewEmptyStack[Value]())
				if err != nil {
					return nil, err
				}
				if int(i) >= size {
					return nil, fmt.Errorf("index out of bounds %d>=size(%d)", i, size)
				} else {
					return l.items[i], nil
				}
			}
		} else {
			return nil, fmt.Errorf("not an int: %s", TypeName(index))
		}
	} else {
		return nil, fmt.Errorf("not a list: %s", TypeName(list))
	}
}

func (fg *FunctionGenerator) GenerateCustom(ast parser2.AST, gc funcGen.GeneratorContext, g *funcGen.FunctionGenerator[Value]) (funcGen.ParserFunc[Value], error) {
	if tc, ok := ast.(*parser2.TryCatch); ok {
		if cl, ok := tc.Catch.(*parser2.ClosureLiteral); ok && len(cl.Names) == 1 {
			tryFunc, err := g.GenerateFunc(tc.Try, gc)
			if err != nil {
				return nil, tc.EnhanceErrorf(err, "error in try expression")
			}
			catchFunc, err := g.GenerateFunc(tc.Catch, gc)
			if err != nil {
				return nil, tc.EnhanceErrorf(err, "error in catch expression")
			}
			l := tc.GetLine()
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				tryVal, tryErr := tryFunc(st, cs)
				if tryErr == nil {
					return tryVal, nil
				}
				catchVal, err := catchFunc(st, cs)
				if err != nil {
					return nil, l.EnhanceErrorf(err, "error in getting catch function")
				}
				theFunc, ok := g.ExtractFunction(catchVal)
				if !ok || theFunc.Args != 1 {
					// impossible because condition is checked above
					return nil, l.Errorf("internal catch error")
				}
				st.Push(String(tryErr.Error()))
				return theFunc.Func(st.CreateFrame(1), cs)
			}, nil
		}
	}
	if op, ok := ast.(*parser2.Operate); ok {
		// AND and OR with short evaluation
		switch op.Operator {
		case "&":
			aFunc, err := g.GenerateFunc(op.A, gc)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, gc)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				aVal, err := aFunc(st, cs)
				if err != nil {
					return nil, err
				}
				if a, ok := aVal.(Bool); ok {
					if !a {
						return Bool(false), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.(Bool); ok {
							return Bool(b), nil
						} else {
							return nil, fmt.Errorf("not a bool: %s", TypeName(bVal))
						}
					}
				} else {
					return nil, fmt.Errorf("not a bool: %s", TypeName(aVal))
				}
			}, nil
		case "|":
			aFunc, err := g.GenerateFunc(op.A, gc)
			if err != nil {
				return nil, err
			}
			bFunc, err := g.GenerateFunc(op.B, gc)
			if err != nil {
				return nil, err
			}
			return func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				aVal, err := aFunc(st, cs)
				if err != nil {
					return nil, err
				}
				if a, ok := aVal.(Bool); ok {
					if a {
						return Bool(true), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.(Bool); ok {
							return Bool(b), nil
						} else {
							return nil, fmt.Errorf("not a bool: %s", TypeName(bVal))
						}
					}
				} else {
					return nil, fmt.Errorf("not a bool: %s", TypeName(aVal))
				}
			}, nil
		}
	}
	return nil, nil
}

func simpleOnlyFloatFunc(name string, f func(float64) float64) funcGen.Function[Value] {
	return funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			v := st.Get(0)
			if fl, ok := v.ToFloat(); ok {
				return Float(f(fl)), nil
			}
			return nil, fmt.Errorf("%s not alowed on %s", name, TypeName(v))
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("float", "The mathematical "+name+" function.")
}

func simpleOnlyFloatFuncCheck(name string, argValid func(float65 float64) bool, f func(float64) float64) funcGen.Function[Value] {
	return funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			v := st.Get(0)
			if fl, ok := v.ToFloat(); ok {
				if !argValid(fl) {
					return nil, fmt.Errorf("%s not allowed with argument %f", name, fl)
				}
				f2 := f(fl)
				if math.IsNaN(f2) {
					fmt.Println(name, fl, f2)
				}
				return Float(f2), nil
			}
			return nil, fmt.Errorf("%s not alowed on %s", name, TypeName(v))
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("float", "The mathematical "+name+" function.")
}

type FunctionGenerator struct {
	*funcGen.FunctionGenerator[Value]
	methods   [maxTypeId]MethodMap
	equal     funcGen.BoolFunc[Value]
	less      funcGen.BoolFunc[Value]
	typeNames [maxTypeId]string

	typeId Type
}

func (fg *FunctionGenerator) RegisterType(name string) Type {
	fg.typeId++
	if fg.typeId >= maxTypeId {
		panic("too many types")
	}
	fg.typeNames[fg.typeId] = name
	return fg.typeId
}

func (fg *FunctionGenerator) GetMethod(value Value, methodName string) (funcGen.Function[Value], error) {
	typ := value.GetType()
	methodMap := fg.methods[typ]
	if methodMap == nil {
		return funcGen.Function[Value]{}, fmt.Errorf("no methods for type %s", TypeName(value))
	}
	m, err := methodMap.Get(methodName)
	if err != nil {
		return funcGen.Function[Value]{}, err
	} else {
		return m, nil
	}
}

func (fg *FunctionGenerator) RegisterMethods(id Type, methods MethodMap) *FunctionGenerator {
	if int(id) >= len(fg.methods) {
		panic(fmt.Sprintf("id %d is too big", id))
	} else if id == 0 {
		panic(fmt.Sprintf("type not registered"))
	}
	if fg.methods[id] == nil {
		fg.methods[id] = methods
	} else {
		fg.methods[id].add(methods)
	}
	return fg
}

func (fg *FunctionGenerator) GetOpMatrix(op string) OperationMatrix {
	impl := fg.GetOpImpl(op)
	if impl == nil {
		return nil
	}
	if im, ok := impl.(OperationMatrix); ok {
		return im
	} else {
		return nil
	}
}

func (fg *FunctionGenerator) GetUnaryList(op string) *SimpleUnary {
	impl := fg.GetUnaryOpImpl(op)
	if impl == nil {
		return nil
	}
	if im, ok := impl.(*SimpleUnary); ok {
		return im
	} else {
		return nil
	}
}

func (fg *FunctionGenerator) Modify(f func(*FunctionGenerator)) *FunctionGenerator {
	f(fg)
	return fg
}

func (fg *FunctionGenerator) GetDocumentation() []funcGen.TypeDocumentation {
	var td []funcGen.TypeDocumentation
	for i := Type(1); i <= fg.typeId; i++ {
		methodMap := fg.methods[i]
		if len(methodMap) > 0 {
			td = append(td, funcGen.CreateTypeDocumentation(fg.typeNames[i], methodMap))
		}
	}
	td = append(td, fg.FunctionGenerator.GetStaticDocumentation())
	return td
}

func New() *FunctionGenerator {
	f := &FunctionGenerator{}
	nilTypeId = f.RegisterType("nil")
	IntTypeId = f.RegisterType("int")
	FloatTypeId = f.RegisterType("float")
	StringTypeId = f.RegisterType("string")
	BoolTypeId = f.RegisterType("bool")
	ListTypeId = f.RegisterType("list")
	MapTypeId = f.RegisterType("map")
	ClosureTypeId = f.RegisterType("closure")
	FormatTypeId = f.RegisterType("format")
	LinkTypeId = f.RegisterType("link")
	FileTypeId = f.RegisterType("file")

	fg := funcGen.New[Value]().
		AddConstant("pi", Float(math.Pi)).
		AddConstant("true", Bool(true)).
		AddConstant("false", Bool(false)).
		SetNumberParser(f).
		SetKeyWords("let", "func", "if", "then", "else", "func", "switch", "case", "default", "const", "try", "catch").
		SetListHandler(f).
		SetMapHandler(f).
		SetClosureHandler(f).
		SetMethodHandler(f).
		SetCustomGenerator(f).
		SetStringConverter(f).
		SetToBool(func(c Value) (bool, bool) {
			if b, ok := c.(Bool); ok {
				return bool(b), true
			}
			return false, false
		}).
		AddOpImpl("|", true, Or(f)).
		AddOpImpl("&", true, And(f))

	f.FunctionGenerator = fg
	equal := Equal(f)
	less := Less(f)

	fg.AddOpImpl("=", true, equal)
	fg.AddOp("!=", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		eq, err := equal.Calc(st, a, b)
		return !(eq.(Bool)), err
	})
	fg.AddOp("~", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		if list, ok := b.(*List); ok {
			if search, ok := a.(*List); ok {
				items, err := list.containsAllItems(st, search, f)
				return Bool(items), err
			} else {
				item, err := list.containsItem(st, a, f)
				return Bool(item), err
			}
		}
		if m, ok := b.(Map); ok {
			if key, ok := a.(String); ok {
				return m.ContainsKey(key), nil
			}
		}
		if strToLookFor, ok := a.(String); ok {
			if strToLookIn, ok := b.(String); ok {
				return Bool(strings.Contains(string(strToLookIn), string(strToLookFor))), nil
			}
		}
		return nil, notAllowed("~", a, b)
	})
	fg.AddOpImpl("<", false, less)
	fg.AddOp(">", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		return less.Calc(st, b, a)
	})
	fg.AddOp("<=", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		le, err := less.Calc(st, a, b)
		if err != nil {
			return nil, err
		}
		if le.(Bool) {
			return Bool(true), nil
		}
		return equal.Calc(st, a, b)
	})
	fg.AddOp(">=", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		le, err := less.Calc(st, b, a)
		if err != nil {
			return nil, err
		}
		if le.(Bool) {
			return Bool(true), nil
		}
		return equal.Calc(st, a, b)
	})

	fg.AddOpImpl("+", false, Add(f)).
		AddOpImpl("-", false, Sub(f)).
		AddOpImpl("<<", false, Left(f)).
		AddOpImpl(">>", false, Right(f)).
		AddOpImpl("*", true, Mul(f)).
		AddOpImpl("%", false, Mod(f)).
		AddOpImpl("/", false, Div(f)).
		AddOpImpl("^", false, Pow(f)).
		AddUnary("-", Neg(f)).
		AddUnary("!", Not(f)).
		AddStaticFunction("throw", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				if s, ok := st.Get(0).(String); ok {
					return nil, errors.New(string(s))
				} else {
					return nil, errors.New("throw needs a string as argument")
				}
			},
			Args:   1,
			IsPure: false,
		}.SetDescription("message", "Throws an exception.")).
		AddStaticFunction("string", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				s, err := st.Get(0).ToString(st)
				return String(s), err
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the string representation of the value.")).
		AddStaticFunction("isFloat", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				_, ok := v.(Float)
				return Bool(ok), nil
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns true if the value is a float.")).
		AddStaticFunction("isInt", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				_, ok := v.(Int)
				return Bool(ok), nil
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns true if the value is a int.")).
		AddStaticFunction("float", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if f, ok := v.ToFloat(); ok {
					return Float(f), nil
				}
				return nil, fmt.Errorf("float not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the float representation of the value.")).
		AddStaticFunction("int", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if i, ok := v.(Int); ok {
					return i, nil
				} else if f, ok := v.ToFloat(); ok {
					return Int(f), nil
				}
				return nil, fmt.Errorf("int not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the int representation of the value.")).
		AddStaticFunction("abs", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v < 0 {
						return -v, nil
					}
					return v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Float(math.Abs(f)), nil
				}
				return nil, fmt.Errorf("abs not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "If value is negative, returns -value. Otherwise returns the value unchanged.")).
		AddStaticFunction("sign", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					if v < 0 {
						return Int(-1), nil
					} else if v == 0 {
						return Int(0), nil
					}
					return Int(1), nil
				}
				if f, ok := v.ToFloat(); ok {
					if f < 0 {
						return Float(-1), nil
					} else if f == 0 {
						return Float(0), nil
					}
					return Float(1), nil
				}
				return nil, fmt.Errorf("abs not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "If value is negative -1 is returned. Otherwise 1 is returned.")).
		AddStaticFunction("sqr", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v * v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Float(f * f), nil
				}
				return nil, fmt.Errorf("sqr not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the square of the value.")).
		AddStaticFunction("random", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if n, ok := v.(Int); ok {
					return Int(rand.Intn(int(n))), nil
				}
				return nil, errors.New("random only allowed on int")
			},
			Args:   1,
			IsPure: false,
		}.SetDescription("n", "Returns a random integer between 0 and n-1.")).
		AddStaticFunction("round", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if v, ok := v.(Int); ok {
					return v, nil
				}
				if f, ok := v.ToFloat(); ok {
					return Int(math.Round(f)), nil
				}
				return nil, fmt.Errorf("sqr not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("value", "Returns the value rounded to the nearest integer.")).
		AddStaticFunction("binAnd", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				if a, ok := st.Get(0).(Int); ok {
					if b, ok := st.Get(1).(Int); ok {
						return Int(a & b), nil
					}
				}
				return nil, fmt.Errorf("binAnd not alowed on %s, %s", TypeName(st.Get(0)), TypeName(st.Get(1)))
			},
			Args:   2,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the binary and of a, b.")).
		AddStaticFunction("binOr", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				if a, ok := st.Get(0).(Int); ok {
					if b, ok := st.Get(1).(Int); ok {
						return Int(a | b), nil
					}
				}
				return nil, fmt.Errorf("binOr not alowed on %s, %s", TypeName(st.Get(0)), TypeName(st.Get(1)))
			},
			Args:   2,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the binary or of a, b.")).
		AddStaticFunction("bisection", funcGen.Function[Value]{
			Func:   bisectionValue,
			Args:   4,
			IsPure: true,
		}.SetDescription("func(float) float", "min", "max", "eps", "Searches a zero in the given function by using the bisection method.").VarArgs(3, 4)).
		AddStaticFunction("createLowPass", funcGen.Function[Value]{
			Func:   createLowPass,
			Args:   4,
			IsPure: true,
		}.SetDescription("name", "func(p) float", "func(p) float", "tau", "Returns a low pass filter creating signal [name].")).
		AddStaticFunction("list", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if size, ok := v.(Int); ok {
					return NewListFromSizedIterable(iterator.Generate[Value, funcGen.Stack[Value]](int(size), func(i int) (Value, error) { return Int(i), nil }), int(size)), nil
				}
				return nil, fmt.Errorf("list not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a list with n integer values, starting with 0.")).
		AddStaticFunction("goto", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if state, ok := v.(Int); ok {
					return createState(int(state)), nil
				}
				return nil, errors.New("goto requires an int")
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a map with the key 'state' set to the given value.")).
		AddStaticFunction("sprintf", funcGen.Function[Value]{Func: sprintf, Args: -1, IsPure: true}.
			SetDescription("format", "args", "The classic, well known sprintf function.")).
		AddStaticFunction("sqrt", simpleOnlyFloatFuncCheck("sqrt", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Sqrt(x) })).
		AddStaticFunction("ln", simpleOnlyFloatFuncCheck("ln", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Log(x) })).
		AddStaticFunction("log10", simpleOnlyFloatFuncCheck("log", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Log10(x) })).
		AddStaticFunction("trunc", simpleOnlyFloatFunc("trunc", func(x float64) float64 { return math.Trunc(x) })).
		AddStaticFunction("floor", simpleOnlyFloatFunc("floor", func(x float64) float64 { return math.Floor(x) })).
		AddStaticFunction("exp", simpleOnlyFloatFunc("exp", func(x float64) float64 { return math.Exp(x) })).
		AddStaticFunction("sin", simpleOnlyFloatFunc("sin", func(x float64) float64 { return math.Sin(x) })).
		AddStaticFunction("cos", simpleOnlyFloatFunc("cos", func(x float64) float64 { return math.Cos(x) })).
		AddStaticFunction("tan", simpleOnlyFloatFunc("tan", func(x float64) float64 { return math.Tan(x) })).
		AddStaticFunction("asin", simpleOnlyFloatFuncCheck("asin", func(arg float64) bool { return arg >= -1 && arg <= 1 }, func(x float64) float64 { return math.Asin(x) })).
		AddStaticFunction("acos", simpleOnlyFloatFuncCheck("acos", func(arg float64) bool { return arg >= -1 && arg <= 1 }, func(x float64) float64 { return math.Acos(x) })).
		AddStaticFunction("atan", simpleOnlyFloatFunc("atan", func(x float64) float64 { return math.Atan(x) }))

	f.RegisterMethods(ListTypeId, createListMethods(f))
	f.RegisterMethods(MapTypeId, createMapMethods())
	f.RegisterMethods(StringTypeId, createStringMethods())
	f.RegisterMethods(BoolTypeId, createBoolMethods())
	f.RegisterMethods(IntTypeId, createIntMethods())
	f.RegisterMethods(FloatTypeId, createFloatMethods())
	f.RegisterMethods(ClosureTypeId, createClosureMethods())

	f.AddStaticFunction("min", funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			var m Value
			for i := 0; i < st.Size(); i++ {
				v := st.Get(i)
				if i == 0 {
					m = v
				} else {
					le, err := less.Calc(st, v, m)
					if err != nil {
						return nil, err
					}
					if le.(Bool) {
						m = v
					}
				}
			}
			return m, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("a", "b", "Returns the smaller of a and b."))
	f.AddStaticFunction("max", funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			var m Value
			for i := 0; i < st.Size(); i++ {
				v := st.Get(i)
				if i == 0 {
					m = v
				} else {
					le, err := less.Calc(st, m, v)
					if err != nil {
						return nil, err
					}
					if le.(Bool) {
						m = v
					}
				}
			}
			return m, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("a", "b", "Returns the larger of a and b."))
	return f
}

func sprintf(st funcGen.Stack[Value], cs []Value) (Value, error) {
	switch st.Size() {
	case 0:
		return String(""), nil
	case 1:
		v := st.Get(0)
		if st, ok := v.(String); ok {
			return String(fmt.Sprint(string(st))), nil
		} else {
			return String(fmt.Sprint(v)), nil
		}
	default:
		if s, ok := st.Get(0).(String); ok {
			values := make([]any, st.Size()-1)
			for i := 1; i < st.Size(); i++ {
				v := st.Get(i)
				if st, ok := v.(String); ok {
					values[i-1] = string(st)
				} else {
					values[i-1] = v
				}
			}
			return String(fmt.Sprintf(string(s), values...)), nil
		} else {
			return nil, fmt.Errorf("sprintf requires string as first argument")
		}
	}
}

func createLowPass(st funcGen.Stack[Value], store []Value) (Value, error) {
	var name string
	if n, ok := st.Get(0).(String); ok {
		name = string(n)
	} else {
		return nil, fmt.Errorf("createLowPass requires a string as first argument")
	}
	t, err := ToFunc("createLowPass", st, 1, 1)
	if err != nil {
		return nil, err
	}
	xf, err := ToFunc("createLowPass", st, 2, 1)
	if err != nil {
		return nil, err
	}
	tau, err := ToFloat("createLowPass", st, 3)
	if err != nil {
		return nil, err
	}
	lp := Closure(funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			p0 := st.Get(0)
			p1 := st.Get(1)
			ol, _ := st.Get(2).ToMap()
			yv, _ := ol.Get(name)
			t0, err := MustFloat(t.Eval(st, p0))
			if err != nil {
				return nil, err
			}
			t1, err := MustFloat(t.Eval(st, p1))
			if err != nil {
				return nil, err
			}
			x, err := MustFloat(xf.Eval(st, p1))
			if err != nil {
				return nil, err
			}
			y, err := MustFloat(yv, nil)
			dt := t1 - t0
			a := math.Exp(-dt / tau)
			yn := y*a + x*(1-a)
			m, _ := p1.ToMap()
			return NewMap(AppendMap{key: name, value: Float(yn), parent: m}), nil
		},
		Args:   3,
		IsPure: true,
	})
	in := Closure(funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			p0 := st.Get(0)
			x, err := xf.Eval(st, p0)
			if err != nil {
				return nil, err
			}
			m, _ := p0.ToMap()
			return NewMap(AppendMap{key: name, value: x, parent: m}), nil
		},
		Args:   1,
		IsPure: true,
	})
	return NewMap(listMap.New[Value](2).Append("filter", lp).Append("initial", in)), nil
}

func bisectionValue(st funcGen.Stack[Value], _ []Value) (Value, error) {
	if f, err := ToFunc("bisection", st, 0, 1); err != nil {
		return nil, err
	} else {
		if xMin, err := ToFloat("bisection", st, 1); err != nil {
			return nil, err
		} else {
			if xMax, err := ToFloat("bisection", st, 2); err != nil {
				return nil, err
			} else {

				eps := 1e-10
				if e, ok := st.GetOptional(3, Float(1e-10)).ToFloat(); ok {
					eps = e
				}

				fu := func(x float64) (float64, error) {
					r, err := f.Eval(st, Float(x))
					if err != nil {
						return 0, err
					}
					if r, ok := r.ToFloat(); ok {
						return r, nil
					}
					return 0, fmt.Errorf("solve function must return a float, but returned %s", TypeName(r))
				}

				r, err := Bisection(fu, xMin, xMax, eps)
				return Float(r), err
			}
		}
	}
}

func Bisection(f func(float64) (float64, error), xMin, xMax, eps float64) (float64, error) {
	yMin, err := f(xMin)
	if err != nil {
		return 0, err
	}
	if math.Abs(yMin) < eps {
		return xMin, nil
	}

	yMax, err := f(xMax)
	if err != nil {
		return 0, err
	}
	if math.Abs(yMax) < eps {
		return xMax, nil
	}

	if (yMin < 0) == (yMax < 0) {
		return 0, fmt.Errorf("no zero in interval [%f,%f] for function", xMin, xMax)
	}

	n := 0
	for {
		xMid := (xMin + xMax) / 2
		yMid, err := f(xMid)
		if err != nil {
			return 0, err
		}

		if math.Abs(yMid) < eps {
			return xMid, nil
		}

		if (yMin < 0) == (yMid < 0) {
			xMin = xMid
			yMin = yMid
		} else {
			xMax = xMid
			yMax = yMid
		}

		n++
		if n > 1000 {
			return 0, fmt.Errorf("solve function did not converge in 1000 iterations for interval [%f,%f]", xMin, xMax)
		}
	}
}

func Must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

func MustFloat(v Value, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if f, ok := v.ToFloat(); ok {
		return f, nil
	}
	return 0, fmt.Errorf("not a float: %s", TypeName(v))
}

func MustInt(v Value, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	if i, ok := v.(Int); ok {
		return int(i), nil
	}
	return 0, fmt.Errorf("not an int: %s", TypeName(v))
}
