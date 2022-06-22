package lox

type stmtVisitor interface {
	visitExpressionStmt(stmt ExpressionStmt)
	visitPrintStmt(stmt PrintStmt)
	visitVarStmt(stmt VarStmt)
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

type VarStmt struct {
	Name Token
	Init Expr
}

// ----

func (stmt ExpressionStmt) accept(v stmtVisitor) {
	v.visitExpressionStmt(stmt)
}

func (stmt PrintStmt) accept(v stmtVisitor) {
	v.visitPrintStmt(stmt)
}

func (stmt VarStmt) accept(v stmtVisitor) {
	v.visitVarStmt(stmt)
}
