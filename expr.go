package lox

type exprVisitor interface {
	visitBinaryExpr(expr BinaryExpr)
	visitGroupingExpr(expr GroupingExpr)
	visitLiteralExpr(expr LiteralExpr)
	visitUnaryExpr(expr UnaryExpr)
	visitVariableExpr(expr VariableExpr)
	visitAssignmentExpr(expr AssignmentExpr)
	visitLogicExpr(expr LogicExpr)
	visitCallExpr(expr CallExpr)
	visitFunctionExpr(expr FunctionExpr)
}

type Expr interface {
	accept(visitor exprVisitor)
}

// ----

type BinaryExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

type GroupingExpr struct {
	Expression Expr
}

type LiteralExpr struct {
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
	Params []Token
	Body   []Stmt
}

// ----

func (expr BinaryExpr) accept(v exprVisitor) {
	v.visitBinaryExpr(expr)
}

func (expr GroupingExpr) accept(v exprVisitor) {
	v.visitGroupingExpr(expr)
}

func (expr LiteralExpr) accept(v exprVisitor) {
	v.visitLiteralExpr(expr)
}

func (expr UnaryExpr) accept(v exprVisitor) {
	v.visitUnaryExpr(expr)
}

func (expr VariableExpr) accept(v exprVisitor) {
	v.visitVariableExpr(expr)
}

func (expr AssignmentExpr) accept(v exprVisitor) {
	v.visitAssignmentExpr(expr)
}

func (expr LogicExpr) accept(v exprVisitor) {
	v.visitLogicExpr(expr)
}

func (expr CallExpr) accept(v exprVisitor) {
	v.visitCallExpr(expr)
}

func (expr FunctionExpr) accept(v exprVisitor) {
	v.visitFunctionExpr(expr)
}
