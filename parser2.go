// Package parser2 helps to implement simple, configurable expression parsers
package parser2

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Visitor interface {
	Visit(AST)
}

type VisitorFunc func(AST)

func (v VisitorFunc) Visit(a AST) {
	v(a)
}

// Optimizer is used to perform optimization on ast level
type Optimizer interface {
	// Optimize takes an AST and tries to optimize it.
	// If an optimization is found, the optimizes AST is returned.
	// If no optimization is found, nil is returned.
	Optimize(AST) (AST, error)
}

type OptimizerFunc func(AST) AST

func (o OptimizerFunc) Optimize(ast AST) AST {
	return o(ast)
}

// AST represents a node in the AST
type AST interface {
	// Traverse visits the complete AST
	Traverse(Visitor)
	// Optimize is called to optimize the AST
	// At first the children Optimize method is called and
	// After that the own node is to be optimized.
	Optimize(Optimizer) error
	// String return a string representation of the AST
	String() string
	// GetLine returns the line in the source code
	GetLine() Line
}

// Optimize uses the given optimizer to optimize the given AST.
// If no optimization is possible, the given AST is returned unchanged.
func Optimize(ast AST, optimizer Optimizer) (astRet AST, errRet error) {
	defer func() {
		rec := recover()
		if rec != nil {
			if err, ok := rec.(error); ok {
				errRet = err
			} else {
				errRet = fmt.Errorf("%v", rec)
			}
			astRet = nil
		}
	}()
	err := ast.Optimize(optimizer)
	if err != nil {
		return nil, err
	}
	o, err := optimizer.Optimize(ast)
	if err != nil {
		return nil, err
	}
	if o != nil {
		return o, nil
	}
	return ast, nil
}

type Line int

func (l Line) GetLine() Line {
	return l
}

type errorWithLine struct {
	message string
	line    Line
	cause   error
}

func (e errorWithLine) Error() string {
	m := e.message
	if e.line > 0 {
		m += " in line " + strconv.Itoa(int(e.line))
	}
	if e.cause != nil {
		m += ";\n cause: " + e.cause.Error()
	}
	return m
}

func (l Line) Errorf(m string, a ...any) error {
	return errorWithLine{
		message: fmt.Sprintf(m, a...),
		line:    l,
	}
}

func enhanceErrorfInternal(cause any, m string, a ...any) errorWithLine {
	c, ok := cause.(error)
	if !ok {
		c = fmt.Errorf("%v", cause)
	}
	return errorWithLine{
		message: fmt.Sprintf(m, a...),
		cause:   c,
	}
}

func EnhanceErrorf(cause any, m string, a ...any) error {
	return enhanceErrorfInternal(cause, m, a...)
}

func (l Line) EnhanceErrorf(cause any, m string, a ...any) error {
	e := enhanceErrorfInternal(cause, m, a...)
	e.line = l
	return e
}

type Let struct {
	Name  string
	Value AST
	Inner AST
	Line
}

func (l *Let) Traverse(visitor Visitor) {
	visitor.Visit(l)
	l.Value.Traverse(visitor)
	l.Inner.Traverse(visitor)
}

func (l *Let) String() string {
	return "let " + l.Name + "=" + l.Value.String() + "; " + l.Inner.String()
}

func opt(a *AST, optimizer Optimizer) error {
	err := (*a).Optimize(optimizer)
	if err != nil {
		return err
	}
	o, err := optimizer.Optimize(*a)
	if err != nil {
		return err
	}
	if o != nil {
		*a = o
	}
	return nil
}

func (l *Let) Optimize(optimizer Optimizer) error {
	err := opt(&l.Value, optimizer)
	if err != nil {
		return err
	}
	return opt(&l.Inner, optimizer)
}

type If struct {
	Cond AST
	Then AST
	Else AST
	Line
}

func (i *If) Traverse(visitor Visitor) {
	visitor.Visit(i)
	i.Cond.Traverse(visitor)
	i.Then.Traverse(visitor)
	i.Else.Traverse(visitor)
}

