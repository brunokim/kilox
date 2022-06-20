package lox

type exprVisitor interface {
	visitBinaryExpr(expr Binary)
	visitGroupingExpr(expr Grouping)
	visitLiteralExpr(expr Literal)
	visitUnaryExpr(expr Unary)
}

type Expr interface {
	accept(visitor exprVisitor)
}

// ----

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

type Grouping struct {
	Expression Expr
}

type Literal struct {
	Value interface{}
}

type Unary struct {
	Operator Token
	Right    Expr
}

// ----

func (expr Binary) accept(v exprVisitor) {
	v.visitBinaryExpr(expr)
}

func (expr Grouping) accept(v exprVisitor) {
	v.visitGroupingExpr(expr)
}

func (expr Literal) accept(v exprVisitor) {
	v.visitLiteralExpr(expr)
}

func (expr Unary) accept(v exprVisitor) {
	v.visitUnaryExpr(expr)
}
