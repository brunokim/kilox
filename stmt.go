// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/stmt.spec -dest stmt.go
package lox

type Stmt interface {
	Accept(v stmtVisitor)
}

type stmtVisitor interface {
	visitExpressionStmt(s ExpressionStmt)
	visitPrintStmt(s PrintStmt)
	visitVarStmt(s VarStmt)
	visitIfStmt(s IfStmt)
	visitBlockStmt(s BlockStmt)
	visitLoopStmt(s LoopStmt)
	visitBreakStmt(s BreakStmt)
	visitContinueStmt(s ContinueStmt)
	visitFunctionStmt(s FunctionStmt)
	visitReturnStmt(s ReturnStmt)
	visitClassStmt(s ClassStmt)
}

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

type LoopStmt struct {
	Condition Expr
	Body      Stmt
	OnLoop    Expr
}

type BreakStmt struct {
	Keyword Token
}

type ContinueStmt struct {
	Keyword Token
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

type ReturnStmt struct {
	Keyword Token
	Result  Expr
}

type ClassStmt struct {
	Name          Token
	Methods       []FunctionStmt
	Vars          []VarStmt
	StaticMethods []FunctionStmt
	StaticVars    []VarStmt
}

func (s ExpressionStmt) Accept(v stmtVisitor) {
	v.visitExpressionStmt(s)
}

func (s PrintStmt) Accept(v stmtVisitor) {
	v.visitPrintStmt(s)
}

func (s VarStmt) Accept(v stmtVisitor) {
	v.visitVarStmt(s)
}

func (s IfStmt) Accept(v stmtVisitor) {
	v.visitIfStmt(s)
}

func (s BlockStmt) Accept(v stmtVisitor) {
	v.visitBlockStmt(s)
}

func (s LoopStmt) Accept(v stmtVisitor) {
	v.visitLoopStmt(s)
}

func (s BreakStmt) Accept(v stmtVisitor) {
	v.visitBreakStmt(s)
}

func (s ContinueStmt) Accept(v stmtVisitor) {
	v.visitContinueStmt(s)
}

func (s FunctionStmt) Accept(v stmtVisitor) {
	v.visitFunctionStmt(s)
}

func (s ReturnStmt) Accept(v stmtVisitor) {
	v.visitReturnStmt(s)
}

func (s ClassStmt) Accept(v stmtVisitor) {
	v.visitClassStmt(s)
}