func (i *If) Optimize(optimizer Optimizer) error {
	err := opt(&i.Cond, optimizer)
	if err != nil {
		return err
	}
	err = opt(&i.Then, optimizer)
	if err != nil {
		return err
	}
	return opt(&i.Else, optimizer)
}

func (i *If) String() string {
	return "if " + i.Cond.String() + " then " + i.Then.String() + " else " + i.Else.String()
}

type Case[V any] struct {
	CaseConst AST
	Value     AST
}

type Switch[V any] struct {
	SwitchValue AST
	Cases       []Case[V]
	Default     AST
	Line
}

func (s *Switch[V]) Traverse(visitor Visitor) {
	visitor.Visit(s)
	s.SwitchValue.Traverse(visitor)
	for _, c := range s.Cases {
		c.Value.Traverse(visitor)
	}
	s.Default.Traverse(visitor)
}

func (s *Switch[V]) Optimize(o Optimizer) error {
	err := opt(&s.SwitchValue, o)
	if err != nil {
		return err
	}
	for _, c := range s.Cases {
		err := opt(&c.Value, o)
		if err != nil {
			return err
		}
	}
	return opt(&s.Default, o)
}

func (s *Switch[V]) String() string {
	var b bytes.Buffer
	b.WriteString("switch ")
	b.WriteString(s.SwitchValue.String())
	for _, c := range s.Cases {
		b.WriteString(" case ")
		b.WriteString(fmt.Sprint(c.CaseConst))
		b.WriteString(" : ")
		b.WriteString(c.Value.String())
	}
	b.WriteString(" default ")
	b.WriteString(s.Default.String())
	return b.String()
}

type Operate struct {
	Operator string
	A, B     AST
	Line
}

func (o *Operate) Traverse(visitor Visitor) {
	visitor.Visit(o)
	o.A.Traverse(visitor)
	o.B.Traverse(visitor)
}

func (o *Operate) Optimize(optimizer Optimizer) error {
	err := opt(&o.A, optimizer)
	if err != nil {
		return err
	}
	return opt(&o.B, optimizer)
}

func (o *Operate) String() string {
	return braceStr(o.A) + o.Operator + braceStr(o.B)
}

func braceStr(a AST) string {
	if _, ok := a.(*Operate); ok {
		return "(" + a.String() + ")"
	}
	return a.String()
}

type Unary struct {
	Operator string
	Value    AST
	Line
}

func (u *Unary) Traverse(visitor Visitor) {
	visitor.Visit(u)
	u.Value.Traverse(visitor)
}

func (u *Unary) Optimize(optimizer Optimizer) error {
	return opt(&u.Value, optimizer)
}

func (u *Unary) String() string {
	return u.Operator + braceStr(u.Value)
}

type MapAccess struct {
	Key      string
	MapValue AST
	Line
}

func (m *MapAccess) Traverse(visitor Visitor) {
	visitor.Visit(m)
	m.MapValue.Traverse(visitor)
}

func (m *MapAccess) Optimize(optimizer Optimizer) error {
	return opt(&m.MapValue, optimizer)
}

func (m *MapAccess) String() string {
	return braceStr(m.MapValue) + "." + m.Key
}

type MethodCall struct {
	Name  string
	Args  []AST
	Value AST
	Line
}

func (m *MethodCall) Traverse(visitor Visitor) {
	visitor.Visit(m)
	for _, a := range m.Args {
		a.Traverse(visitor)
	}
	m.Value.Traverse(visitor)
}

func (m *MethodCall) Optimize(optimizer Optimizer) error {
	err := opt(&m.Value, optimizer)
	if err != nil {
		return err
	}
	for i := range m.Args {
		err := opt(&m.Args[i], optimizer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MethodCall) String() string {
	return braceStr(m.Value) + "." + m.Name + "(" + sliceToString(m.Args) + ")"
}

func sliceToString[V fmt.Stringer](items []V) string {
	b := bytes.Buffer{}
	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.String())
	}
	return b.String()
}

func stringsToString(items []string) string {
	b := bytes.Buffer{}
	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item)
	}
	return b.String()
}

type ListAccess struct {
	Index AST
	List  AST
	Line
}

func (a *ListAccess) Traverse(visitor Visitor) {
	visitor.Visit(a)
	a.Index.Traverse(visitor)
	a.List.Traverse(visitor)
}

