// Package parser2 helps to implement simple, configurable expression parsers
package parser2

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Visitor interface {
	Visit(AST)
}

// Optimizer is used to perform optimization on ast level
type Optimizer interface {
	// Optimize takes an AST and tries to optimize it.
	// If an optimization is found, the optimizes AST is returned.
	// If no optimization is found, nil is returned.
	Optimize(AST) AST
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
}

func Optimize(ast AST, optimizer Optimizer) AST {
	ast.Optimize(optimizer)
	if o := optimizer.Optimize(ast); o != nil {
		return o
	}
	return ast
}

type Let struct {
	Name  string
	Value AST
	Inner AST
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

type Operate struct {
	Operator string
	A, B     AST
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

type ArrayAccess struct {
	Index AST
	List  AST
}

func (a *ArrayAccess) Traverse(visitor Visitor) {
	visitor.Visit(a)
	a.Index.Traverse(visitor)
	a.List.Traverse(visitor)
}

func (a *ArrayAccess) Optimize(optimizer Optimizer) {
	a.Index.Optimize(optimizer)
	if o := optimizer.Optimize(a.Index); o != nil {
		a.Index = o
	}
	a.List.Optimize(optimizer)
	if o := optimizer.Optimize(a.List); o != nil {
		a.List = o
	}
}

func (a *ArrayAccess) String() string {
	return braceStr(a.List) + "[" + a.Index.String() + "]"
}

type ClosureLiteral struct {
	Names []string
	Func  AST
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

type MapLiteral map[string]AST

func (ml MapLiteral) Traverse(visitor Visitor) {
	visitor.Visit(ml)
	for _, v := range ml {
		v.Traverse(visitor)
	}
}

func (ml MapLiteral) Optimize(optimizer Optimizer) {
	for _, v := range ml {
		v.Optimize(optimizer)
	}
	for k, v := range ml {
		if o := optimizer.Optimize(v); o != nil {
			ml[k] = o
		}
	}
}

func (ml MapLiteral) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for k, v := range ml {
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

type ListLiteral []AST

func (al ListLiteral) Traverse(visitor Visitor) {
	visitor.Visit(al)
	for _, v := range al {
		v.Traverse(visitor)
	}
}

func (al ListLiteral) Optimize(optimizer Optimizer) {
	for i := range al {
		al[i].Optimize(optimizer)
		if o := optimizer.Optimize(al[i]); o != nil {
			al[i] = o
		}
	}
}

func (al ListLiteral) String() string {
	return "[" + sliceToString(al) + "]"
}

type Ident string

func (i Ident) Traverse(visitor Visitor) {
	visitor.Visit(i)
}

func (i Ident) Optimize(optimizer Optimizer) {

}

func (i Ident) String() string {
	return string(i)
}

type Const[V any] struct {
	Value V
}

func (n Const[V]) Traverse(visitor Visitor) {
	visitor.Visit(n)
}

func (n Const[V]) Optimize(optimizer Optimizer) {
}

func (n Const[V]) String() string {
	return fmt.Sprint(n.Value)
}

type FunctionCall struct {
	Func AST
	Args []AST
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

type simpleNumber struct {
}

func (s simpleNumber) MatchesFirst(r rune) bool {
	return unicode.IsNumber(r)
}

func (s simpleNumber) Matches(r rune) bool {
	return s.MatchesFirst(r) || r == '.'
}

type simpleIdentifier struct {
}

func (s simpleIdentifier) MatchesFirst(r rune) bool {
	return unicode.IsLetter(r)
}

func (s simpleIdentifier) Matches(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

type simpleOperator struct {
}

func (s simpleOperator) MatchesFirst(r rune) bool {
	return s.Matches(r)
}

func (s simpleOperator) Matches(r rune) bool {
	return strings.ContainsRune("+-*/&|!~<=>^", r)
}

// NumberParser is used to convert a string to a number
type NumberParser[V any] interface {
	ParseNumber(n string) (V, error)
}

type NumberParserFunc[V any] func(n string) (V, error)

func (npf NumberParserFunc[V]) ParseNumber(n string) (V, error) {
	return npf(n)
}

type StringHandler[V any] interface {
	FromString(s string) V
}

// Parser is the base class of the parser
type Parser[V any] struct {
	operators     []string
	unary         map[string]struct{}
	textOperators map[string]string
	numberParser  NumberParser[V]
	stringHandler StringHandler[V]
	number        Matcher
	identifier    Matcher
	operator      Matcher
	allowComments bool
}

// NewParser creates a new Parser
func NewParser[V any]() *Parser[V] {
	return &Parser[V]{
		unary:      map[string]struct{}{},
		number:     simpleNumber{},
		identifier: simpleIdentifier{},
		operator:   simpleOperator{},
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

// SetStringHandler sets the string handler
func (p *Parser[V]) SetStringHandler(stringHandler StringHandler[V]) *Parser[V] {
	p.stringHandler = stringHandler
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
			if thisErr, ok := rec.(error); ok {
				err = thisErr
			} else {
				err = fmt.Errorf("%s", rec)
			}
			ast = nil
		}
	}()
	tokenizer :=
		NewTokenizer(str, p.number, p.identifier, p.operator, p.textOperators, p.allowComments)

	ast = p.parseExpression(tokenizer)
	t := tokenizer.Next()
	if t.typ != tEof {
		return nil, errors.New(unexpected("EOF", t))
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
			panic("no identifier followed by let")
		}
		name := t.image
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
				panic("invalid token: " + t.image)
			}
			name := t.image
			if tokenizer.Peek().typ != tOpen {
				expression = &MapAccess{
					Key:      name,
					MapValue: expression,
				}
			} else {
				//Method call
				tokenizer.Next()
				args := p.parseArgs(tokenizer, tClose)
				expression = &MethodCall{
					Name:  name,
					Args:  args,
					Value: expression,
				}
			}
		case tOpen:
			tokenizer.Next()
			args := p.parseArgs(tokenizer, tClose)
			expression = &FunctionCall{
				Func: expression,
				Args: args,
			}

		case tOpenBracket:
			tokenizer.Next()
			indexExpr := p.parseExpression(tokenizer)
			t := tokenizer.Next()
			if t.typ != tCloseBracket {
				panic(unexpected("}", t))
			}
			expression = &ArrayAccess{
				Index: indexExpr,
				List:  expression,
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
			}
		} else {
			if name == "closure" {
				// multi arg closure definition: closure(a,b)->[exp]
				t := tokenizer.Next()
				if !(t.typ == tOpen) {
					panic(unexpected("(", t))
				}
				names := p.parseIdentList(tokenizer)
				t = tokenizer.Next()
				if !(t.typ == tOperate && t.image == "->") {
					panic(unexpected("->", t))
				}
				e := p.parseExpression(tokenizer)
				return &ClosureLiteral{
					Names: names,
					Func:  e,
				}
			} else {
				return Ident(name)
			}
		}
	case tOpenCurly:
		return p.parseMap(tokenizer)
	case tOpenBracket:
		args := p.parseArgs(tokenizer, tCloseBracket)
		return ListLiteral(args)
	case tNumber:
		if p.numberParser != nil {
			if number, err := p.numberParser.ParseNumber(t.image); err == nil {
				return Const[V]{number}
			} else {
				panic(fmt.Sprintf("not a number: %v", err))
			}
		}
	case tString:
		if p.stringHandler != nil {
			return Const[V]{p.stringHandler.FromString(t.image)}
		}
	case tOpen:
		e := p.parseExpression(tokenizer)
		t := tokenizer.Next()
		if t.typ != tClose {
			panic(unexpected(")", t))
		}
		return e
	}
	panic("unexpected token type: " + t.image)
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

func (p *Parser[V]) parseMap(tokenizer *Tokenizer) MapLiteral {
	m := MapLiteral{}
	for {
		switch t := tokenizer.Next(); t.typ {
		case tCloseCurly:
			return m
		case tIdent:
			if c := tokenizer.Next(); c.typ != tColon {
				panic(unexpected(":", c))
			}
			entry := p.parseExpression(tokenizer)
			m[t.image] = entry
			if tokenizer.Peek().typ == tComma {
				tokenizer.Next()
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
				panic("expected ',' or ')'")
			}
		} else {
			panic("expected identifier")
		}
	}
}

func unexpected(expected string, found Token) string {
	return fmt.Sprintf("unexpected token, expected '%s', found %v", expected, found)
}
