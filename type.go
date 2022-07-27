// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/type.spec -dest type.go
package lox

type Type interface {
	Accept(v typeVisitor)
}

type typeVisitor interface {
	VisitNilType(t NilType)
	VisitBoolType(t BoolType)
	VisitNumberType(t NumberType)
	VisitStringType(t StringType)
	VisitFunctionType(t FunctionType)
	VisitRefType(t *RefType)
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
	Value Type
	ID    int
}

func (t NilType) Accept(v typeVisitor) {
	v.VisitNilType(t)
}

func (t BoolType) Accept(v typeVisitor) {
	v.VisitBoolType(t)
}

func (t NumberType) Accept(v typeVisitor) {
	v.VisitNumberType(t)
}

func (t StringType) Accept(v typeVisitor) {
	v.VisitStringType(t)
}

func (t FunctionType) Accept(v typeVisitor) {
	v.VisitFunctionType(t)
}

func (t *RefType) Accept(v typeVisitor) {
	v.VisitRefType(t)
}
