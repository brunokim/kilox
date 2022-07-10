// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/expr.spec -dest expr.go
package lox

type Expr interface {
	accept(v exprVisitor)
}

type exprVisitor interface {
	visitBinaryExpr(e BinaryExpr)
	visitGroupingExpr(e GroupingExpr)
	visitLiteralExpr(e LiteralExpr)
	visitUnaryExpr(e UnaryExpr)
	visitVariableExpr(e VariableExpr)
	visitAssignmentExpr(e AssignmentExpr)
	visitLogicExpr(e LogicExpr)
	visitCallExpr(e CallExpr)
	visitFunctionExpr(e FunctionExpr)
	visitGetExpr(e GetExpr)
}

type BinaryExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

type GroupingExpr struct {
	Expression Expr
}

type LiteralExpr struct {
	Token Token
	Value any
}

type UnaryExpr struct {
	Operator Token
	Right    Expr
}

type VariableExpr struct {
	Name Token
}

type AssignmentExpr struct {
	Name  Token
	Value Expr
}

type LogicExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

type CallExpr struct {
	Callee Expr
	Paren  Token
	Args   []Expr
}

type FunctionExpr struct {
	Keyword Token
	Params  []Token
	Body    []Stmt
}

type GetExpr struct {
	Object Expr
	Name   Token
}

func (e BinaryExpr) accept(v exprVisitor) {
	v.visitBinaryExpr(e)
}

func (e GroupingExpr) accept(v exprVisitor) {
	v.visitGroupingExpr(e)
}

func (e LiteralExpr) accept(v exprVisitor) {
	v.visitLiteralExpr(e)
}

func (e UnaryExpr) accept(v exprVisitor) {
	v.visitUnaryExpr(e)
}

func (e VariableExpr) accept(v exprVisitor) {
	v.visitVariableExpr(e)
}

func (e AssignmentExpr) accept(v exprVisitor) {
	v.visitAssignmentExpr(e)
}

func (e LogicExpr) accept(v exprVisitor) {
	v.visitLogicExpr(e)
}

func (e CallExpr) accept(v exprVisitor) {
	v.visitCallExpr(e)
}

func (e FunctionExpr) accept(v exprVisitor) {
	v.visitFunctionExpr(e)
}

func (e GetExpr) accept(v exprVisitor) {
	v.visitGetExpr(e)
}
