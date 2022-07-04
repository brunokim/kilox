package lox

import (
	"fmt"
)

type funcType int

const (
	noFunc funcType = iota
	namedFunc
	anonymousFunc
)

type Resolver struct {
	i      *Interpreter
	scopes []map[string]bool
	errors []resolveError

	currFunc funcType
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

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope[name.Lexeme]; ok {
		r.addError(resolveError{name, "already a variable with this name in scope"})
	}
	scope[name.Lexeme] = false
}

func (r *Resolver) define(name Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = true
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
	for i := 0; i < n; i++ {
		scope := r.scopes[n-i-1]
		if _, ok := scope[name.Lexeme]; ok {
			r.i.resolve(expr, i)
			return
		}
	}
}

func (r *Resolver) resolveFunction(params []Token, body []Stmt, t funcType) {
	defer func(oldType funcType) { r.currFunc = oldType }(r.currFunc)
	r.currFunc = t

	r.beginScope()
	for _, param := range params {
		r.declare(param)
		r.define(param)
	}
	r.resolveStmts(body)
	r.endScope()
}

// ----

func (r *Resolver) visitExpressionStmt(stmt ExpressionStmt) {
	r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitPrintStmt(stmt PrintStmt) {
	r.resolveExpr(stmt.Expression)
}

func (r *Resolver) visitVarStmt(stmt VarStmt) {
	r.declare(stmt.Name)
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
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	if stmt.OnLoop != nil {
		r.resolveExpr(stmt.OnLoop)
	}
}

func (r *Resolver) visitBreakStmt(stmt BreakStmt)       {}
func (r *Resolver) visitContinueStmt(stmt ContinueStmt) {}

func (r *Resolver) visitFunctionStmt(stmt FunctionStmt) {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt.Params, stmt.Body, namedFunc)
}

func (r *Resolver) visitReturnStmt(stmt ReturnStmt) {
	if r.currFunc == noFunc {
		r.addError(resolveError{stmt.Token, "'return' can only be used within functions"})
	}
	if stmt.Result != nil {
		r.resolveExpr(stmt.Result)
	}
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
	if isDefined, ok := scope[expr.Name.Lexeme]; ok && !isDefined {
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
