package xmlWriter

import (
	"bytes"
	"log"
)

func New() *XMLWriter {
	return NewWithBuffer(new(bytes.Buffer))
}

func NewWithBuffer(b *bytes.Buffer) *XMLWriter {
	return &XMLWriter{b: b, depth: -1}
}

type XMLWriter struct {
	b          *bytes.Buffer
	open       []string
	depth      int
	inLine     bool
	tagIsOpen  bool
	avoidShort bool
}

func (w *XMLWriter) AvoidShort() *XMLWriter {
	w.avoidShort = true
	return w
}

func (w *XMLWriter) Open(tag string) *XMLWriter {
	w.checkOpenTag()
	w.newLine()
	w.depth++
	w.checkIndent()
	w.write("<")
	w.write(tag)
	w.open = append(w.open, tag)
	w.tagIsOpen = true
	w.inLine = true
	return w
}

func (w *XMLWriter) Attr(key, value string) *XMLWriter {
	if w.tagIsOpen {
		w.b.WriteString(" ")
		w.b.WriteString(key)
		w.b.WriteString("=\"")
		w.writeEsc(value)
		w.b.WriteString("\"")
	} else {
		log.Print("tag is not open")
	}
	return w
}

func (w *XMLWriter) Close() *XMLWriter {
	if w.tagIsOpen && !w.avoidShort {
		w.b.WriteString("/>")
		w.tagIsOpen = false
		w.depth--
	} else {
		w.checkIndent()
		w.depth--
		w.write("</")
		w.write(w.open[len(w.open)-1])
		w.write(">")
	}
	w.open = w.open[:len(w.open)-1]
	w.newLine()
	return w
}

func (w *XMLWriter) write(s string) {
	w.checkOpenTag()
	w.checkIndent()
	w.b.WriteString(s)
}

func (w *XMLWriter) Write(s string) *XMLWriter {
	w.checkOpenTag()
	w.checkIndent()
	w.writeEsc(s)
	return w
}

func (w *XMLWriter) writeEsc(s string) {
	for _, r := range s {
		switch r {
		case '\'':
			w.b.WriteString("&apos;")
		case '"':
			w.b.WriteString("&quot;")
		case '<':
			w.b.WriteString("&lt;")
		case '>':
			w.b.WriteString("&gt;")
		case '&':
			w.b.WriteString("&amp;")
		default:
			w.b.WriteRune(r)
		}
	}
}

func (w *XMLWriter) checkIndent() {
	if !w.inLine {
		for i := 0; i < w.depth; i++ {
			w.b.WriteString("\t")
		}
		w.inLine = true
	}
}

func (w *XMLWriter) newLine() {
	if w.inLine {
		w.b.WriteRune('\n')
		w.inLine = false
	}
}

func (w *XMLWriter) String() string {
	return w.b.String()
}

func (w *XMLWriter) checkOpenTag() {
	if w.tagIsOpen {
		w.b.WriteRune('>')
		w.tagIsOpen = false
	}
}
