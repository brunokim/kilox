package lox

type stmtVisitor interface {
	visitExpressionStmt(stmt ExpressionStmt)
	visitPrintStmt(stmt PrintStmt)
	visitVarStmt(stmt VarStmt)
	visitIfStmt(stmt IfStmt)
	visitBlockStmt(stmt BlockStmt)
	visitWhileStmt(stmt WhileStmt)
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

type IfStmt struct {
	Condition Expr
	Then      Stmt
	Else      Stmt
}

type BlockStmt struct {
	Statements []Stmt
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
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

func (stmt IfStmt) accept(v stmtVisitor) {
	v.visitIfStmt(stmt)
}

func (stmt BlockStmt) accept(v stmtVisitor) {
	v.visitBlockStmt(stmt)
}

func (stmt WhileStmt) accept(v stmtVisitor) {
	v.visitWhileStmt(stmt)
}
