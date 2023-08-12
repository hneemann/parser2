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
	Optimize(AST) AST
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
	Optimize(Optimizer)
	// String return a string representation of the AST
	String() string
	// GetLine returns the line in the source code
	GetLine() Line
}

func Optimize(ast AST, optimizer Optimizer) AST {
	ast.Optimize(optimizer)
	if o := optimizer.Optimize(ast); o != nil {
		return o
	}
	return ast
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

func (l *Let) Optimize(optimizer Optimizer) {
	l.Value.Optimize(optimizer)
	if o := optimizer.Optimize(l.Value); o != nil {
		l.Value = o
	}
	l.Inner.Optimize(optimizer)
	if o := optimizer.Optimize(l.Inner); o != nil {
		l.Inner = o
	}
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

func (i *If) Optimize(optimizer Optimizer) {
	i.Cond.Optimize(optimizer)
	if o := optimizer.Optimize(i.Cond); o != nil {
		i.Cond = o
	}
	i.Then.Optimize(optimizer)
	if o := optimizer.Optimize(i.Then); o != nil {
		i.Then = o
	}
	i.Else.Optimize(optimizer)
	if o := optimizer.Optimize(i.Else); o != nil {
		i.Else = o
	}
}

func (i *If) String() string {
	return "if " + i.Cond.String() + " then " + i.Then.String() + " else " + i.Else.String()
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

func (o *Operate) Optimize(optimizer Optimizer) {
	o.A.Optimize(optimizer)
	if opt := optimizer.Optimize(o.A); opt != nil {
		o.A = opt
	}
	o.B.Optimize(optimizer)
	if opt := optimizer.Optimize(o.B); opt != nil {
		o.B = opt
	}
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

func (u *Unary) Optimize(optimizer Optimizer) {
	u.Value.Optimize(optimizer)
	if o := optimizer.Optimize(u.Value); o != nil {
		u.Value = o
	}
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

func (m *MapAccess) Optimize(optimizer Optimizer) {
	m.MapValue.Optimize(optimizer)
	if o := optimizer.Optimize(m.MapValue); o != nil {
		m.MapValue = o
	}
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

func (m *MethodCall) Optimize(optimizer Optimizer) {
	m.Value.Optimize(optimizer)
	if o := optimizer.Optimize(m.Value); o != nil {
		m.Value = o
	}
	for i := range m.Args {
		m.Args[i].Optimize(optimizer)
		if o := optimizer.Optimize(m.Args[i]); o != nil {
			m.Args[i] = o
		}
	}
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

func (a *ListAccess) Optimize(optimizer Optimizer) {
	a.Index.Optimize(optimizer)
	if o := optimizer.Optimize(a.Index); o != nil {
		a.Index = o
	}
	a.List.Optimize(optimizer)
	if o := optimizer.Optimize(a.List); o != nil {
		a.List = o
	}
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

func (c *ClosureLiteral) Optimize(optimizer Optimizer) {
	c.Func.Optimize(optimizer)
	if o := optimizer.Optimize(c.Func); o != nil {
		c.Func = o
	}
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

func (ml *MapLiteral) Optimize(optimizer Optimizer) {
	for _, v := range ml.Map {
		v.Optimize(optimizer)
	}
	for k, v := range ml.Map {
		if o := optimizer.Optimize(v); o != nil {
			ml.Map[k] = o
		}
	}
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

func (al *ListLiteral) Optimize(optimizer Optimizer) {
	for i := range al.List {
		al.List[i].Optimize(optimizer)
		if o := optimizer.Optimize(al.List[i]); o != nil {
			al.List[i] = o
		}
	}
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

func (i *Ident) Optimize(Optimizer) {

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

func (n *Const[V]) Optimize(Optimizer) {
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

func (f *FunctionCall) Optimize(optimizer Optimizer) {
	f.Func.Optimize(optimizer)
	if o := optimizer.Optimize(f.Func); o != nil {
		f.Func = o
	}
	for i := range f.Args {
		f.Args[i].Optimize(optimizer)
		if o := optimizer.Optimize(f.Args[i]); o != nil {
			f.Args[i] = o
		}
	}
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
	number        Matcher
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
		operator:   simpleOperator,
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

// AllowComments allows C style end of line comments
func (p *Parser[V]) AllowComments() *Parser[V] {
	p.allowComments = true
	return p
}

// Parse parses the given string and returns an ast
func (p *Parser[V]) Parse(str string) (ast AST, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = EnhanceErrorf(rec, "error parsing expression")
			ast = nil
		}
	}()
	tokenizer :=
		NewTokenizer(str, p.number, p.identifier, p.operator, p.textOperators, p.allowComments)

	ast = p.parseExpression(tokenizer)
	t := tokenizer.Next()
	if t.typ != tEof {
		return nil, unexpected("EOF", t)
	}

	return ast, nil
}

type parserFunc func(tokenizer *Tokenizer) AST

func (p *Parser[V]) parseExpression(tokenizer *Tokenizer) AST {
	t := tokenizer.Peek()
	if t.typ == tIdent && t.image == "let" {
		tokenizer.Next()
		t = tokenizer.Next()
		if t.typ != tIdent {
			panic(t.Errorf("no identifier followed by let"))
		}
		name := t.image
		line := t.GetLine()
		if t := tokenizer.Next(); t.typ != tOperate || t.image != "=" {
			panic(unexpected("=", t))
		}
		exp := p.parse(tokenizer, 0)
		if t := tokenizer.Next(); t.typ != tSemicolon || t.image != ";" {
			panic(unexpected(";", t))
		}
		inner := p.parseExpression(tokenizer)
		return &Let{
			Name:  name,
			Value: exp,
			Inner: inner,
			Line:  line,
		}
	} else {
		return p.parse(tokenizer, 0)
	}
}

func (p *Parser[V]) parse(tokenizer *Tokenizer, op int) AST {
	next := p.nextParserCall(op)
	operator := p.operators[op]
	a := next(tokenizer)
	for {
		t := tokenizer.Peek()
		if t.typ == tOperate && t.image == operator {
			tokenizer.Next()
			aa := a
			bb := next(tokenizer)
			a = &Operate{
				Operator: operator,
				A:        aa,
				B:        bb,
				Line:     t.Line,
			}
		} else {
			return a
		}
	}
}

func (p *Parser[V]) nextParserCall(op int) parserFunc {
	if op+1 < len(p.operators) {
		return func(tokenizer *Tokenizer) AST {
			return p.parse(tokenizer, op+1)
		}
	} else {
		return p.parseUnary
	}
}

func (p *Parser[V]) parseUnary(tokenizer *Tokenizer) AST {
	if t := tokenizer.Peek(); t.typ == tOperate {
		if _, ok := p.unary[t.image]; ok {
			t = tokenizer.Next()
			e := p.parseNonOperator(tokenizer)
			return &Unary{
				Operator: t.image,
				Value:    e,
				Line:     t.Line,
			}
		}
	}
	return p.parseNonOperator(tokenizer)
}

func (p *Parser[V]) parseNonOperator(tokenizer *Tokenizer) AST {
	expression := p.parseLiteral(tokenizer)
	for {
		switch tokenizer.Peek().typ {
		case tDot:
			tokenizer.Next()
			t := tokenizer.Next()
			if t.typ != tIdent {
				panic(unexpected("ident", t))
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
				args := p.parseArgs(tokenizer, tClose)
				expression = &MethodCall{
					Name:  name,
					Args:  args,
					Value: expression,
					Line:  t.Line,
				}
			}
		case tOpen:
			t := tokenizer.Next()
			args := p.parseArgs(tokenizer, tClose)
			expression = &FunctionCall{
				Func: expression,
				Args: args,
				Line: t.Line,
			}

		case tOpenBracket:
			tokenizer.Next()
			indexExpr := p.parseExpression(tokenizer)
			t := tokenizer.Next()
			if t.typ != tCloseBracket {
				panic(unexpected("}", t))
			}
			expression = &ListAccess{
				Index: indexExpr,
				List:  expression,
				Line:  t.Line,
			}
		default:
			return expression
		}
	}
}

func (p *Parser[V]) parseLiteral(tokenizer *Tokenizer) AST {
	t := tokenizer.Next()
	switch t.typ {
	case tIdent:
		name := t.image
		if cl := tokenizer.Peek(); cl.typ == tOperate && cl.image == "->" {
			// closure, short definition x->[exp]
			tokenizer.Next()
			e := p.parseExpression(tokenizer)
			return &ClosureLiteral{
				Names: []string{name},
				Func:  e,
				Line:  t.Line,
			}
		} else if name == "if" {
			cond := p.parseExpression(tokenizer)
			t := tokenizer.Next()
			if !(t.typ == tIdent && t.image == "then") {
				panic(unexpected("then", t))
			}
			thenExp := p.parseExpression(tokenizer)
			t = tokenizer.Next()
			if !(t.typ == tIdent && t.image == "else") {
				panic(unexpected("else", t))
			}
			elseExp := p.parseExpression(tokenizer)
			return &If{
				Cond: cond,
				Then: thenExp,
				Else: elseExp,
				Line: t.Line,
			}

		} else {
			return &Ident{Name: name, Line: t.Line}
		}
	case tOpenCurly:
		return p.parseMap(tokenizer)
	case tOpenBracket:
		args := p.parseArgs(tokenizer, tCloseBracket)
		return &ListLiteral{args, t.Line}
	case tNumber:
		if p.numberParser != nil {
			if number, err := p.numberParser.ParseNumber(t.image); err == nil {
				return &Const[V]{number, t.Line}
			} else {
				panic(t.EnhanceErrorf(err, "not a number"))
			}
		}
	case tString:
		if p.stringHandler != nil {
			return &Const[V]{p.stringHandler.FromString(t.image), t.Line}
		}
	case tOpen:
		if tokenizer.Peek().typ == tIdent && tokenizer.PeekPeek().typ == tComma {
			names := p.parseIdentList(tokenizer)
			t = tokenizer.Next()
			if !(t.typ == tOperate && t.image == "->") {
				panic(unexpected("->", t))
			}
			e := p.parseExpression(tokenizer)
			return &ClosureLiteral{
				Names: names,
				Func:  e,
				Line:  t.Line,
			}
		} else {
			e := p.parseExpression(tokenizer)
			t := tokenizer.Next()
			if t.typ != tClose {
				panic(unexpected(")", t))
			}
			return e
		}
	}
	panic(t.Errorf("unexpected token type: %v", t.image))
}

func (p *Parser[V]) parseArgs(tokenizer *Tokenizer, closeList TokenType) []AST {
	var args []AST
	if tokenizer.Peek().typ == closeList {
		tokenizer.Next()
		return args
	}
	for {
		element := p.parseExpression(tokenizer)
		args = append(args, element)
		t := tokenizer.Next()
		if t.typ == closeList {
			return args
		}
		if t.typ != tComma {
			panic(unexpected(",", t))
		}
	}
}

func (p *Parser[V]) parseMap(tokenizer *Tokenizer) *MapLiteral {
	m := MapLiteral{Map: map[string]AST{}}
	for {
		switch t := tokenizer.Next(); t.typ {
		case tCloseCurly:
			m.Line = t.Line
			return &m
		case tIdent:
			if c := tokenizer.Next(); c.typ != tColon {
				panic(unexpected(":", c))
			}
			entry := p.parseExpression(tokenizer)
			m.Map[t.image] = entry
			if tokenizer.Peek().typ == tComma {
				tokenizer.Next()
			} else {
				if tokenizer.Peek().typ != tCloseCurly {
					found := tokenizer.Next()
					panic(t.Errorf("unexpected token, expected ',' or '}', found %v", found))
				}
			}
		default:
			panic(unexpected(",", t))
		}
	}
}

func (p *Parser[V]) parseIdentList(tokenizer *Tokenizer) []string {
	var names []string
	for {
		t := tokenizer.Next()
		if t.typ == tIdent {
			names = append(names, t.image)
			t = tokenizer.Next()
			switch t.typ {
			case tClose:
				return names
			case tComma:
			default:
				panic(t.Errorf("expected ',' or ')', found %v", t))
			}
		} else {
			panic(t.Errorf("expected identifier, found %v", t))
		}
	}
}

func unexpected(expected string, found Token) error {
	return found.Errorf("unexpected token, expected '%s', found %v", expected, found)
}
