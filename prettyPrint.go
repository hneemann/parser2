package parser2

import (
	"bytes"
	"fmt"
)

func PrettyPrint[V any](ast AST) string {
	var buf bytes.Buffer
	prettyPrintAST[V](&buf, ast, "")
	return buf.String()
}

func prettyPrintAST[V any](buf *bytes.Buffer, ast AST, indent string) {
	switch e := ast.(type) {
	case *Const[V]:
		buf.WriteString(fmt.Sprint(e.Value))
	case *Ident:
		buf.WriteString(e.Name)
	case *MapAccess:
		prettyPrintAST[V](buf, e.MapValue, indent)
		buf.WriteString(".")
		buf.WriteString(e.Key)
	case *ListAccess:
		prettyPrintAST[V](buf, e.List, indent)
		buf.WriteString("[")
		prettyPrintAST[V](buf, e.Index, indent)
		buf.WriteString("]")
	case *Unary:
		buf.WriteString(e.Operator)
		prettyPrintAST[V](buf, e.Value, indent)
	case *FunctionCall:
		prettyPrintAST[V](buf, e.Func, indent)
		buf.WriteString("(")
		for i, arg := range e.Args {
			if i > 0 {
				buf.WriteString(", ")
			}
			prettyPrintAST[V](buf, arg, indent)
		}
		buf.WriteString(")")
	case *Operate:
		if io, ok := e.A.(*Operate); ok && io.Priority < e.Priority {
			buf.WriteString("(")
			prettyPrintAST[V](buf, e.A, indent)
			buf.WriteString(")")
		} else {
			prettyPrintAST[V](buf, e.A, indent)
		}
		buf.WriteString(e.Operator)
		if io, ok := e.B.(*Operate); ok && io.Priority < e.Priority {
			buf.WriteString("(")
			prettyPrintAST[V](buf, e.B, indent)
			buf.WriteString(")")
		} else {
			prettyPrintAST[V](buf, e.B, indent)
		}
	case *ListLiteral:
		buf.WriteString("[")
		for i, item := range e.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			prettyPrintAST[V](buf, item, indent)
		}
		buf.WriteString("]")
	case *MapLiteral:
		buf.WriteString("{")
		i := 0
		for k, v := range e.Map.Iter {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(k + ":")
			prettyPrintAST[V](buf, v, indent)
			i++
		}
		buf.WriteString("}")
	case *ClosureLiteral:
		if len(e.Names) == 1 {
			buf.WriteString(e.Names[0])
		} else {
			buf.WriteString("(")
			for i, n := range e.Names {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(n)
			}
			buf.WriteString(")")
		}
		buf.WriteString(" -> ")
		prettyPrintAST[V](buf, e.Func, indent)
	case *MethodCall:
		if methodParenthesesNeeded(e.Value) {
			buf.WriteString("(")
			prettyPrintAST[V](buf, e.Value, indent)
			buf.WriteString(")")
		} else {
			prettyPrintAST[V](buf, e.Value, indent)
		}
		buf.WriteString("\n" + indent + "  ." + e.Name + "(")
		for i, arg := range e.Args {
			if i > 0 {
				buf.WriteString(", ")
			}
			prettyPrintAST[V](buf, arg, indent+"  ")
		}
		buf.WriteString(")")
	case *Let:
		if cl, ok := e.Value.(*ClosureLiteral); ok {
			if cl.ThisName != "" {
				buf.WriteString(indent + "func " + cl.ThisName + "(")
				for i, n := range cl.Names {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(n)
				}
				buf.WriteString(")\n  " + indent)
				prettyPrintAST[V](buf, cl.Func, indent+"  ")
				if indent == "" {
					buf.WriteString(";\n\n" + indent)
				} else {
					buf.WriteString(";\n" + indent)
				}
				prettyPrintAST[V](buf, e.Inner, indent)
				return
			}
		}
		buf.WriteString("let " + e.Name + " = ")
		prettyPrintAST[V](buf, e.Value, indent)
		if indent == "" {
			buf.WriteString(";\n\n" + indent)
		} else {
			buf.WriteString(";\n" + indent)
		}
		prettyPrintAST[V](buf, e.Inner, indent)
	case *If:
		buf.WriteString("if ")
		prettyPrintAST[V](buf, e.Cond, indent+"  ")
		buf.WriteString("\n" + indent + "then\n" + indent + "  ")
		prettyPrintAST[V](buf, e.Then, indent+"  ")
		buf.WriteString("\n" + indent + "else\n" + indent + "  ")
		prettyPrintAST[V](buf, e.Else, indent+"  ")
	case *Switch[V]:
		buf.WriteString("switch ")
		prettyPrintAST[V](buf, e.SwitchValue, indent+"  ")
		for _, cs := range e.Cases {
			buf.WriteString("\n" + indent + "case ")
			prettyPrintAST[V](buf, cs.CaseConst, indent+"  ")
			buf.WriteString(":\n" + indent + "  ")
			prettyPrintAST[V](buf, cs.Value, indent+"  ")
		}
		buf.WriteString("\n" + indent + "default\n" + indent + "  ")
		prettyPrintAST[V](buf, e.Default, indent+"  ")
	case *TryCatch:
		buf.WriteString("try ")
		prettyPrintAST[V](buf, e.Try, indent)
		buf.WriteString(" catch ")
		prettyPrintAST[V](buf, e.Catch, indent)
	default:
		buf.WriteString("<unknown AST node>")
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