func (a *ListAccess) Optimize(optimizer Optimizer) error {
	err := opt(&a.Index, optimizer)
	if err != nil {
		return err
	}
	return opt(&a.List, optimizer)
}

func (a *ListAccess) String() string {
	return braceStr(a.List) + "[" + a.Index.String() + "]"
}

type ClosureLiteral struct {
	Names []string
	Func  AST
	Line
}

func (c *ClosureLiteral) Traverse(visitor Visitor) {
	visitor.Visit(c)
	c.Func.Traverse(visitor)
}

func (c *ClosureLiteral) Optimize(optimizer Optimizer) error {
	return opt(&c.Func, optimizer)
}

func (c *ClosureLiteral) String() string {
	return "(" + stringsToString(c.Names) + ")->" + c.Func.String()
}

type MapLiteral struct {
	Map map[string]AST
	Line
}

func (ml *MapLiteral) Traverse(visitor Visitor) {
	visitor.Visit(ml)
	for _, v := range ml.Map {
		v.Traverse(visitor)
	}
}

func (ml *MapLiteral) Optimize(optimizer Optimizer) error {
	m := map[string]AST{}
	for k, v := range ml.Map {
		err := opt(&v, optimizer)
		if err != nil {
			return err
		}
		m[k] = v
	}
	ml.Map = m
	return nil
}

func (ml *MapLiteral) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for k, v := range ml.Map {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteString(":")
		b.WriteString(v.String())
	}
	b.WriteString("}")
	return b.String()
}

type ListLiteral struct {
	List []AST
	Line
}

func (al *ListLiteral) Traverse(visitor Visitor) {
	visitor.Visit(al)
	for _, v := range al.List {
		v.Traverse(visitor)
	}
}

