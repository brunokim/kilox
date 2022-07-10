package lox

import (
	"fmt"
	"strings"
)

type funcType int

const (
	noFunc funcType = iota
	namedFunc
	anonymousFunc
)

type declType int

const (
	local declType = iota
	funcName
	funcParam
	className
)

type variableState struct {
	name      Token
	decl      declType
	index     int
	isDefined bool
	isRead    bool
}

type scope struct {
	vars  []*variableState
	index map[string]int
}

type Resolver struct {
	i      *Interpreter
	scopes []*scope
	errors []resolveError

	currFunc funcType
	isInLoop bool
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		i:        interpreter,
		currFunc: noFunc,
	}
}

func (r *Resolver) Resolve(stmts []Stmt) error {
	r.resolveStmts(stmts)
	if len(r.errors) > 0 {
		return errors[resolveError](r.errors)
	}
	return nil
}

// ----

type resolveError struct {
	token Token
	msg   string
}

func (err resolveError) Error() string {
	if err.token.TokenType == EOF {
		return fmt.Sprintf("line %d at end: %s", err.token.Line, err.msg)
	}
	return fmt.Sprintf("line %d at '%s': %s", err.token.Line, err.token.Lexeme, err.msg)
}

func (r *Resolver) addError(err resolveError) {
	r.errors = append(r.errors, err)
}

// ----

func newScope() *scope {
	return &scope{
		index: make(map[string]int),
	}
}

func (s *scope) get(name string) (*variableState, bool) {
	i, ok := s.index[name]
	if !ok {
		return nil, false
	}
	return s.vars[i], true
}

func (s *scope) put(name Token, decl declType) {
	i := len(s.vars)
	s.vars = append(s.vars, &variableState{
		name:      name,
		decl:      decl,
		index:     i,
		isDefined: false,
		isRead:    false,
	})
	s.index[name.Lexeme] = i
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, newScope())
}

func (r *Resolver) endScope() {
	n := len(r.scopes)
	r.checkVariables(r.scopes[n-1])
	r.scopes = r.scopes[:n-1]
}

func (r *Resolver) declare(name Token, decl declType) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope.index[name.Lexeme]; ok {
		r.addError(resolveError{name, "already a variable with this name in scope"})
		return
	}
	scope.put(name, decl)
}

func (r *Resolver) define(name Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	state, _ := scope.get(name.Lexeme)
	state.isDefined = true
}

// ----

