package typing

import (
	"fmt"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/errlist"
)

type Goal interface {
	isGoal()
}

type BindingGoal struct {
	Ref  *lox.RefType
	Type lox.Type
}

type UnificationGoal struct {
	T1, T2 lox.Type
}

type CallGoal struct {
	Name string
	Type lox.Type
}

func (BindingGoal) isGoal()     {}
func (UnificationGoal) isGoal() {}
func (CallGoal) isGoal()        {}

type TypeClause struct {
	Name string
	Head lox.FunctionType
	Body []Goal
}

// ---- Error

type logicError struct {
	msg string
}

func (err logicError) Error() string {
	return err.msg
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

func (s *scope) search(name string) (*lox.RefType, bool) {
	for s != nil {
		if x, ok := s.refs[name]; ok {
			return x, true
		}
		s = s.enclosing
	}
	return nil, false
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
	m := &logicModel{}
	m.scope = newScope(m)
	return m
}

func BuildClauses(stmts []lox.Stmt) ([]TypeClause, error) {
	m := newLogicModel()
	m.visitStmts(stmts)
	if len(m.errors) > 0 {
		return nil, errlist.Of[logicError](m.errors)
	}
	return m.clauses, nil
}

func (m *logicModel) addError(err logicError) {
	m.errors = append(m.errors, err)
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

func (m *logicModel) localRef(name lox.Token) *lox.RefType {
	return m.scope.ref("_" + name.Lexeme)
}

func (m *logicModel) search(name lox.Token) (*lox.RefType, bool) {
	return m.scope.search("_" + name.Lexeme)
}

func (m *logicModel) appendBinding(x *lox.RefType, t lox.Type) {
	cl := m.scope.clause
	cl.Body = append(cl.Body, BindingGoal{x, t})
}

func (m *logicModel) appendUnification(t1, t2 lox.Type) {
	cl := m.scope.clause
	cl.Body = append(cl.Body, UnificationGoal{t1, t2})
}

func (m *logicModel) appendCall(name lox.Token, t lox.Type) {
	cl := m.scope.clause
	cl.Body = append(cl.Body, CallGoal{name.Lexeme, t})
}

// ---- Expr

func (m *logicModel) VisitBinaryExpr(e *lox.BinaryExpr) {
	panic("typing.(*logicModel).VisitBinaryExpr is not implemented")
}

func (m *logicModel) VisitGroupingExpr(e *lox.GroupingExpr) {
	panic("typing.(*logicModel).VisitGroupingExpr is not implemented")
}

func (m *logicModel) VisitLiteralExpr(e *lox.LiteralExpr) {
	switch e.Value.(type) {
	case bool:
		m.currType = bool_
	case float64:
		m.currType = num_
	case string:
		m.currType = str_
	default:
		if e.Value == nil {
			m.currType = nil_
		} else {
			panic(fmt.Sprintf("unhandled literal type %[1]T (%[1]v)", e.Value))
		}
	}
}

func (m *logicModel) VisitUnaryExpr(e *lox.UnaryExpr) {
	panic("typing.(*logicModel).VisitUnaryExpr is not implemented")
}

func (m *logicModel) VisitVariableExpr(e *lox.VariableExpr) {
	x, ok := m.search(e.Name)
	if !ok {
		x = m.newRef()
		m.appendCall(e.Name, x)
	}
	m.currType = x
}

func (m *logicModel) VisitAssignmentExpr(e *lox.AssignmentExpr) {
	t := m.visitExpr(e.Value)
	x := m.localRef(e.Name)
	m.appendBinding(x, t)
	m.currType = t
}

func (m *logicModel) VisitLogicExpr(e *lox.LogicExpr) {
	panic("typing.(*logicModel).VisitLogicExpr is not implemented")
}

func (m *logicModel) VisitCallExpr(e *lox.CallExpr) {
	calleeType := m.visitExpr(e.Callee)
	args := make([]lox.Type, len(e.Args))
	for i, arg := range e.Args {
		args[i] = m.visitExpr(arg)
	}
	returnType := m.newRef()
	funcType := lox.FunctionType{Params: args, Return: returnType}
	m.appendUnification(calleeType, funcType)
	m.currType = returnType
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
	m.visitExpr(s.Expression)
}

func (m *logicModel) VisitPrintStmt(s lox.PrintStmt) {
	m.visitExpr(s.Expression)
}

func (m *logicModel) VisitVarStmt(s lox.VarStmt) {
	x := m.localRef(s.Name)
	if s.Init != nil {
		x.Value = m.visitExpr(s.Init)
	}
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
		params[i] = m.localRef(param)
	}

	funType := lox.FunctionType{params, m.scope.ref("ret")}
	cl := TypeClause{
		Name: s.Name.Lexeme,
		Head: funType,
	}

	m.scope.clause = &cl
	m.visitStmts(s.Body)
	if !m.scope.hasReturn {
		m.appendBinding(m.scope.ref("ret"), nil_)
	}
	m.clauses = append(m.clauses, cl)
}

func (m *logicModel) VisitReturnStmt(s lox.ReturnStmt) {
	t := m.visitExpr(s.Result)
	m.appendBinding(m.scope.ref("ret"), t)
	m.scope.hasReturn = true
}

func (m *logicModel) VisitClassStmt(s lox.ClassStmt) {
	panic("typing.(*logicModel).VisitClassStmt is not implemented")
}
