package lox

type stmtVisitor interface {
	visitExpressionStmt(stmt ExpressionStmt)
	visitPrintStmt(stmt PrintStmt)
	visitVarStmt(stmt VarStmt)
	visitIfStmt(stmt IfStmt)
	visitBlockStmt(stmt BlockStmt)
	visitWhileStmt(stmt WhileStmt)
	visitBreakStmt(stmt BreakStmt)
	visitContinueStmt(stmt ContinueStmt)
	visitFunctionStmt(stmt FunctionStmt)
	visitReturnStmt(stmt ReturnStmt)
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

type BreakStmt struct {
	Token Token
}

type ContinueStmt struct {
	Token Token
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

type ReturnStmt struct {
	Token  Token
	Result Expr
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

func (stmt BreakStmt) accept(v stmtVisitor) {
	v.visitBreakStmt(stmt)
}

func (stmt ContinueStmt) accept(v stmtVisitor) {
	v.visitContinueStmt(stmt)
}

func (stmt FunctionStmt) accept(v stmtVisitor) {
	v.visitFunctionStmt(stmt)
}

func (stmt ReturnStmt) accept(v stmtVisitor) {
	v.visitReturnStmt(stmt)
}
