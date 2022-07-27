// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/stmt.spec -dest stmt.go
package lox

type Stmt interface {
	Accept(v stmtVisitor)
}

type stmtVisitor interface {
	VisitExpressionStmt(s ExpressionStmt)
	VisitPrintStmt(s PrintStmt)
	VisitVarStmt(s VarStmt)
	VisitIfStmt(s IfStmt)
	VisitBlockStmt(s BlockStmt)
	VisitLoopStmt(s LoopStmt)
	VisitBreakStmt(s BreakStmt)
	VisitContinueStmt(s ContinueStmt)
	VisitFunctionStmt(s FunctionStmt)
	VisitReturnStmt(s ReturnStmt)
	VisitClassStmt(s ClassStmt)
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
	v.VisitExpressionStmt(s)
}

func (s PrintStmt) Accept(v stmtVisitor) {
	v.VisitPrintStmt(s)
}

func (s VarStmt) Accept(v stmtVisitor) {
	v.VisitVarStmt(s)
}

func (s IfStmt) Accept(v stmtVisitor) {
	v.VisitIfStmt(s)
}

func (s BlockStmt) Accept(v stmtVisitor) {
	v.VisitBlockStmt(s)
}

func (s LoopStmt) Accept(v stmtVisitor) {
	v.VisitLoopStmt(s)
}

func (s BreakStmt) Accept(v stmtVisitor) {
	v.VisitBreakStmt(s)
}

func (s ContinueStmt) Accept(v stmtVisitor) {
	v.VisitContinueStmt(s)
}

func (s FunctionStmt) Accept(v stmtVisitor) {
	v.VisitFunctionStmt(s)
}

func (s ReturnStmt) Accept(v stmtVisitor) {
	v.VisitReturnStmt(s)
}

func (s ClassStmt) Accept(v stmtVisitor) {
	v.VisitClassStmt(s)
}
