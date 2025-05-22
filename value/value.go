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

const (
	maxType = 30
)

var (
	nilTypeId     Type
	IntTypeId     Type
	FloatTypeId   Type
	StringTypeId  Type
	BoolTypeId    Type
	ListTypeId    Type
	MapTypeId     Type
	closureTypeId Type
	FormatTypeId  Type
	LinkTypeId    Type
	FileTypeId    Type
)

type Value interface {
	ToList() (*List, bool)
	ToMap() (Map, bool)
	ToInt() (int, bool)
	ToFloat() (float64, bool)
	ToString(st funcGen.Stack[Value]) (string, error)
	ToBool() (bool, bool)
	ToClosure() (funcGen.Function[Value], bool)
	GetType() Type
}

type OperatorTable struct {
	Op     string
	aTable [maxType][maxType]funcGen.OperatorImpl[Value]
}

func (ot *OperatorTable) Do(st funcGen.Stack[Value], a, b Value) (Value, error) {
	at := ot.aTable[a.GetType()]
	fu := at[b.GetType()]
	if fu == nil {
		return nil, fmt.Errorf("operator %s not available for types %s, %s", ot.Op, TypeName(a), TypeName(b))
	}
	do, err := fu.Do(st, a, b)
	if do == nil {
		return nil, fmt.Errorf("internal operator error: %s not available for types %s, %s", ot.Op, TypeName(a), TypeName(b))
	}
	return do, err
}

func (ot *OperatorTable) add(aT, bT Type, f funcGen.OperatorImpl[Value]) {
	ot.aTable[aT][bT] = f
}

func (ot *OperatorTable) addCommutative(aT, bT Type, f funcGen.OperatorImpl[Value]) {
	ot.add(aT, bT, f)
	ot.add(bT, aT, funcGen.OperatorImplFunc[Value](func(st funcGen.Stack[Value], a, b Value) (Value, error) {
		return f.Do(st, b, a)
	}))
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
	return funcGen.Function[Value]{}, parser2.NewNotFoundError(name, fmt.Errorf("method '%s' not found; available are:\n%s", name, b.String()))
}

func (mm MethodMap) add(more MethodMap) {
	for k, m := range more {
		mm[k] = m
	}
}

const NIL = nilType(0)

type nilType int

func (n nilType) ToList() (*List, bool) {
	return nil, false
}

func (n nilType) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (n nilType) ToInt() (int, bool) {
	return 0, false
}

func (n nilType) ToFloat() (float64, bool) {
	return 0, false
}

func (n nilType) ToString(funcGen.Stack[Value]) (string, error) {
	return "nil", nil
}

func (n nilType) String() string {
	return "nil"
}

func (n nilType) ToBool() (bool, bool) {
	return false, false
}

func (n nilType) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
}

func (n nilType) GetType() Type {
	return nilTypeId
}

type Closure funcGen.Function[Value]

func (c Closure) ToList() (*List, bool) {
	return nil, false
}

func (c Closure) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (c Closure) ToInt() (int, bool) {
	return 0, false
}

func (c Closure) ToFloat() (float64, bool) {
	return 0, false
}

func (c Closure) ToString(funcGen.Stack[Value]) (string, error) {
	return "<function>", nil
}

func (c Closure) ToBool() (bool, bool) {
	return false, false
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
	return closureTypeId
}

func (c Closure) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value](c), true
}

type Bool bool

func (b Bool) ToList() (*List, bool) {
	return nil, false
}

