// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/type.spec -dest type.go
package lox

type Type interface {
	Accept(v typeVisitor)
}

type typeVisitor interface {
	visitNilType(t NilType)
	visitBoolType(t BoolType)
	visitNumberType(t NumberType)
	visitStringType(t StringType)
	visitFunctionType(t FunctionType)
	visitRefType(t *RefType)
}

type NilType struct {
	Token Token
}

type BoolType struct {
	Token Token
}

type NumberType struct {
	Token Token
}

type StringType struct {
	Token Token
}

type FunctionType struct {
	Params []Type
	Return Type
}

type RefType struct {
	Value       Type
	id          int
	constraints []Constraint
}

func (t NilType) Accept(v typeVisitor) {
	v.visitNilType(t)
}

func (t BoolType) Accept(v typeVisitor) {
	v.visitBoolType(t)
}

func (t NumberType) Accept(v typeVisitor) {
	v.visitNumberType(t)
}

func (t StringType) Accept(v typeVisitor) {
	v.visitStringType(t)
}

func (t FunctionType) Accept(v typeVisitor) {
	v.visitFunctionType(t)
}

func (t *RefType) Accept(v typeVisitor) {
	v.visitRefType(t)
}
