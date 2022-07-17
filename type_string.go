package lox

import (
	"fmt"
	"strings"
)

type typePrinter struct {
	b *strings.Builder
}

func TypePrint(t Type) string {
	p := newPrinter()
	t.accept(p)
	return p.b.String()
}

func (t NilType) String() string      { return TypePrint(t) }
func (t BoolType) String() string     { return TypePrint(t) }
func (t NumberType) String() string   { return TypePrint(t) }
func (t StringType) String() string   { return TypePrint(t) }
func (t FunctionType) String() string { return TypePrint(t) }
func (t *RefType) String() string     { return TypePrint(t) }

// ----

func newPrinter() *typePrinter {
	return &typePrinter{
		b: new(strings.Builder),
	}
}

func (p *typePrinter) write(t Type) {
	t.accept(p)
}

func (p *typePrinter) visitNilType(t NilType) {
	p.b.WriteString("Nil")
}

func (p *typePrinter) visitBoolType(t BoolType) {
	p.b.WriteString("Bool")
}

func (p *typePrinter) visitNumberType(t NumberType) {
	p.b.WriteString("Number")
}

func (p *typePrinter) visitStringType(t StringType) {
	p.b.WriteString("String")
}

func (p *typePrinter) visitFunctionType(t FunctionType) {
	p.b.WriteRune('(')
	for i, param := range t.Params {
		p.write(param)
		if i < len(t.Params)-1 {
			p.b.WriteString(", ")
		}
	}
	p.b.WriteString(") -> ")
	p.write(t.Return)
}

func (p *typePrinter) visitRefType(x *RefType) {
	if x.Value == nil {
		fmt.Fprintf(p.b, "_%d", x.id)
	} else {
		p.b.WriteRune('&')
		p.write(x.Value)
	}
}