func (b Bool) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (b Bool) ToInt() (int, bool) {
	return 0, false
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

func (b Bool) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
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

func (b Bool) ToBool() (bool, bool) {
	return bool(b), true
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

func (f Float) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
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

func (f Float) ToBool() (bool, bool) {
	if f != 0 {
		return true, true
	}
	return false, true
}

func (f Float) ToInt() (int, bool) {
	return int(f), true
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

func (i Int) ToClosure() (funcGen.Function[Value], bool) {
	return funcGen.Function[Value]{}, false
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

func (i Int) ToBool() (bool, bool) {
	if i != 0 {
		return true, true
	}
	return false, true
}

func (i Int) ToInt() (int, bool) {
	return int(i), true
}

func (i Int) ToFloat() (float64, bool) {
	return float64(i), true
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
	return value.ToClosure()
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
		if i, ok := index.ToInt(); ok {
			if i < 0 {
				return nil, fmt.Errorf("negative list index")
			} else {
				size, err := l.Size(funcGen.NewEmptyStack[Value]())
				if err != nil {
					return nil, err
				}
				if i >= size {
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
				if a, ok := aVal.ToBool(); ok {
					if !a {
						return Bool(false), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.ToBool(); ok {
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
				if a, ok := aVal.ToBool(); ok {
					if a {
						return Bool(true), nil
					} else {
						bVal, err := bFunc(st, cs)
						if err != nil {
							return nil, err
						}
						if b, ok := bVal.ToBool(); ok {
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
	methods  [maxType]MethodMap
	opTables map[string]*OperatorTable
	equal    funcGen.BoolFunc[Value]
	less     funcGen.BoolFunc[Value]

	typeId Type
}

func (fg *FunctionGenerator) RegisterType() Type {
	fg.typeId++
	if fg.typeId >= maxType {
		panic(fmt.Sprintf("type id %d is too big", fg.typeId))
	}
	return fg.typeId
}

func (fg *FunctionGenerator) CreateOpTable(op string) *OperatorTable {
	if fg.opTables == nil {
		fg.opTables = make(map[string]*OperatorTable)
	}
	if _, ok := fg.opTables[op]; ok {
		panic(fmt.Sprintf("operator %s already registered", op))
	}
	ot := &OperatorTable{Op: op}
	fg.opTables[op] = ot
	return ot
}

func (fg *FunctionGenerator) GetOp(op string) funcGen.OperatorImpl[Value] {
	return fg.opTables[op]
}

type UpCast struct {
	From Type
	To   Type
	Cast func(Value) (Value, error)
}

func (fg *FunctionGenerator) AddUpCast(op string, uc ...UpCast) *FunctionGenerator {
	t := fg.opTables[op]
	for _, u := range uc {
		main := t.aTable[u.To][u.To]
		if main == nil {
			panic(fmt.Sprintf("operator %s not registered", op))
		}

		cast := u.Cast
		t.aTable[u.From][u.To] = funcGen.OperatorImplFunc[Value](func(st funcGen.Stack[Value], a, b Value) (Value, error) {
			if aa, err := cast(a); err != nil {
				return nil, err
			} else {
				return main.Do(st, aa, b)
			}
		})
		t.aTable[u.To][u.From] = funcGen.OperatorImplFunc[Value](func(st funcGen.Stack[Value], a, b Value) (Value, error) {
			if bb, err := cast(b); err != nil {
				return nil, err
			} else {
				return main.Do(st, a, bb)
			}
		})
	}

	return fg
}

func (fg *FunctionGenerator) AddTypedOpFunc(op string, a, b Type, impl funcGen.OperatorImplFunc[Value]) *FunctionGenerator {
	return fg.AddTypedOp(op, a, b, impl)
}

func (fg *FunctionGenerator) AddTypedOpCommFunc(op string, a, b Type, impl funcGen.OperatorImplFunc[Value]) *FunctionGenerator {
	return fg.AddTypedOpComm(op, a, b, impl)
}

func (fg *FunctionGenerator) AddTypedToAll(op string, a Type, impl funcGen.OperatorImpl[Value]) *FunctionGenerator {
	if fg.opTables == nil {
		panic("no operations registered")
	}
	if t, ok := fg.opTables[op]; ok {
		for i := 0; i < maxType; i++ {
			t.add(a, Type(i), impl)
		}
	} else {
		panic(fmt.Sprintf("operator %s not registered", op))
	}
	return fg
}

func (fg *FunctionGenerator) AddTypedOp(op string, a, b Type, impl funcGen.OperatorImpl[Value]) *FunctionGenerator {
	if fg.opTables == nil {
		panic("no operations registered")
	}
	if t, ok := fg.opTables[op]; ok {
		t.add(a, b, impl)
	} else {
		panic(fmt.Sprintf("operator %s not registered", op))
	}
	return fg
}

func (fg *FunctionGenerator) AddTypedOpComm(op string, a, b Type, impl funcGen.OperatorImpl[Value]) *FunctionGenerator {
	if fg.opTables == nil {
		panic("no operations registered")
	}
	if t, ok := fg.opTables[op]; ok {
		t.addCommutative(a, b, impl)
	} else {
		panic(fmt.Sprintf("operator %s not registered", op))
	}
	return fg
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
		panic("type not registered")
	}
	if fg.methods[id] == nil {
		fg.methods[id] = methods
	} else {
		fg.methods[id].add(methods)
	}
	return fg
}

func (fg *FunctionGenerator) Modify(f func(*FunctionGenerator)) *FunctionGenerator {
	f(fg)
	return fg
}

func New() *FunctionGenerator {
	f := &FunctionGenerator{}

	nilTypeId = f.RegisterType()
	IntTypeId = f.RegisterType()
	FloatTypeId = f.RegisterType()
	StringTypeId = f.RegisterType()
	BoolTypeId = f.RegisterType()
	ListTypeId = f.RegisterType()
	MapTypeId = f.RegisterType()
	closureTypeId = f.RegisterType()
	FormatTypeId = f.RegisterType()
	LinkTypeId = f.RegisterType()
	FileTypeId = f.RegisterType()

	fg := funcGen.New[Value]().
		AddConstant("nil", NIL).
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
		SetToBool(func(c Value) (bool, bool) { return c.ToBool() }).
		AddOp("|", true, Or).
		AddOp("&", true, And)

	equal := f.CreateOpTable("=")
	equalBool := func(st funcGen.Stack[Value], a Value, b Value) (bool, error) {
		eq, err := equal.Do(st, a, b)
		if err != nil {
			return false, err
		}
		if b, ok := eq.ToBool(); ok && b {
			return true, nil
		}
		return false, nil
	}
	f.equal = equalBool
	fg.SetIsEqual(equalBool)

	fg.AddOpPure("=", true, equal, true).
		AddOp("!=", true, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
			eq, err := equalBool(st, a, b)
			return Bool(!eq), err
		})
	fg.AddOp("~", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		if list, ok := b.(*List); ok {
			if search, ok := a.(*List); ok {
				items, err := list.containsAllItems(st, search, equalBool)
				return Bool(items), err
			} else {
				item, err := list.containsItem(st, a, equalBool)
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
	less := f.CreateOpTable("<")
	lessBool := func(st funcGen.Stack[Value], a Value, b Value) (bool, error) {
		eq, err := less.Do(st, a, b)
		if err != nil {
			return false, err
		}
		if b, ok := eq.ToBool(); ok && b {
			return true, nil
		}
		return false, nil
	}
	f.less = lessBool
	fg.AddOpPure("<", true, less, true)
	fg.AddOp(">", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		le, err := lessBool(st, b, a)
		return Bool(le), err
	})
	fg.AddOp("<=", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		le, err := lessBool(st, a, b)
		if err != nil {
			return nil, err
		}
		if le {
			return Bool(true), nil
		}
		eq, err := equalBool(st, a, b)
		return Bool(eq), err
	})
	fg.AddOp(">=", false, func(st funcGen.Stack[Value], a Value, b Value) (Value, error) {
		le, err := lessBool(st, b, a)
		if err != nil {
			return nil, err
		}
		if le {
			return Bool(true), nil
		}
		eq, err := equalBool(st, a, b)
		return Bool(eq), err
	})

	fg.AddOpPure("+", false, f.CreateOpTable("+"), true).
		AddOpPure("-", true, f.CreateOpTable("-"), true).
		AddOp("<<", false, Left).
		AddOp(">>", false, Right).
		AddOpPure("*", false, f.CreateOpTable("*"), true).
		AddOp("%", false, Mod).
		AddOpPure("/", false, f.CreateOpTable("/"), true).
		AddOpPure("^", false, f.CreateOpTable("^"), true).
		AddUnary("-", func(a Value) (Value, error) { return Neg(a) }).
		AddUnary("!", func(a Value) (Value, error) { return Not(a) }).
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
				if i, ok := v.ToInt(); ok {
					return Int(i), nil
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
				if n, ok := v.ToInt(); ok {
					return Int(rand.Intn(n)), nil
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
				if a, ok := st.Get(0).ToInt(); ok {
					if b, ok := st.Get(1).ToInt(); ok {
						return Int(a & b), nil
					}
				}
				return nil, fmt.Errorf("binAnd not alowed on %s, %s", TypeName(st.Get(0)), TypeName(st.Get(1)))
			},
			Args:   2,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the binary and of a, b")).
		AddStaticFunction("binOr", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				if a, ok := st.Get(0).ToInt(); ok {
					if b, ok := st.Get(1).ToInt(); ok {
						return Int(a | b), nil
					}
				}
				return nil, fmt.Errorf("binOr not alowed on %s, %s", TypeName(st.Get(0)), TypeName(st.Get(1)))
			},
			Args:   2,
			IsPure: true,
		}.SetDescription("a", "b", "Returns the binary or of a, b")).
		AddStaticFunction("createLowPass", funcGen.Function[Value]{
			Func:   createLowPass,
			Args:   4,
			IsPure: true,
		}.SetDescription("name", "func(p) float", "func(p) float", "tau", "Returns a low pass filter creating signal [name]")).
		AddStaticFunction("list", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if size, ok := v.ToInt(); ok {
					return NewListFromIterable(iterator.Generate[Value, funcGen.Stack[Value]](size, func(i int) (Value, error) { return Int(i), nil })), nil
				}
				return nil, fmt.Errorf("list not alowed on %s", TypeName(v))
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a list with n integer values, starting with 0.")).
		AddStaticFunction("goto", funcGen.Function[Value]{
			Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
				v := st.Get(0)
				if state, ok := v.ToInt(); ok {
					return createState(state), nil
				}
				return nil, errors.New("goto requires an int")
			},
			Args:   1,
			IsPure: true,
		}.SetDescription("n", "Returns a map with the key 'state' set to the given value.")).
		AddStaticFunction("sprintf", funcGen.Function[Value]{Func: sprintf, Args: -1, IsPure: true}.
			SetDescription("format", "args", "the classic, well known sprintf function")).
		AddStaticFunction("sqrt", simpleOnlyFloatFuncCheck("sqrt", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Sqrt(x) })).
		AddStaticFunction("ln", simpleOnlyFloatFuncCheck("ln", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Log(x) })).
		AddStaticFunction("log10", simpleOnlyFloatFuncCheck("log", func(arg float64) bool { return arg >= 0 }, func(x float64) float64 { return math.Log10(x) })).
		AddStaticFunction("exp", simpleOnlyFloatFunc("exp", func(x float64) float64 { return math.Exp(x) })).
		AddStaticFunction("sin", simpleOnlyFloatFunc("sin", func(x float64) float64 { return math.Sin(x) })).
		AddStaticFunction("cos", simpleOnlyFloatFunc("cos", func(x float64) float64 { return math.Cos(x) })).
		AddStaticFunction("tan", simpleOnlyFloatFunc("tan", func(x float64) float64 { return math.Tan(x) })).
		AddStaticFunction("asin", simpleOnlyFloatFuncCheck("asin", func(arg float64) bool { return arg >= -1 && arg <= 1 }, func(x float64) float64 { return math.Asin(x) })).
		AddStaticFunction("acos", simpleOnlyFloatFuncCheck("acos", func(arg float64) bool { return arg >= -1 && arg <= 1 }, func(x float64) float64 { return math.Acos(x) })).
		AddStaticFunction("atan", simpleOnlyFloatFunc("atan", func(x float64) float64 { return math.Atan(x) }))

	f.FunctionGenerator = fg

	registerOperators(f)

	f.RegisterMethods(ListTypeId, createListMethods(f))
	f.RegisterMethods(MapTypeId, createMapMethods())
	f.RegisterMethods(StringTypeId, createStringMethods())
	f.RegisterMethods(BoolTypeId, createBoolMethods())
	f.RegisterMethods(IntTypeId, createIntMethods())
	f.RegisterMethods(FloatTypeId, createFloatMethods())
	f.RegisterMethods(closureTypeId, createClosureMethods())
	f.AddStaticFunction("min", funcGen.Function[Value]{
		Func: func(st funcGen.Stack[Value], cs []Value) (Value, error) {
			var m Value
			for i := 0; i < st.Size(); i++ {
				v := st.Get(i)
				if i == 0 {
					m = v
				} else {
					le, err := less.Do(st, v, m)
					if err != nil {
						return nil, err
					}
					if b, ok := le.(Bool); ok {
						if b {
							m = v
						}
					} else {
						return nil, fmt.Errorf("less not allowed on %s, %s", TypeName(v), TypeName(m))
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
					b, err := less.Do(st, m, v)
					if err != nil {
						return nil, err
					}
					if b, ok := b.(Bool); ok {
						if b {
							m = v
						}
					} else {
						return nil, fmt.Errorf("less not allowed on %s, %s", TypeName(m), TypeName(v))
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
	if i, ok := v.ToInt(); ok {
		return i, nil
	}
	return 0, fmt.Errorf("not an int: %s", TypeName(v))
}
