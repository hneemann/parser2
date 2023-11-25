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
	return "'" + t.image + "'"
}

type Tokenizer struct {
	str            string
	isLast         bool
	last           rune
	tok            chan Token
	tokenAvail     int
	token          [2]Token
	line           Line
	number         Matcher
	identifier     Matcher
	operatorDetect Detect
	textOperators  map[string]string
	allowComments  bool
}

type Matcher func(r rune) (func(r rune) bool, bool)

type Detect func(r rune) (Detect, bool)

func NewDetect(operators []string) Detect {
	if len(operators) == 1 && operators[0] == "" {
		return func(r rune) (Detect, bool) {
			return nil, true
		}
	}

	m := map[rune]*[]string{}
	endIsValid := false
	for _, op := range operators {
		r, n := utf8.DecodeRuneInString(op)
		if n == 0 {
			endIsValid = true
		} else {
			remainingOp := op[n:]
			l, ok := m[r]
			if !ok {
				l = &[]string{}
				m[r] = l
			}
			*l = append(*l, remainingOp)
		}
	}

	type runeListEntry struct {
		r rune
		f Detect
	}

	var rl []runeListEntry
	for r, l := range m {
		rl = append(rl, runeListEntry{r: r, f: NewDetect(*l)})
	}
	return func(r rune) (Detect, bool) {
		for _, rle := range rl {
			if rle.r == r {
				return rle.f, true
			}
		}
		return nil, endIsValid
	}
}

func NewTokenizer(text string, number, identifier Matcher, operatorDetect Detect, textOp map[string]string, allowComments bool) *Tokenizer {
	t := make(chan Token)
	tok := &Tokenizer{
		str:            text,
		textOperators:  textOp,
		number:         number,
		identifier:     identifier,
		operatorDetect: operatorDetect,
		allowComments:  allowComments,
		line:           1,
		tok:            t}
	go tok.run(t)
	return tok
}

func (t *Tokenizer) Peek() Token {
	return t.forward(1)
}

func (t *Tokenizer) PeekPeek() Token {
	return t.forward(2)
}

func (t *Tokenizer) forward(i int) Token {
	for t.tokenAvail < i {
		var ok bool
		t.token[t.tokenAvail], ok = <-t.tok
		if ok {
			t.tokenAvail++
		} else {
			return TokenEof
		}
	}
	return t.token[i-1]
}

func (t *Tokenizer) Next() Token {
	switch t.tokenAvail {
	case 2:
		to := t.token[0]
		t.token[0] = t.token[1]
		t.tokenAvail--
		return to
	case 1:
		t.tokenAvail--
		return t.token[0]
	default:
		to, ok := <-t.tok
		if ok {
			return to
		} else {
			return TokenEof
		}
	}
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
			tokens <- t.readStr()
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
			} else {
				t.unread()
				if op, ok := t.parseOperator(); ok {
					tokens <- Token{tOperate, op, t.getLine()}
				} else {
					tokens <- Token{tInvalid, op, t.getLine()}
				}
			}
		}
	}
}

func (t *Tokenizer) parseOperator() (string, bool) {
	r := t.next(false)
	if d, _ := t.operatorDetect(r); d != nil {
		op := string(r)
		for {
			r = t.next(false)
			var ok bool
			if d, ok = d(r); d != nil {
				op += string(r)
			} else {
				t.unread()
				return op, ok
			}
		}
	} else {
		return string(r), false
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

func (t *Tokenizer) readStr() Token {
	str := strings.Builder{}
	for {
		if c := t.next(false); c != '"' {
			switch c {
			case 0, '\n', '\r':
				return Token{tInvalid, "EOL", t.getLine()}
			case '\\':
				i := t.next(false)
				switch i {
				case 'n':
					str.WriteRune('\n')
				case 'r':
					str.WriteRune('\r')
				case 't':
					str.WriteRune('\t')
				case '"':
					str.WriteRune('"')
				case '\\':
					str.WriteRune('\\')
				default:
					return Token{tInvalid, fmt.Sprintf("Escape %c", i), t.getLine()}
				}
			default:
				str.WriteRune(c)
			}
		} else {
			return Token{tString, str.String(), t.getLine()}
		}
	}
}
