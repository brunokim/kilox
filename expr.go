// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/expr.spec -dest expr.go -extensions typename
package lox

type Expr interface {
	Accept(v exprVisitor)
	TypeName() string
}

type exprVisitor interface {
	visitBinaryExpr(e *BinaryExpr)
	visitGroupingExpr(e *GroupingExpr)
	visitLiteralExpr(e *LiteralExpr)
	visitUnaryExpr(e *UnaryExpr)
	visitVariableExpr(e *VariableExpr)
	visitAssignmentExpr(e *AssignmentExpr)
	visitLogicExpr(e *LogicExpr)
	visitCallExpr(e *CallExpr)
	visitFunctionExpr(e *FunctionExpr)
	visitGetExpr(e *GetExpr)
	visitSetExpr(e *SetExpr)
	visitThisExpr(e *ThisExpr)
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

type SetExpr struct {
	Object Expr
	Name   Token
	Value  Expr
}

type ThisExpr struct {
	Keyword Token
}

func (e *BinaryExpr) Accept(v exprVisitor) {
	v.visitBinaryExpr(e)
}

func (e *GroupingExpr) Accept(v exprVisitor) {
	v.visitGroupingExpr(e)
}

func (e *LiteralExpr) Accept(v exprVisitor) {
	v.visitLiteralExpr(e)
}

func (e *UnaryExpr) Accept(v exprVisitor) {
	v.visitUnaryExpr(e)
}

func (e *VariableExpr) Accept(v exprVisitor) {
	v.visitVariableExpr(e)
}

func (e *AssignmentExpr) Accept(v exprVisitor) {
	v.visitAssignmentExpr(e)
}

func (e *LogicExpr) Accept(v exprVisitor) {
	v.visitLogicExpr(e)
}

func (e *CallExpr) Accept(v exprVisitor) {
	v.visitCallExpr(e)
}

func (e *FunctionExpr) Accept(v exprVisitor) {
	v.visitFunctionExpr(e)
}

func (e *GetExpr) Accept(v exprVisitor) {
	v.visitGetExpr(e)
}

func (e *SetExpr) Accept(v exprVisitor) {
	v.visitSetExpr(e)
}

func (e *ThisExpr) Accept(v exprVisitor) {
	v.visitThisExpr(e)
}

func (*BinaryExpr) TypeName() string     { return "binary" }
func (*GroupingExpr) TypeName() string   { return "grouping" }
func (*LiteralExpr) TypeName() string    { return "literal" }
func (*UnaryExpr) TypeName() string      { return "unary" }
func (*VariableExpr) TypeName() string   { return "variable" }
func (*AssignmentExpr) TypeName() string { return "assignment" }
func (*LogicExpr) TypeName() string      { return "logic" }
func (*CallExpr) TypeName() string       { return "call" }
func (*FunctionExpr) TypeName() string   { return "function" }
func (*GetExpr) TypeName() string        { return "get" }
func (*SetExpr) TypeName() string        { return "set" }
func (*ThisExpr) TypeName() string       { return "this" }
