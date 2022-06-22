package lox

type stmtVisitor interface {
	visitExpressionStmt(stmt ExpressionStmt)
	visitPrintStmt(expr PrintStmt)
}

type Stmt interface {
	accept(visitor stmtVisitor)
}

// ----

type ExpressionStmt struct {
	Expression Expr
}

type PrintStmt struct {
	Expression Expr
}

// ----

func (stmt ExpressionStmt) accept(v stmtVisitor) {
	v.visitExpressionStmt(stmt)
}

func (stmt PrintStmt) accept(v stmtVisitor) {
	v.visitPrintStmt(stmt)
}
