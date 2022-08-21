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
	ClauseID int
	Type     lox.Type
}

func (BindingGoal) isGoal()     {}
func (UnificationGoal) isGoal() {}
func (*CallGoal) isGoal()       {}

type TypeClause struct {
	ID   int
	Name string
	Head lox.FunctionType
	Body []Goal
}

var builtinClauses = []TypeClause{
	{-99, "+", func_(types(num_, num_), num_), nil},
	{-98, "-", func_(types(num_, num_), num_), nil},
	{-97, "*", func_(types(num_, num_), num_), nil},
	{-96, "*", func_(types(num_, num_), num_), nil},
}

// ---- Error

type logicError struct {
	msg string
}

func (err logicError) Error() string {
	return err.msg
}

// ---- Scope

type dynType int

const (
	dynamicScope dynType = iota
	staticScope
)

type scope struct {
	m         *logicModel
	enclosing *scope
	clause    *TypeClause

	refs      map[string]*lox.RefType
	clauseIDs map[string]int
	forwards  map[string][]*CallGoal

	returnRef *lox.RefType
	hasReturn bool
	dynType   dynType
}

func builtinScope(m *logicModel) *scope {
	s := newScope(m, dynamicScope)
	for _, clause := range builtinClauses {
		s.refs[clause.Name] = m.newRef()
		s.refs[clause.Name].Value = clause.Head
		s.clauseIDs[clause.Name] = clause.ID
	}
	return s
}

func newScope(m *logicModel, dynType dynType) *scope {
	return &scope{
		m:         m,
		enclosing: m.scope,
		refs:      make(map[string]*lox.RefType),
		clauseIDs: make(map[string]int),
		forwards:  make(map[string][]*CallGoal),
		dynType:   dynType,
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

func (s *scope) addForward(name string, call *CallGoal) {
	s.forwards[name] = append(s.forwards[name], call)
}

// ---- Logic model

type logicModel struct {
	clauses []TypeClause
	errors  []logicError
	scope   *scope

	currType lox.Type

	refID    int
	clauseID int
}

func newLogicModel() *logicModel {
	m := &logicModel{}
	m.scope = builtinScope(m)
	m.scope = newScope(m, dynamicScope) // global scope
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
	return m.scope.ref(name.Lexeme)
}

func (m *logicModel) search(name lox.Token) (*lox.RefType, bool) {
	return m.scope.search(name.Lexeme)
}

func (m *logicModel) appendBinding(x *lox.RefType, t lox.Type) {
	cl := m.scope.clause
	cl.Body = append(cl.Body, BindingGoal{x, t})
}

func (m *logicModel) appendUnification(t1, t2 lox.Type) {
	cl := m.scope.clause
	cl.Body = append(cl.Body, UnificationGoal{t1, t2})
}

func (m *logicModel) appendCall(clauseID int, t lox.Type) *CallGoal {
	cl := m.scope.clause
	call := &CallGoal{clauseID, t}
	cl.Body = append(cl.Body, call)
	return call
}

// ----

func (m *logicModel) nameType(name lox.Token) lox.Type {
	x, ok := m.search(name)
	if !ok {
		x = m.newRef()
		call := m.appendCall(0, x)
		m.scope.addForward(name.Lexeme, call)
	}
	return x
}

func (m *logicModel) callType(calleeType lox.Type, args ...lox.Expr) lox.Type {
	argTypes := make([]lox.Type, len(args))
	for i, arg := range args {
		argTypes[i] = m.visitExpr(arg)
	}
	returnType := m.newRef()
	funcType := lox.FunctionType{Params: argTypes, Return: returnType}
	m.appendUnification(calleeType, funcType)
	return returnType
}

// ---- Expr

func (m *logicModel) VisitBinaryExpr(e *lox.BinaryExpr) {
	opType := m.nameType(e.Operator)
	m.currType = m.callType(opType, e.Left, e.Right)
}

func (m *logicModel) VisitGroupingExpr(e *lox.GroupingExpr) {
	m.currType = m.visitExpr(e.Expression)
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
	opType := m.nameType(e.Operator)
	m.currType = m.callType(opType, e.Right)
}

func (m *logicModel) VisitVariableExpr(e *lox.VariableExpr) {
	m.currType = m.nameType(e.Name)
}

func (m *logicModel) VisitAssignmentExpr(e *lox.AssignmentExpr) {
	t := m.visitExpr(e.Value)
	x := m.localRef(e.Name)
	m.appendBinding(x, t)
	m.currType = t
}

func (m *logicModel) VisitLogicExpr(e *lox.LogicExpr) {
	opType := m.nameType(e.Operator)
	m.currType = m.callType(opType, e.Left, e.Right)
}

func (m *logicModel) VisitCallExpr(e *lox.CallExpr) {
	calleeType := m.visitExpr(e.Callee)
	m.currType = m.callType(calleeType, e.Args...)
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
	// Declare clause in the surrounding scope.
	m.clauseID++
	m.scope.clauseIDs[s.Name.Lexeme] = m.clauseID
	funRef := m.localRef(s.Name)

	if goals, ok := m.scope.forwards[s.Name.Lexeme]; ok && m.scope.dynType == dynamicScope {
		// This statement defines a name forward-referenced before. Mutate those call goals to point
		// to this newly created clause.
		for _, goal := range goals {
			goal.ClauseID = m.clauseID
		}
		delete(m.scope.forwards, s.Name.Lexeme)
	}

	// Push new scope.
	m.scope = newScope(m, staticScope)

	// Create clause with function type as head.
	params := make([]lox.Type, len(s.Params))
	for i, param := range s.Params {
		params[i] = m.localRef(param)
	}
	m.scope.returnRef = m.newRef()
	funType := lox.FunctionType{params, m.scope.returnRef}
	funRef.Value = funType
	cl := TypeClause{
		ID:   m.clauseID,
		Name: s.Name.Lexeme,
		Head: funType,
	}
	m.scope.clause = &cl
	m.visitStmts(s.Body)
	if !m.scope.hasReturn {
		m.appendBinding(m.scope.returnRef, nil_)
	}
	m.clauses = append(m.clauses, cl)

	// Append forwards to enclosing scope and restore it.
	for name, goals := range m.scope.forwards {
		m.scope.enclosing.forwards[name] = append(m.scope.enclosing.forwards[name], goals...)
	}
	m.scope = m.scope.enclosing
}

func (m *logicModel) VisitReturnStmt(s lox.ReturnStmt) {
	t := m.visitExpr(s.Result)
	m.appendBinding(m.scope.returnRef, t)
	m.scope.hasReturn = true
}

func (m *logicModel) VisitClassStmt(s lox.ClassStmt) {
	panic("typing.(*logicModel).VisitClassStmt is not implemented")
}
