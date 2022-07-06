package lox

import (
	"fmt"
)

type typeScope map[string]Type

var (
	t  = &RefType{}
	t1 = &RefType{}
	t2 = &RefType{}
)

var builtinTypes = typeScope{
	// Arithmetic operators
	"+": FunctionType{[]Type{NumberType{}, NumberType{}}, NumberType{}},
	"-": FunctionType{[]Type{NumberType{}, NumberType{}}, NumberType{}},
	"*": FunctionType{[]Type{NumberType{}, NumberType{}}, NumberType{}},
	"/": FunctionType{[]Type{NumberType{}, NumberType{}}, NumberType{}},
	// Logic operators
	"<":  FunctionType{[]Type{NumberType{}, NumberType{}}, BoolType{}},
	"<=": FunctionType{[]Type{NumberType{}, NumberType{}}, BoolType{}},
	">":  FunctionType{[]Type{NumberType{}, NumberType{}}, BoolType{}},
	">=": FunctionType{[]Type{NumberType{}, NumberType{}}, BoolType{}},
	"==": FunctionType{[]Type{t1, t2}, BoolType{}},
	"!=": FunctionType{[]Type{t1, t2}, BoolType{}},
	// Logic control
	"and": FunctionType{[]Type{t1, t2}, unionTypes(t1, t2)},
	"or":  FunctionType{[]Type{t1, t2}, unionTypes(t1, t2)},
	// Builtin
	"clock": FunctionType{[]Type{}, NumberType{}},
}

type TypeChecker struct {
	i      *Interpreter
	errors []typeError
	scopes []typeScope

	currType   Type
	returnType Type

	refID int
}

func NewTypeChecker(interpreter *Interpreter) *TypeChecker {
	return &TypeChecker{
		i: interpreter,
		scopes: []typeScope{
			builtinTypes,
			make(typeScope), // Top-level scope
		},
	}
}

func (c *TypeChecker) Check(stmts []Stmt) error {
	c.checkStmts(stmts)
	if len(c.errors) > 0 {
		return errors[typeError](c.errors)
	}
	return nil
}

func (c *TypeChecker) newRefType() *RefType {
	c.refID++
	return &RefType{id: c.refID}
}

// ----

func (c *TypeChecker) beginScope() {
	c.scopes = append(c.scopes, make(typeScope))
}

func (c *TypeChecker) endScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *TypeChecker) bind(name string, type_ Type) {
	scope := c.scopes[len(c.scopes)-1]
	if prevType, ok := scope[name]; ok {
		c.unify(prevType, type_)
	}
	scope[name] = type_
}

func (c *TypeChecker) unify(t1, t2 Type) {
	u := &unifier{c: c}
	u.unify(t1, t2)
}

func (c *TypeChecker) getBinding(name string) Type {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		scope := c.scopes[i]
		if t, ok := scope[name]; ok {
			return t
		}
	}
	panic(fmt.Sprintf("compiler error: variable %q not found, shouldn't happen after resolver", name))
}

// ----

func (c *TypeChecker) checkExpr(expr Expr) Type {
	expr.accept(c)
	return c.currType
}

func (c *TypeChecker) checkStmts(stmts []Stmt) {
	for _, stmt := range stmts {
		c.checkStmt(stmt)
	}
}

func (c *TypeChecker) checkStmt(stmt Stmt) {
	stmt.accept(c)
}

func (c *TypeChecker) checkFunctionType(params []Token, body []Stmt) Type {
	defer func(old Type) { c.returnType = old }(c.returnType)
	c.returnType = c.newRefType()

	c.beginScope()
	refs := make([]Type, len(params))
	for i, param := range params {
		refs[i] = c.newRefType()
		c.bind(param.Lexeme, refs[i])
	}
	c.checkStmts(body)
	c.endScope()

	t := FunctionType{
		Params: refs,
		Return: c.returnType,
	}
	c.currType = t
	return t
}