func (al *ListLiteral) Optimize(optimizer Optimizer) error {
	for i := range al.List {
		err := opt(&al.List[i], optimizer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (al *ListLiteral) String() string {
	return "[" + sliceToString(al.List) + "]"
}

type Ident struct {
	Name string
	Line
}

func (i *Ident) Traverse(visitor Visitor) {
	visitor.Visit(i)
}

func (i *Ident) Optimize(Optimizer) error {
	return nil
}

func (i *Ident) String() string {
	return i.Name
}

type Const[V any] struct {
	Value V
	Line
}

func (n *Const[V]) Traverse(visitor Visitor) {
	visitor.Visit(n)
}

func (n *Const[V]) Optimize(Optimizer) error {
	return nil
}

func (n *Const[V]) String() string {
	return fmt.Sprint(n.Value)
}

type FunctionCall struct {
	Func AST
	Args []AST
	Line
}

func (f *FunctionCall) Traverse(visitor Visitor) {
	visitor.Visit(f)
	f.Func.Traverse(visitor)
	for _, a := range f.Args {
		a.Traverse(visitor)
	}
}

func (f *FunctionCall) Optimize(optimizer Optimizer) error {
	err := opt(&f.Func, optimizer)
	if err != nil {
		return err
	}
	for i := range f.Args {
		err := opt(&f.Args[i], optimizer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FunctionCall) String() string {
	return braceStr(f.Func) + "(" + sliceToString(f.Args) + ")"
}

func simpleNumber(r rune) (func(r rune) bool, bool) {
	if unicode.IsNumber(r) {
		var last rune
		return func(r rune) bool {
			ok := unicode.IsNumber(r) || r == '.' || r == 'e' || (last == 'e' && r == '-') || (last == 'e' && r == '+')
			last = r
			return ok
		}, true
	} else {
		return nil, false
	}
}

func simpleIdentifier(r rune) (func(r rune) bool, bool) {
	if unicode.IsLetter(r) {
		return func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
		}, true
	} else {
		return nil, false
	}
}

func simpleOperator(r rune) (func(r rune) bool, bool) {
	const opStr = "+-*/&|!~<=>^"

	if strings.ContainsRune(opStr, r) {
		return func(r rune) bool {
			return strings.ContainsRune(opStr, r)
		}, true
	} else {
		return nil, false
	}
}

// NumberParser is used to convert a string to a number
type NumberParser[V any] interface {
	ParseNumber(n string) (V, error)
}

type NumberParserFunc[V any] func(n string) (V, error)

func (npf NumberParserFunc[V]) ParseNumber(n string) (V, error) {
	return npf(n)
}

type StringConverter[V any] interface {
	FromString(s string) V
}

type StringConverterFunc[V any] func(n string) V

func (shf StringConverterFunc[V]) FromString(s string) V {
	return shf(s)
}

// Parser is the base class of the parser
type Parser[V any] struct {
	operators     []string
	unary         map[string]struct{}
	textOperators map[string]string
	numberParser  NumberParser[V]
	stringHandler StringConverter[V]
	optimizer     Optimizer
	number        Matcher
	constants     Constants[V]
	identifier    Matcher
	operator      Matcher
	allowComments bool
}

// NewParser creates a new Parser
func NewParser[V any]() *Parser[V] {
	return &Parser[V]{
		unary:      map[string]struct{}{},
		number:     simpleNumber,
		identifier: simpleIdentifier,
		constants: ConstantsFunc[V](func(name string) (V, bool) {
			var zero V
			return zero, false
		}),
		operator: simpleOperator,
	}
}

// Op adds an operation to the parser
// The name gives the operations name e.g."+"
// The operation with the lowest priority needs to be added first.
// The operation with the highest priority needs to be added last.
func (p *Parser[V]) Op(name ...string) *Parser[V] {
	if len(p.operators) == 0 {
		p.operators = name
	} else {
		p.operators = append(p.operators, name...)
	}
	return p
}

// Unary is used to declare unary operations like "-" or "!".
func (p *Parser[V]) Unary(operators ...string) *Parser[V] {
	for _, o := range operators {
		p.unary[o] = struct{}{}
	}
	return p
}

// SetNumberParser sets the number parser
func (p *Parser[V]) SetNumberParser(numberParser NumberParser[V]) *Parser[V] {
	p.numberParser = numberParser
	return p
}

// SetStringConverter sets the string handler
func (p *Parser[V]) SetStringConverter(stringConverter StringConverter[V]) *Parser[V] {
	p.stringHandler = stringConverter
	return p
}

// SetNumberMatcher sets the number Matcher
func (p *Parser[V]) SetNumberMatcher(num Matcher) *Parser[V] {
	p.number = num
	return p
}

// SetIdentMatcher sets the identifier Matcher
func (p *Parser[V]) SetIdentMatcher(ident Matcher) *Parser[V] {
	p.identifier = ident
	return p
}

// SetOperatorMatcher sets the operator Matcher
func (p *Parser[V]) SetOperatorMatcher(operator Matcher) *Parser[V] {
	p.operator = operator
	return p
}

// TextOperator sets a map of text aliases for operators.
// Allows setting e.g. "plus" as an alias for "+"
func (p *Parser[V]) TextOperator(textOperators map[string]string) *Parser[V] {
	p.textOperators = textOperators
	return p
}

// SetConstants sets the constants for the parser
func (p *Parser[V]) SetConstants(constants Constants[V]) *Parser[V] {
	p.constants = constants
	return p
}

// SetOptimizer sets a optimizer used to optimize constants
func (p *Parser[V]) SetOptimizer(optimizer Optimizer) *Parser[V] {
	p.optimizer = optimizer
	return p
}

// AllowComments allows C style end of line comments
func (p *Parser[V]) AllowComments() *Parser[V] {
	p.allowComments = true
	return p
}

// Parse parses the given string and returns an ast
func (p *Parser[V]) Parse(str string) (ast AST, err error) {
	tokenizer :=
		NewTokenizer(str, p.number, p.identifier, p.operator, p.textOperators, p.allowComments)

	ast, err = p.parseExpression(tokenizer, p.constants)
	if err != nil {
		return nil, err
	}
	t := tokenizer.Next()
	if t.typ != tEof {
		return nil, unexpected("EOF", t)
	}

	return ast, nil
}

type parserFunc[V any] func(tokenizer *Tokenizer, constants Constants[V]) (AST, error)

type Constants[V any] interface {
	GetConst(name string) (V, bool)
}

type ConstantsFunc[V any] func(name string) (V, bool)

func (cf ConstantsFunc[V]) GetConst(name string) (V, bool) {
	return cf(name)
}

type constant[V any] struct {
	name  string
	value V
	other Constants[V]
}

func (c *constant[V]) GetConst(name string) (V, bool) {
	if c.name == name {
		return c.value, true
	}
	if c == nil {
		var zero V
		return zero, false
	}
	return c.other.GetConst(name)
}

func (p *Parser[V]) parseExpression(tokenizer *Tokenizer, constants Constants[V]) (AST, error) {
	t := tokenizer.Peek()
	if t.typ == tIdent {
		if t.image == "const" {
			tokenizer.Next()
			t = tokenizer.Next()
			if t.typ != tIdent {
				return nil, t.Errorf("no identifier followed by const")
			}
			name := t.image
			if _, ok := constants.GetConst(name); ok {
				return nil, t.Errorf("there is already a constant named '%s'", name)
			}
			if t := tokenizer.Next(); t.typ != tOperate || t.image != "=" {
				return nil, unexpected("=", t)
			}
			exp, err := p.parse(tokenizer, 0, constants)
			if err != nil {
				return nil, err
			}
			if t := tokenizer.Next(); t.typ != tSemicolon || t.image != ";" {
				return nil, unexpected(";", t)
			}
			if p.optimizer != nil {
				exp, err = Optimize(exp, p.optimizer)
				if err != nil {
					return nil, t.EnhanceErrorf(err, "error optimizing a constant")
				}
			}
			if c, ok := exp.(*Const[V]); ok {
				constants = &constant[V]{
					name:  name,
					value: c.Value,
					other: constants,
				}
			} else {
				return nil, t.Errorf("not a constant")
			}
			return p.parseExpression(tokenizer, constants)
		} else if t.image == "let" {
			tokenizer.Next()
			t = tokenizer.Next()
			if t.typ != tIdent {
				return nil, t.Errorf("no identifier followed by let")
			}
			name := t.image
			if _, ok := constants.GetConst(name); ok {
				return nil, t.Errorf("there is already a constant named '%s'", name)
			}
			line := t.GetLine()
			if t := tokenizer.Next(); t.typ != tOperate || t.image != "=" {
				return nil, unexpected("=", t)
			}
			exp, err := p.parse(tokenizer, 0, constants)
			if err != nil {
				return nil, err
			}
			if t := tokenizer.Next(); t.typ != tSemicolon || t.image != ";" {
				return nil, unexpected(";", t)
			}
			inner, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &Let{
				Name:  name,
				Value: exp,
				Inner: inner,
				Line:  line,
			}, nil
		} else if t.image == "func" {
			tokenizer.Next()
			t = tokenizer.Next()
			if t.typ != tIdent {
				return nil, t.Errorf("no identifier followed by func")
			}
			name := t.image
			if _, ok := constants.GetConst(name); ok {
				return nil, t.Errorf("there is already a constant named '%s'", name)
			}
			line := t.GetLine()
			if t := tokenizer.Next(); t.typ != tOpen {
				return nil, unexpected("(", t)
			}
			names, err := p.parseIdentList(tokenizer)
			if err != nil {
				return nil, err
			}
			exp, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			if t := tokenizer.Next(); t.typ != tSemicolon || t.image != ";" {
				return nil, unexpected(";", t)
			}
			inner, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &Let{
				Name: name,
				Value: &ClosureLiteral{
					Names: names,
					Func:  exp,
					Line:  line,
				},
				Inner: inner,
				Line:  line,
			}, nil
		}
	}
	return p.parse(tokenizer, 0, constants)
}

func (p *Parser[V]) parse(tokenizer *Tokenizer, op int, constants Constants[V]) (AST, error) {
	next := p.nextParserCall(op)
	operator := p.operators[op]
	a, err := next(tokenizer, constants)
	if err != nil {
		return nil, err
	}
	for {
		t := tokenizer.Peek()
		if t.typ == tOperate && t.image == operator {
			tokenizer.Next()
			aa := a
			bb, err := next(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			a = &Operate{
				Operator: operator,
				A:        aa,
				B:        bb,
				Line:     t.Line,
			}
		} else {
			return a, nil
		}
	}
}

func (p *Parser[V]) nextParserCall(op int) parserFunc[V] {
	if op+1 < len(p.operators) {
		return func(tokenizer *Tokenizer, constants Constants[V]) (AST, error) {
			return p.parse(tokenizer, op+1, constants)
		}
	} else {
		return p.parseUnary
	}
}

func (p *Parser[V]) parseUnary(tokenizer *Tokenizer, constants Constants[V]) (AST, error) {
	if t := tokenizer.Peek(); t.typ == tOperate {
		if _, ok := p.unary[t.image]; ok {
			t = tokenizer.Next()
			e, err := p.parseNonOperator(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &Unary{
				Operator: t.image,
				Value:    e,
				Line:     t.Line,
			}, nil
		}
	}
	return p.parseNonOperator(tokenizer, constants)
}

func (p *Parser[V]) parseNonOperator(tokenizer *Tokenizer, constants Constants[V]) (AST, error) {
	expression, err := p.parseLiteral(tokenizer, constants)
	if err != nil {
		return nil, err
	}
	for {
		switch tokenizer.Peek().typ {
		case tDot:
			tokenizer.Next()
			t := tokenizer.Next()
			if t.typ != tIdent {
				return nil, unexpected("ident", t)
			}
			name := t.image
			if tokenizer.Peek().typ != tOpen {
				expression = &MapAccess{
					Key:      name,
					MapValue: expression,
					Line:     t.Line,
				}
			} else {
				//Method call
				tokenizer.Next()
				args, err := p.parseArgs(tokenizer, tClose, constants)
				if err != nil {
					return nil, err
				}
				expression = &MethodCall{
					Name:  name,
					Args:  args,
					Value: expression,
					Line:  t.Line,
				}
			}
		case tOpen:
			t := tokenizer.Next()
			args, err := p.parseArgs(tokenizer, tClose, constants)
			if err != nil {
				return nil, err
			}
			expression = &FunctionCall{
				Func: expression,
				Args: args,
				Line: t.Line,
			}

		case tOpenBracket:
			tokenizer.Next()
			indexExpr, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			t := tokenizer.Next()
			if t.typ != tCloseBracket {
				return nil, unexpected("}", t)
			}
			expression = &ListAccess{
				Index: indexExpr,
				List:  expression,
				Line:  t.Line,
			}
		default:
			return expression, nil
		}
	}
}

func (p *Parser[V]) parseLiteral(tokenizer *Tokenizer, constants Constants[V]) (AST, error) {
	t := tokenizer.Next()
	switch t.typ {
	case tIdent:
		name := t.image
		if cl := tokenizer.Peek(); cl.typ == tOperate && cl.image == "->" {
			// closure, short definition x->[exp]
			tokenizer.Next()
			e, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &ClosureLiteral{
				Names: []string{name},
				Func:  e,
				Line:  t.Line,
			}, nil
		} else if name == "if" {
			cond, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			t := tokenizer.Next()
			if !(t.typ == tIdent && t.image == "then") {
				return nil, unexpected("then", t)
			}
			thenExp, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			t = tokenizer.Next()
			if !(t.typ == tIdent && t.image == "else") {
				return nil, unexpected("else", t)
			}
			elseExp, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &If{
				Cond: cond,
				Then: thenExp,
				Else: elseExp,
				Line: t.Line,
			}, nil
		} else if name == "switch" {
			switchValue, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			var cases []Case[V]
			for {
				t := tokenizer.Next()
				if t.typ == tIdent {
					if t.image == "case" {
						constFunc, err := p.parseExpression(tokenizer, constants)
						if err != nil {
							return nil, err
						}
						t = tokenizer.Next()
						if !(t.typ == tColon) {
							return nil, unexpected(":", t)
						}
						resultExp, err := p.parseExpression(tokenizer, constants)
						if err != nil {
							return nil, err
						}
						cases = append(cases, Case[V]{
							CaseConst: constFunc,
							Value:     resultExp,
						})
					} else if t.image == "default" {
						resultExp, err := p.parseExpression(tokenizer, constants)
						if err != nil {
							return nil, err
						}
						return &Switch[V]{
							SwitchValue: switchValue,
							Cases:       cases,
							Default:     resultExp,
							Line:        t.Line,
						}, nil
					} else {
						return nil, unexpected("case or default", t)
					}
				} else {
					return nil, unexpected("case or default", t)
				}
			}
		} else {
			if v, ok := constants.GetConst(name); ok {
				return &Const[V]{
					Value: v,
					Line:  t.Line,
				}, nil
			} else {
				return &Ident{Name: name, Line: t.Line}, nil
			}
		}
	case tOpenCurly:
		return p.parseMap(tokenizer, constants)
	case tOpenBracket:
		args, err := p.parseArgs(tokenizer, tCloseBracket, constants)
		if err != nil {
			return nil, err
		}
		return &ListLiteral{args, t.Line}, nil
	case tNumber:
		if p.numberParser != nil {
			if number, err := p.numberParser.ParseNumber(t.image); err == nil {
				return &Const[V]{number, t.Line}, nil
			} else {
				return nil, t.EnhanceErrorf(err, "not a number")
			}
		}
	case tString:
		if p.stringHandler != nil {
			return &Const[V]{p.stringHandler.FromString(t.image), t.Line}, nil
		}
	case tOpen:
		if tokenizer.Peek().typ == tIdent && tokenizer.PeekPeek().typ == tComma {
			names, err := p.parseIdentList(tokenizer)
			if err != nil {
				return nil, err
			}
			t = tokenizer.Next()
			if !(t.typ == tOperate && t.image == "->") {
				return nil, unexpected("->", t)
			}
			e, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			return &ClosureLiteral{
				Names: names,
				Func:  e,
				Line:  t.Line,
			}, nil
		} else {
			e, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			t := tokenizer.Next()
			if t.typ != tClose {
				return nil, unexpected(")", t)
			}
			return e, nil
		}
	}
	return nil, t.Errorf("unexpected token type: %v", t.image)
}

func (p *Parser[V]) parseArgs(tokenizer *Tokenizer, closeList TokenType, constants Constants[V]) ([]AST, error) {
	var args []AST
	if tokenizer.Peek().typ == closeList {
		tokenizer.Next()
		return args, nil
	}
	for {
		element, err := p.parseExpression(tokenizer, constants)
		if err != nil {
			return nil, err
		}
		args = append(args, element)
		t := tokenizer.Next()
		if t.typ == closeList {
			return args, nil
		}
		if t.typ != tComma {
			return nil, unexpected(",", t)
		}
	}
}

func (p *Parser[V]) parseMap(tokenizer *Tokenizer, constants Constants[V]) (*MapLiteral, error) {
	m := MapLiteral{Map: map[string]AST{}}
	for {
		switch t := tokenizer.Next(); t.typ {
		case tCloseCurly:
			m.Line = t.Line
			return &m, nil
		case tIdent:
			if c := tokenizer.Next(); c.typ != tColon {
				return nil, unexpected(":", c)
			}
			entry, err := p.parseExpression(tokenizer, constants)
			if err != nil {
				return nil, err
			}
			m.Map[t.image] = entry
			if tokenizer.Peek().typ == tComma {
				tokenizer.Next()
			} else {
				if tokenizer.Peek().typ != tCloseCurly {
					found := tokenizer.Next()
					return nil, t.Errorf("unexpected token, expected ',' or '}', found %v", found)
				}
			}
		default:
			return nil, unexpected(",", t)
		}
	}
}

func (p *Parser[V]) parseIdentList(tokenizer *Tokenizer) ([]string, error) {
	var names []string
	for {
		t := tokenizer.Next()
		if t.typ == tIdent {
			names = append(names, t.image)
			t = tokenizer.Next()
			switch t.typ {
			case tClose:
				return names, nil
			case tComma:
			default:
				return nil, t.Errorf("expected ',' or ')', found %v", t)
			}
		} else {
			return nil, t.Errorf("expected identifier, found %v", t)
		}
	}
}

func unexpected(expected string, found Token) error {
	return found.Errorf("unexpected token, expected '%s', found %v", expected, found)
}
