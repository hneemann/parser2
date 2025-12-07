package parser2

import (
	"bytes"
	"fmt"
)

func PrettyPrint[V any](ast AST) string {
	w := newWriter()
	prettyPrintAST[V](w, ast)
	return w.String()
}

func prettyPrintAST[V any](buf *writer, ast AST) {
	switch e := ast.(type) {
	case *Const[V]:
		buf.writeString(fmt.Sprint(e.Value))
	case *Ident:
		buf.writeString(e.Name)
	case *MapAccess:
		prettyPrintAST[V](buf, e.MapValue)
		buf.writeString(".")
		buf.writeString(e.Key)
	case *ListAccess:
		prettyPrintAST[V](buf, e.List)
		buf.writeString("[")
		prettyPrintAST[V](buf, e.Index)
		buf.writeString("]")
	case *Unary:
		buf.writeString(e.Operator)
		prettyPrintAST[V](buf, e.Value)
	case *FunctionCall:
		prettyPrintAST[V](buf, e.Func)
		writeArgs[V](buf, e.Args)
	case *Operate:
		if io, ok := e.A.(*Operate); ok && io.Priority < e.Priority {
			buf.writeString("(")
			prettyPrintAST[V](buf, e.A)
			buf.writeString(")")
		} else {
			prettyPrintAST[V](buf, e.A)
		}
		buf.writeString(e.Operator)
		if io, ok := e.B.(*Operate); ok && io.Priority < e.Priority {
			buf.writeString("(")
			prettyPrintAST[V](buf, e.B)
			buf.writeString(")")
		} else {
			prettyPrintAST[V](buf, e.B)
		}
	case *ListLiteral:
		buf.writeString("[")
		for i, item := range e.List {
			if i > 0 {
				buf.writeString(", ")
			}
			prettyPrintAST[V](buf, item)
		}
		buf.writeString("]")
	case *MapLiteral:
		buf.writeString("{")
		ib := buf.down()
		i := 0
		for k, v := range e.Map.Iter {
			if i > 0 {
				ib.writeString(",")
				ib.newLine()
			}
			ib.writeString(k + ": ")
			prettyPrintAST[V](ib, v)
			i++
		}
		ib.writeString("}")
	case *ClosureLiteral:
		if len(e.Names) == 1 {
			buf.writeString(e.Names[0])
		} else {
			buf.writeString("(")
			for i, n := range e.Names {
				if i > 0 {
					buf.writeString(", ")
				}
				buf.writeString(n)
			}
			buf.writeString(")")
		}
		buf.writeString(" -> ")
		prettyPrintAST[V](buf.down(), e.Func)
	case *MethodCall:
		do := buf.down()
		if methodParenthesesNeeded(e.Value) {
			do.writeString("(")
			prettyPrintAST[V](do, e.Value)
			do.writeString(")")
		} else {
			prettyPrintAST[V](do, e.Value)
		}
		do.newLine()
		do.writeString(" ." + e.Name)
		writeArgs[V](do, e.Args)
	case *Let:
		if cl, ok := e.Value.(*ClosureLiteral); ok && cl.ThisName != "" {
			buf.writeString("func " + cl.ThisName + "(")
			for i, n := range cl.Names {
				if i > 0 {
					buf.writeString(", ")
				}
				buf.writeString(n)
			}
			buf.writeString(")")
			ib := buf.indent()
			ib.newLine()
			prettyPrintAST[V](ib, cl.Func)
			ib.writeString(";")
		} else {
			buf.writeString("let " + e.Name + " = ")
			prettyPrintAST[V](buf, e.Value)
			buf.writeString(";")
		}
		if buf.tab == 0 {
			buf.writeString("\n")
		}
		buf.newLine()
		prettyPrintAST[V](buf, e.Inner)
	case *If:
		do := buf.down()
		do.writeString("if ")
		prettyPrintAST[V](do, e.Cond)
		do.newLine()
		do.writeString("then ")
		prettyPrintAST[V](do, e.Then)
		do.newLine()
		do.writeString("else ")
		prettyPrintAST[V](do, e.Else)
	case *Switch[V]:
		do := buf.down()
		do.writeString("switch ")
		do = do.indent()
		prettyPrintAST[V](do, e.SwitchValue)
		for _, cs := range e.Cases {
			do.newLine()
			do.writeString("case ")
			prettyPrintAST[V](do, cs.CaseConst)
			do.writeString(": ")
			prettyPrintAST[V](do, cs.Value)
		}
		do.newLine()
		do.writeString("default ")
		prettyPrintAST[V](do, e.Default)
	case *TryCatch:
		buf.writeString("try ")
		prettyPrintAST[V](buf, e.Try)
		buf.writeString(" catch ")
		prettyPrintAST[V](buf, e.Catch)
	default:
		buf.writeString("<unknown AST node>")
	}
}

func writeArgs[V any](buf *writer, args []AST) {
	const maxCmplx = 6
	cmplx := 0
	for _, arg := range args {
		arg.Traverse(VisitorFunc(func(ast AST) bool {
			switch ast.(type) {
			case *Let, *If, *Switch[V], *MethodCall:
				cmplx += 2
			case *TryCatch, *FunctionCall:
				cmplx++
			}
			return cmplx < maxCmplx
		}))
	}
	if cmplx < maxCmplx {
		buf.writeString("(")
		for i, arg := range args {
			if i > 0 {
				buf.writeString(", ")
			}
			prettyPrintAST[V](buf, arg)
		}
		buf.writeString(")")
	} else {
		buf.writeString("(")
		do := buf.down()
		for i, arg := range args {
			if i > 0 {
				do.writeString(",")
				do.newLine()
			}
			prettyPrintAST[V](do, arg)
		}
		do.newLine()
		do.writeString(")")
	}
}

func methodParenthesesNeeded(value AST) bool {
	if _, ok := value.(*ClosureLiteral); ok {
		return true
	}
	if _, ok := value.(*Operate); ok {
		return true
	}
	return false
}

type posWriter struct {
	buf bytes.Buffer
	col int
}

func (w *posWriter) writeString(s string) {
	w.buf.WriteString(s)
	w.col += len(s)
}

func (w *posWriter) newLine() {
	w.buf.WriteRune('\n')
	w.col = 0
}

func (w *posWriter) String() string {
	return w.buf.String()
}

type writer struct {
	pw  *posWriter
	tab int
}

func newWriter() *writer {
	return &writer{
		pw:  &posWriter{},
		tab: 0,
	}
}

func (w *writer) down() *writer {
	return &writer{
		pw:  w.pw,
		tab: w.pw.col,
	}
}

func (w *writer) indent() *writer {
	return &writer{
		pw:  w.pw,
		tab: w.tab + 2,
	}
}

func (w *writer) writeString(s string) {
	w.pw.writeString(s)
}

func (w *writer) writeIndent() {
	for i := 0; i < w.tab; i++ {
		w.writeString(" ")
	}
}

func (w *writer) newLine() {
	w.pw.newLine()
	w.writeIndent()
}

func (w *writer) String() string {
	return w.pw.String()
}
