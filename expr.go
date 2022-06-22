package lox

type exprVisitor interface {
	visitBinaryExpr(expr BinaryExpr)
	visitGroupingExpr(expr GroupingExpr)
	visitLiteralExpr(expr LiteralExpr)
	visitUnaryExpr(expr UnaryExpr)
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
	Value interface{}
}

type UnaryExpr struct {
	Operator Token
	Right    Expr
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
