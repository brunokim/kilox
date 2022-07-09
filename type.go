// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/type.spec -dest type.go
package lox

type Type interface {
	accept(v typeVisitor)
}

type typeVisitor interface {
	visitNilType(t NilType)
	visitBoolType(t BoolType)
	visitNumberType(t NumberType)
	visitStringType(t StringType)
	visitFunctionType(t FunctionType)
	visitRefType(t *RefType)
	visitUnionType(t *UnionType)
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
	id    int
}

type UnionType struct {
	Types []Type
}

func (t NilType) accept(v typeVisitor) {
	v.visitNilType(t)
}

func (t BoolType) accept(v typeVisitor) {
	v.visitBoolType(t)
}

func (t NumberType) accept(v typeVisitor) {
	v.visitNumberType(t)
}

func (t StringType) accept(v typeVisitor) {
	v.visitStringType(t)
}

func (t FunctionType) accept(v typeVisitor) {
	v.visitFunctionType(t)
}

func (t *RefType) accept(v typeVisitor) {
	v.visitRefType(t)
}

func (t *UnionType) accept(v typeVisitor) {
	v.visitUnionType(t)
}
