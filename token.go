package parser2

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type TokenType int

const (
	tIdent TokenType = iota
	tOpen
	tClose
	tOpenBracket
	tCloseBracket
	tOpenCurly
	tCloseCurly
	tDot
	tComma
	tColon
	tSemicolon
	tNumber
	tString
	tOperate
	tEof
	tInvalid
)

const (
	EOF rune = 0
)

var TokenEof = Token{tEof, "EOF", -1}

type Token struct {
	typ   TokenType
	image string
	Line
}

func (t Token) String() string {
	return fmt.Sprintf("'%v'", t.image)
}

type Tokenizer struct {
	str           string
	isLast        bool
	last          rune
	tok           chan Token
	isToken       bool
	token         Token
	line          Line
	number        Matcher
	identifier    Matcher
	operator      Matcher
	textOperators map[string]string
	allowComments bool
}

type Matcher func(r rune) (func(r rune) bool, bool)

func NewTokenizer(text string, number, identifier, operator Matcher, textOp map[string]string, allowComments bool) *Tokenizer {
	t := make(chan Token)
	tok := &Tokenizer{
		str:           text,
		textOperators: textOp,
		number:        number,
		identifier:    identifier,
		operator:      operator,
		allowComments: allowComments,
		line:          1,
		tok:           t}
	go tok.run(t)
	return tok
}

func (t *Tokenizer) Peek() Token {
	if t.isToken {
		return t.token
	}

	var ok bool
	t.token, ok = <-t.tok
	if ok {
		t.isToken = true
		return t.token
	} else {
		return TokenEof
	}
}

func (t *Tokenizer) Next() Token {
	tok := t.Peek()
	t.isToken = false
	return tok
}

func (t *Tokenizer) getLine() Line {
	return t.line
}

func (t *Tokenizer) run(tokens chan<- Token) {
	for {
		switch t.next(true) {
		case '\n':
			t.line++
			continue
		case ' ', '\r', '\t':
			continue
		case EOF:
			close(tokens)
			return
		case '(':
			tokens <- Token{tOpen, "(", t.getLine()}
		case ')':
			tokens <- Token{tClose, ")", t.getLine()}
		case '[':
			tokens <- Token{tOpenBracket, "[", t.getLine()}
		case ']':
			tokens <- Token{tCloseBracket, "]", t.getLine()}
		case '{':
			tokens <- Token{tOpenCurly, "{", t.getLine()}
		case '}':
			tokens <- Token{tCloseCurly, "}", t.getLine()}
		case '.':
			tokens <- Token{tDot, ".", t.getLine()}
		case ':':
			tokens <- Token{tColon, ":", t.getLine()}
		case ',':
			tokens <- Token{tComma, ",", t.getLine()}
		case ';':
			tokens <- Token{tSemicolon, ";", t.getLine()}
		case '"':
			image := t.readSkip(func(c rune) bool { return c != '"' }, false)
			t.next(false)
			tokens <- Token{tString, image, t.getLine()}
		case '\'':
			image := t.readSkip(func(c rune) bool { return c != '\'' }, false)
			t.next(false)
			tokens <- Token{tIdent, image, t.getLine()}
		default:
			t.unread()
			c := t.peek(true)
			if f, ok := t.number(c); ok {
				image := t.read(f)
				tokens <- Token{tNumber, image, t.getLine()}
			} else if f, ok := t.identifier(c); ok {
				image := t.read(f)
				if to, ok := t.textOperators[image]; ok {
					tokens <- Token{tOperate, to, t.getLine()}
				} else {
					tokens <- Token{tIdent, image, t.getLine()}
				}
			} else if f, ok := t.operator(c); ok {
				image := t.read(f)
				tokens <- Token{tOperate, image, t.getLine()}
			} else {
				tokens <- Token{tInvalid, string(t.peek(true)), t.getLine()}
			}
		}
	}
}

func (t *Tokenizer) peek(skipComment bool) rune {
	if t.isLast {
		return t.last
	}
	if len(t.str) == 0 {
		t.last = EOF
		return EOF
	}
	var size int
	t.last, size = utf8.DecodeRuneInString(t.str)

	if t.allowComments && skipComment {
		if t.last == '/' && len(t.str) > size {
			s, l := utf8.DecodeRuneInString(t.str[size:])
			if s == '/' {
				t.str = t.str[size+l:]
				for {
					s, l := utf8.DecodeRuneInString(t.str)
					if s != '\n' && s != '\r' {
						t.str = t.str[l:]
						if len(t.str) == 0 {
							return EOF
						}
					} else {
						break
					}
				}
				t.last, size = utf8.DecodeRuneInString(t.str)
			}
		}
	}

	t.isLast = true
	t.str = t.str[size:]
	return t.last
}

func (t *Tokenizer) consume(skipComment bool) {
	if !t.isLast {
		t.peek(skipComment)
	}
	t.isLast = false
}

func (t *Tokenizer) unread() {
	t.isLast = true
}

func (t *Tokenizer) next(skipComment bool) rune {
	n := t.peek(skipComment)
	t.consume(skipComment)
	return n
}
func (t *Tokenizer) read(valid func(c rune) bool) string {
	return t.readSkip(valid, true)
}

func (t *Tokenizer) readSkip(valid func(c rune) bool, skipComment bool) string {
	str := strings.Builder{}
	for {
		if c := t.next(skipComment); c != 0 && valid(c) {
			str.WriteRune(c)
		} else {
			t.unread()
			return str.String()
		}
	}
}
