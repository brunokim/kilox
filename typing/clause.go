package typing

import (
	"github.com/brunokim/lox"
	"github.com/brunokim/lox/errlist"
)

type TypeClause struct {
	Head lox.FunctionType
	Body []lox.FunctionType
}

// ---- Error

type logicError struct{}

func (err logicError) Error() string {
	return "logic error"
}

// ---- Scope

type scope struct {
	m         *logicModel
	enclosing *scope
	clause    *TypeClause
	hasReturn bool

	refs map[string]*lox.RefType
}

func newScope(m *logicModel) *scope {
	return &scope{
		m:         m,
		enclosing: m.scope,
		refs:      make(map[string]*lox.RefType),
	}
}

func (s *scope) ref(name string) *lox.RefType {
	if name == "_" {
		return s.m.newRef()
	}
	if _, ok := s.refs[name]; !ok {
		s.refs[name] = s.m.newRef()
	}
	return s.refs[name]
}

// ---- Logic model

type logicModel struct {
	clauses []TypeClause
	errors  []logicError
	scope   *scope

	currType lox.Type

	refID int
}

func newLogicModel() *logicModel {
	return &logicModel{}
}

func BuildClauses(stmts []lox.Stmt) ([]TypeClause, error) {
	m := newLogicModel()
	m.visitStmts(stmts)
	if len(m.errors) > 0 {
		return nil, errlist.Of[logicError](m.errors)
	}
	return m.clauses, nil
}

// ----

func (m *logicModel) visitStmts(stmts []lox.Stmt) {
	for _, stmt := range stmts {
		m.visitStmt(stmt)
	}
}

func (m *logicModel) visitStmt(stmt lox.Stmt) {
	stmt.Accept(m)
}

func (m *logicModel) visitExpr(expr lox.Expr) lox.Type {
	expr.Accept(m)
	return m.currType
}

func (m *logicModel) newRef() *lox.RefType {
	m.refID++
	return &lox.RefType{ID: m.refID}
}

// ---- Expr

func (m *logicModel) VisitBinaryExpr(e *lox.BinaryExpr) {
	panic("typing.(*logicModel).VisitBinaryExpr is not implemented")
}

func (m *logicModel) VisitGroupingExpr(e *lox.GroupingExpr) {
	panic("typing.(*logicModel).VisitGroupingExpr is not implemented")
}

func (m *logicModel) VisitLiteralExpr(e *lox.LiteralExpr) {
	panic("typing.(*logicModel).VisitLiteralExpr is not implemented")
}

func (m *logicModel) VisitUnaryExpr(e *lox.UnaryExpr) {
	panic("typing.(*logicModel).VisitUnaryExpr is not implemented")
}

func (m *logicModel) VisitVariableExpr(e *lox.VariableExpr) {
	panic("typing.(*logicModel).VisitVariableExpr is not implemented")
}

func (m *logicModel) VisitAssignmentExpr(e *lox.AssignmentExpr) {
	panic("typing.(*logicModel).VisitAssignmentExpr is not implemented")
}

func (m *logicModel) VisitLogicExpr(e *lox.LogicExpr) {
	panic("typing.(*logicModel).VisitLogicExpr is not implemented")
}

func (m *logicModel) VisitCallExpr(e *lox.CallExpr) {
	panic("typing.(*logicModel).VisitCallExpr is not implemented")
}

func (m *logicModel) VisitFunctionExpr(e *lox.FunctionExpr) {
	panic("typing.(*logicModel).VisitFunctionExpr is not implemented")
}

func (m *logicModel) VisitGetExpr(e *lox.GetExpr) {
	panic("typing.(*logicModel).VisitGetExpr is not implemented")
}

func (m *logicModel) VisitSetExpr(e *lox.SetExpr) {
	panic("typing.(*logicModel).VisitSetExpr is not implemented")
}

func (m *logicModel) VisitThisExpr(e *lox.ThisExpr) {
	panic("typing.(*logicModel).VisitThisExpr is not implemented")
}

// ---- Stmt

func (m *logicModel) VisitExpressionStmt(s lox.ExpressionStmt) {
	panic("typing.(*logicModel).VisitExpressionStmt is not implemented")
}

func (m *logicModel) VisitPrintStmt(s lox.PrintStmt) {
	panic("typing.(*logicModel).VisitPrintStmt is not implemented")
}

func (m *logicModel) VisitVarStmt(s lox.VarStmt) {
	panic("typing.(*logicModel).VisitVarStmt is not implemented")
}

func (m *logicModel) VisitIfStmt(s lox.IfStmt) {
	panic("typing.(*logicModel).VisitIfStmt is not implemented")
}

func (m *logicModel) VisitBlockStmt(s lox.BlockStmt) {
	panic("typing.(*logicModel).VisitBlockStmt is not implemented")
}

func (m *logicModel) VisitLoopStmt(s lox.LoopStmt) {
	panic("typing.(*logicModel).VisitLoopStmt is not implemented")
}

func (m *logicModel) VisitBreakStmt(s lox.BreakStmt) {
	panic("typing.(*logicModel).VisitBreakStmt is not implemented")
}

func (m *logicModel) VisitContinueStmt(s lox.ContinueStmt) {
	panic("typing.(*logicModel).VisitContinueStmt is not implemented")
}

func (m *logicModel) VisitFunctionStmt(s lox.FunctionStmt) {
	defer func(old *scope) { m.scope = old }(m.scope)
	m.scope = newScope(m)

	params := make([]lox.Type, len(s.Params))
	for i, param := range s.Params {
		params[i] = m.scope.ref("_" + param.Lexeme)
	}

	cl := TypeClause{Head: lox.FunctionType{params, m.scope.ref("ret")}}
	m.visitStmts(s.Body)
	if !m.scope.hasReturn {
		m.scope.ref("ret").Value = lox.NilType{}
	}
	m.clauses = append(m.clauses, cl)
}

func (m *logicModel) VisitReturnStmt(s lox.ReturnStmt) {
	panic("typing.(*logicModel).VisitReturnStmt is not implemented")
}

func (m *logicModel) VisitClassStmt(s lox.ClassStmt) {
	panic("typing.(*logicModel).VisitClassStmt is not implemented")
}
