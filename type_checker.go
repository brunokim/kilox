package lox

import (
	"fmt"

	"github.com/brunokim/lox/errlist"
)

type typeScope map[string]Type

func types(ts ...Type) []Type {
	return ts
}

func func_(params []Type, result Type) FunctionType {
	return FunctionType{
		Params: params,
		Return: result,
	}
}

var (
	nil_  = NilType{}
	num_  = NumberType{}
	bool_ = BoolType{}
	str_  = StringType{}
)

func makeBuiltinTypes() typeScope {
	scope := make(typeScope)

	var id int
	newRef := func() *RefType {
		id--
		return &RefType{id: id}
	}
	t := newRef()
	t1 := newRef()
	t2 := newRef()

	// Arithmetic operators
	{
		x := newRef()
		x.constraints = []Constraint{
			Constraint1(x, num_),
			Constraint1(x, str_),
		}
		scope["+"] = func_(types(x, x), x)
	}
	{
		x := newRef()
		x.constraints = []Constraint{
			Constraint1(x, func_(types(num_, num_), num_)),
			Constraint1(x, func_(types(num_), num_)),
		}
		scope["-"] = x
	}
	scope["*"] = func_(types(num_, num_), num_)
	scope["/"] = func_(types(num_, num_), num_)

	// Logic operators
	scope["<"] = func_(types(num_, num_), bool_)
	scope["<="] = func_(types(num_, num_), bool_)
	scope[">"] = func_(types(num_, num_), bool_)
	scope[">="] = func_(types(num_, num_), bool_)
	scope["=="] = func_(types(t1, t2), bool_)
	scope["!="] = func_(types(t1, t2), bool_)
	scope["!"] = func_(types(t), bool_)

	// Logic control
	{
		x := newRef()
		x.constraints = []Constraint{
			Constraint1(x, t1),
			Constraint1(x, t2),
		}
		scope["and"] = func_(types(t1, t2), x)
	}
	{
		x := newRef()
		x.constraints = []Constraint{
			Constraint1(x, t1),
			Constraint1(x, t2),
		}
		scope["or"] = func_(types(t1, t2), x)
	}

	// Builtin
	scope["clock"] = func_(types(), num_)
	scope["type"] = func_(types(t), t1)
	scope["random"] = func_(types(), num_)
	scope["randomSeed"] = func_(types(num_), nil_)

	return scope
}

type TypeChecker struct {
	errors []typeError
	scopes []typeScope
	types  map[Expr]Type

	currType   Type
	returnType *RefType

	refID int
}

func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		scopes: []typeScope{
			makeBuiltinTypes(),
			make(typeScope), // Top-level scope
		},
		types: make(map[Expr]Type),
	}
}

func (c *TypeChecker) Check(stmts []Stmt) (map[Expr]Type, error) {
	c.checkStmts(stmts)
	if len(c.errors) > 0 {
		return nil, errlist.Errors[typeError](c.errors)
	}
	return c.types, nil
}

func (c *TypeChecker) newRefType() *RefType {
	c.refID++
	return &RefType{id: c.refID}
}

func (c *TypeChecker) getRefID() int   { return c.refID }
func (c *TypeChecker) setRefID(id int) { c.refID = id }

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
	u := newUnifier(c)
	u.unify(t1, t2)
}

func (c *TypeChecker) getBinding(expr Expr, name string) Type {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		scope := c.scopes[i]
		if t, ok := scope[name]; ok {
			c.types[expr] = t
			return t
		}
	}
	panic(fmt.Sprintf("compiler error: variable %q not found, shouldn't happen after resolver", name))
}

// ----

func (c *TypeChecker) checkExpr(expr Expr) Type {
	expr.Accept(c)
	return c.currType
}

func (c *TypeChecker) checkStmts(stmts []Stmt) {
	for _, stmt := range stmts {
		c.checkStmt(stmt)
	}
}

func (c *TypeChecker) checkStmt(stmt Stmt) {
	stmt.Accept(c)
}