func (r *Resolver) resolveStmts(stmts []Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func (r *Resolver) resolveStmt(stmt Stmt) {
	stmt.accept(r)
}

func (r *Resolver) resolveExpr(expr Expr) {
	expr.accept(r)
}

func (r *Resolver) resolveLocal(expr Expr, name Token) {
	n := len(r.scopes)
	for dist := 0; dist < n; dist++ {
		scope := r.scopes[(n-1)-dist]
		if state, ok := scope.get(name.Lexeme); ok {
			state.isRead = true
			r.i.resolve(expr, dist, state.index)
			return
		}
	}
}

func (r *Resolver) resolveFunction(params []Token, body []Stmt, t funcType) {
	defer func(oldType funcType) { r.currFunc = oldType }(r.currFunc)
	r.currFunc = t

	r.beginScope()
	for _, param := range params {
		r.declare(param, funcParam)
		r.define(param)
	}
	r.resolveStmts(body)
	r.endScope()
}

// TODO: the error output order is weird, because scopes are resolved in pre-order.
// This means that 'fun unused(x) {}' reports first for 'x', and then for 'unused'.
// Figure out how to execute this (or at least sort it) in post-order.
func (r *Resolver) checkVariables(scope *scope) {
	for _, state := range scope.vars {
		if !state.isRead && !strings.HasSuffix(state.name.Lexeme, "_") {
			switch state.decl {
			case local:
				r.addError(resolveError{state.name, "local variable is never read"})
			case funcName:
				r.addError(resolveError{state.name, "function is never read or called"})
			case funcParam:
				r.addError(resolveError{state.name, "function param is never read"})
			}
		}
	}
}

// ----

func (r *Resolver) visitExpressionStmt(stmt ExpressionStmt) {
	r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitPrintStmt(stmt PrintStmt) {
	r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitVarStmt(stmt VarStmt) {
	r.declare(stmt.Name, local)
	if stmt.Init != nil {
		r.resolveExpr(stmt.Init)
	}
	r.define(stmt.Name)
}

func (r *Resolver) visitIfStmt(stmt IfStmt) {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Then)
	if stmt.Else != nil {
		r.resolveStmt(stmt.Else)
	}
}

func (r *Resolver) visitBlockStmt(stmt BlockStmt) {
	r.beginScope()
	r.resolveStmts(stmt.Statements)
	r.endScope()
}

func (r *Resolver) visitLoopStmt(stmt LoopStmt) {
	defer func(old bool) { r.isInLoop = old }(r.isInLoop)
	r.isInLoop = true
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	if stmt.OnLoop != nil {
		r.resolveExpr(stmt.OnLoop)
	}
}

func (r *Resolver) visitBreakStmt(stmt BreakStmt) {
	if !r.isInLoop {
		r.addError(resolveError{stmt.Keyword, "'break' can only be used within loops"})
	}
}

func (r *Resolver) visitContinueStmt(stmt ContinueStmt) {
	if !r.isInLoop {
		r.addError(resolveError{stmt.Keyword, "'continue' can only be used within loops"})
	}
}

func (r *Resolver) visitFunctionStmt(stmt FunctionStmt) {
	r.declare(stmt.Name, funcName)
	r.define(stmt.Name)

	r.resolveFunction(stmt.Params, stmt.Body, namedFunc)
}

func (r *Resolver) visitReturnStmt(stmt ReturnStmt) {
	if r.currFunc == noFunc {
		r.addError(resolveError{stmt.Keyword, "'return' can only be used within functions"})
	}
	if stmt.Result != nil {
		r.resolveExpr(stmt.Result)
	}
}

func (r *Resolver) visitClassStmt(stmt ClassStmt) {
	r.declare(stmt.Name, className)
	r.define(stmt.Name)
}

// ----

func (r *Resolver) visitBinaryExpr(expr BinaryExpr) {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
}

func (r *Resolver) visitGroupingExpr(expr GroupingExpr) {
	r.resolveExpr(expr.Expression)
}

func (r *Resolver) visitLiteralExpr(expr LiteralExpr) {}

func (r *Resolver) visitUnaryExpr(expr UnaryExpr) {
	r.resolveExpr(expr.Right)
}

func (r *Resolver) visitVariableExpr(expr VariableExpr) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	state, ok := scope.get(expr.Name.Lexeme)
	if ok && !state.isDefined {
		r.addError(resolveError{expr.Name, "can't read local variable in its own initializer"})
	}
	r.resolveLocal(expr, expr.Name)
}

func (r *Resolver) visitAssignmentExpr(expr AssignmentExpr) {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)
}

func (r *Resolver) visitLogicExpr(expr LogicExpr) {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
}

func (r *Resolver) visitCallExpr(expr CallExpr) {
	r.resolveExpr(expr.Callee)
	for _, arg := range expr.Args {
		r.resolveExpr(arg)
	}
}

func (r *Resolver) visitFunctionExpr(expr FunctionExpr) {
	r.resolveFunction(expr.Params, expr.Body, anonymousFunc)
}

func (r *Resolver) visitGetExpr(expr GetExpr) {
	r.resolveExpr(expr.Object)
	// We don't resolve property access statically, only dinamically.
}

func (r *Resolver) visitSetExpr(expr SetExpr) {
	panic("lox.(*Resolver).visitSetExpr is not implemented")
}