func (c *TypeChecker) checkCall(callee Type, args ...Type) Type {
	result := c.newRefType()
	callType := FunctionType{
		Params: args,
		Return: result,
	}
	c.unify(callee, callType)
	c.currType = result
	return result
}

// ----

func (c *TypeChecker) visitExpressionStmt(stmt ExpressionStmt) {
	c.checkExpr(stmt.Expression)
}

func (c *TypeChecker) visitPrintStmt(stmt PrintStmt) {
	c.checkExpr(stmt.Expression)
}

func (c *TypeChecker) visitVarStmt(stmt VarStmt) {
	var t Type
	if stmt.Init != nil {
		t = c.checkExpr(stmt.Init)
	} else {
		t = c.newRefType()
	}
	c.bind(stmt.Name.Lexeme, t)
}

func (c *TypeChecker) visitIfStmt(stmt IfStmt) {
	// stmt.Condition is always valid.
	c.checkStmt(stmt.Then)
	if stmt.Else != nil {
		c.checkStmt(stmt.Else)
	}
}

func (c *TypeChecker) visitBlockStmt(stmt BlockStmt) {
	c.beginScope()
	for _, stmt := range stmt.Statements {
		c.checkStmt(stmt)
	}
	c.endScope()
}

func (c *TypeChecker) visitLoopStmt(stmt LoopStmt) {
	// stmt.Condition is always valid.
	c.checkStmt(stmt.Body)
	if stmt.OnLoop != nil {
		c.checkExpr(stmt.OnLoop)
	}
}

func (c *TypeChecker) visitBreakStmt(stmt BreakStmt) {
	// Do nothing.
}

func (c *TypeChecker) visitContinueStmt(stmt ContinueStmt) {
	// Do nothing.
}

func (c *TypeChecker) visitFunctionStmt(stmt FunctionStmt) {
	funcType := c.checkFunctionType(stmt.Params, stmt.Body)
	c.bind(stmt.Name.Lexeme, funcType)
}

func (c *TypeChecker) visitReturnStmt(stmt ReturnStmt) {
	if stmt.Result == nil {
		c.returnType = NilType{}
		return
	}
	c.returnType = c.checkExpr(stmt.Result)
}

// ----

func (c *TypeChecker) visitBinaryExpr(expr BinaryExpr) {
	op := c.getBinding(expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *TypeChecker) visitGroupingExpr(expr GroupingExpr) {
	c.checkExpr(expr.Expression)
}

func (c *TypeChecker) visitLiteralExpr(expr LiteralExpr) {
	switch expr.Value.(type) {
	case bool:
		c.currType = BoolType{}
	case float64:
		c.currType = NumberType{}
	case string:
		c.currType = StringType{}
	default:
		if expr.Value == nil {
			c.currType = NilType{}
		} else {
			panic(fmt.Sprintf("unhandled literal type %[1]T (%[1]v)", expr.Value))
		}
	}
}

func (c *TypeChecker) visitUnaryExpr(expr UnaryExpr) {
	op := c.getBinding(expr.Operator.Lexeme)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, right)
}

func (c *TypeChecker) visitVariableExpr(expr VariableExpr) {
	c.currType = c.getBinding(expr.Name.Lexeme)
}

func (c *TypeChecker) visitAssignmentExpr(expr AssignmentExpr) {
	t := c.checkExpr(expr.Value)
	c.bind(expr.Name.Lexeme, t)
	c.currType = t
}

func (c *TypeChecker) visitLogicExpr(expr LogicExpr) {
	op := c.getBinding(expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *TypeChecker) visitCallExpr(expr CallExpr) {
	t := c.checkExpr(expr.Callee)
	args := make([]Type, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = c.checkExpr(arg)
	}
	c.checkCall(t, args...)
}

func (c *TypeChecker) visitFunctionExpr(expr FunctionExpr) {
	c.checkFunctionType(expr.Params, expr.Body)
}