func (c *TypeChecker) checkFunctionType(name string, params []Token, body []Stmt) Type {
	defer func(old *RefType) { c.returnType = old }(c.returnType)
	c.returnType = c.newRefType()
	refs := make([]Type, len(params))
	for i := 0; i < len(refs); i++ {
		refs[i] = c.newRefType()
	}
	t := FunctionType{
		Params: refs,
		Return: c.returnType,
	}
	if name != "" {
		// Binds function name so it can be referred from inside the function.
		c.bind(name, t)
	}
	c.beginScope()
	for i, param := range params {
		c.bind(param.Lexeme, refs[i])
	}
	c.checkStmts(body)
	c.endScope()
	c.currType = simplifyType(t)
	return t
}

func (c *TypeChecker) checkCall(callee Type, args ...Type) Type {
	result := c.newRefType()
	callType := FunctionType{
		Params: args,
		Return: result,
	}
	c.unify(Copy(callee, c.newRefType), callType)
	c.currType = result
	return result
}

func (c *TypeChecker) constraintReturn(t Type) {
	x := c.returnType
	x.constraints = append(x.constraints, Constraint1(x, t))
}

// ----

func (c *TypeChecker) visitExpressionStmt(stmt ExpressionStmt) {
	c.checkExpr(stmt.Expression)
}

func (c *TypeChecker) visitPrintStmt(stmt PrintStmt) {
	c.checkExpr(stmt.Expression)
}

// If var is not initialized, its type may be initialized on first assignment.
// TODO: handle case where an uninitialized variable is read/returned before first
// assignment, in which case it should be nil.
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
	c.checkExpr(stmt.Condition)
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
	c.checkExpr(stmt.Condition)
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
	name := stmt.Name.Lexeme
	c.checkFunctionType(name, stmt.Params, stmt.Body)
}

func (c *TypeChecker) visitReturnStmt(stmt ReturnStmt) {
	if stmt.Result == nil {
		c.constraintReturn(NilType{})
		return
	}
	c.constraintReturn(c.checkExpr(stmt.Result))
}

func (c *TypeChecker) visitClassStmt(stmt ClassStmt) {
	panic("lox.(*TypeChecker).visitClassStmt is not implemented")
}

// ----

func (c *TypeChecker) visitBinaryExpr(expr *BinaryExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *TypeChecker) visitGroupingExpr(expr *GroupingExpr) {
	c.checkExpr(expr.Expression)
}

func (c *TypeChecker) visitLiteralExpr(expr *LiteralExpr) {
	switch expr.Value.(type) {
	case bool:
		c.currType = BoolType{expr.Token}
	case float64:
		c.currType = NumberType{expr.Token}
	case string:
		c.currType = StringType{expr.Token}
	default:
		if expr.Value == nil {
			c.currType = NilType{expr.Token}
		} else {
			panic(fmt.Sprintf("unhandled literal type %[1]T (%[1]v)", expr.Value))
		}
	}
}

func (c *TypeChecker) visitUnaryExpr(expr *UnaryExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, right)
}

func (c *TypeChecker) visitVariableExpr(expr *VariableExpr) {
	c.currType = c.getBinding(expr, expr.Name.Lexeme)
}

func (c *TypeChecker) visitAssignmentExpr(expr *AssignmentExpr) {
	t := c.checkExpr(expr.Value)
	c.bind(expr.Name.Lexeme, t)
	c.currType = t
}

func (c *TypeChecker) visitLogicExpr(expr *LogicExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *TypeChecker) visitCallExpr(expr *CallExpr) {
	t := c.checkExpr(expr.Callee)
	args := make([]Type, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = c.checkExpr(arg)
	}
	c.checkCall(t, args...)
}

func (c *TypeChecker) visitFunctionExpr(expr *FunctionExpr) {
	c.checkFunctionType("", expr.Params, expr.Body)
}

func (c *TypeChecker) visitGetExpr(expr *GetExpr) {
	panic("lox.(*TypeChecker).visitGetExpr is not implemented")
}

func (c *TypeChecker) visitSetExpr(expr *SetExpr) {
	panic("lox.(*TypeChecker).visitSetExpr is not implemented")
}

func (c *TypeChecker) visitThisExpr(expr *ThisExpr) {
	panic("lox.(*TypeChecker).visitThisExpr is not implemented")
}
