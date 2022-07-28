package typing

import (
	"fmt"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/errlist"
)

type typeScope map[string]lox.Type

func types(ts ...lox.Type) []lox.Type {
	return ts
}

func func_(params []lox.Type, result lox.Type) lox.FunctionType {
	return lox.FunctionType{
		Params: params,
		Return: result,
	}
}

var (
	nil_  = lox.NilType{}
	num_  = lox.NumberType{}
	bool_ = lox.BoolType{}
	str_  = lox.StringType{}
)

func makeBuiltinTypes() typeScope {
	scope := make(typeScope)

	var id int
	newRef := func() *lox.RefType {
		id--
		return &lox.RefType{ID: id}
	}
	t := newRef()
	t1 := newRef()
	t2 := newRef()

	// Arithmetic operators
	{
		x := newRef()
		_ = []Constraint{
			Constraint1(x, num_),
			Constraint1(x, str_),
		}
		scope["+"] = func_(types(x, x), x)
	}
	{
		x := newRef()
		_ = []Constraint{
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
		_ = []Constraint{
			Constraint1(x, t1),
			Constraint1(x, t2),
		}
		scope["and"] = func_(types(t1, t2), x)
	}
	{
		x := newRef()
		_ = []Constraint{
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

type Checker struct {
	errors []typeError
	scopes []typeScope
	types  map[lox.Expr]lox.Type

	currType   lox.Type
	returnType *lox.RefType

	refID int
}

func NewChecker() *Checker {
	return &Checker{
		scopes: []typeScope{
			makeBuiltinTypes(),
			make(typeScope), // Top-level scope
		},
		types: make(map[lox.Expr]lox.Type),
	}
}

func (c *Checker) Check(stmts []lox.Stmt) (map[lox.Expr]lox.Type, error) {
	c.checkStmts(stmts)
	if len(c.errors) > 0 {
		return nil, errlist.Of[typeError](c.errors)
	}
	return c.types, nil
}

func (c *Checker) newRefType() *lox.RefType {
	c.refID++
	return &lox.RefType{ID: c.refID}
}

func (c *Checker) GetRefID() int   { return c.refID }
func (c *Checker) SetRefID(id int) { c.refID = id }

// ----

func (c *Checker) beginScope() {
	c.scopes = append(c.scopes, make(typeScope))
}

func (c *Checker) endScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) bind(name string, type_ lox.Type) {
	scope := c.scopes[len(c.scopes)-1]
	if prevType, ok := scope[name]; ok {
		c.unify(prevType, type_)
	}
	scope[name] = type_
}

func (c *Checker) unify(t1, t2 lox.Type) {
	Unify(t1, t2)
}

func (c *Checker) getBinding(expr lox.Expr, name string) lox.Type {
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

func (c *Checker) checkExpr(expr lox.Expr) lox.Type {
	expr.Accept(c)
	return c.currType
}

func (c *Checker) checkStmts(stmts []lox.Stmt) {
	for _, stmt := range stmts {
		c.checkStmt(stmt)
	}
}

func (c *Checker) checkStmt(stmt lox.Stmt) {
	stmt.Accept(c)
}

func (c *Checker) checkFunctionType(name string, params []lox.Token, body []lox.Stmt) lox.Type {
	defer func(old *lox.RefType) { c.returnType = old }(c.returnType)
	c.returnType = c.newRefType()
	refs := make([]lox.Type, len(params))
	for i := 0; i < len(refs); i++ {
		refs[i] = c.newRefType()
	}
	t := lox.FunctionType{
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

func (c *Checker) checkCall(callee lox.Type, args ...lox.Type) lox.Type {
	result := c.newRefType()
	callType := lox.FunctionType{
		Params: args,
		Return: result,
	}
	c.unify(Copy(callee, c.newRefType), callType)
	c.currType = result
	return result
}

func (c *Checker) constraintReturn(t lox.Type) {
	c.returnType.Value = t
}

// ----

func (c *Checker) VisitExpressionStmt(stmt lox.ExpressionStmt) {
	c.checkExpr(stmt.Expression)
}

func (c *Checker) VisitPrintStmt(stmt lox.PrintStmt) {
	c.checkExpr(stmt.Expression)
}

// If var is not initialized, its type may be initialized on first assignment.
// TODO: handle case where an uninitialized variable is read/returned before first
// assignment, in which case it should be nil.
func (c *Checker) VisitVarStmt(stmt lox.VarStmt) {
	var t lox.Type
	if stmt.Init != nil {
		t = c.checkExpr(stmt.Init)
	} else {
		t = c.newRefType()
	}
	c.bind(stmt.Name.Lexeme, t)
}

func (c *Checker) VisitIfStmt(stmt lox.IfStmt) {
	// stmt.Condition is always valid.
	c.checkExpr(stmt.Condition)
	c.checkStmt(stmt.Then)
	if stmt.Else != nil {
		c.checkStmt(stmt.Else)
	}
}

func (c *Checker) VisitBlockStmt(stmt lox.BlockStmt) {
	c.beginScope()
	for _, stmt := range stmt.Statements {
		c.checkStmt(stmt)
	}
	c.endScope()
}

func (c *Checker) VisitLoopStmt(stmt lox.LoopStmt) {
	// stmt.Condition is always valid.
	c.checkExpr(stmt.Condition)
	c.checkStmt(stmt.Body)
	if stmt.OnLoop != nil {
		c.checkExpr(stmt.OnLoop)
	}
}

func (c *Checker) VisitBreakStmt(stmt lox.BreakStmt) {
	// Do nothing.
}

func (c *Checker) VisitContinueStmt(stmt lox.ContinueStmt) {
	// Do nothing.
}

func (c *Checker) VisitFunctionStmt(stmt lox.FunctionStmt) {
	name := stmt.Name.Lexeme
	c.checkFunctionType(name, stmt.Params, stmt.Body)
}

func (c *Checker) VisitReturnStmt(stmt lox.ReturnStmt) {
	if stmt.Result == nil {
		c.constraintReturn(lox.NilType{})
		return
	}
	c.constraintReturn(c.checkExpr(stmt.Result))
}

func (c *Checker) VisitClassStmt(stmt lox.ClassStmt) {
	panic("lox.(*Checker).visitClassStmt is not implemented")
}

// ----

func (c *Checker) VisitBinaryExpr(expr *lox.BinaryExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *Checker) VisitGroupingExpr(expr *lox.GroupingExpr) {
	c.checkExpr(expr.Expression)
}

func (c *Checker) VisitLiteralExpr(expr *lox.LiteralExpr) {
	switch expr.Value.(type) {
	case bool:
		c.currType = lox.BoolType{expr.Token}
	case float64:
		c.currType = lox.NumberType{expr.Token}
	case string:
		c.currType = lox.StringType{expr.Token}
	default:
		if expr.Value == nil {
			c.currType = lox.NilType{expr.Token}
		} else {
			panic(fmt.Sprintf("unhandled literal type %[1]T (%[1]v)", expr.Value))
		}
	}
}

func (c *Checker) VisitUnaryExpr(expr *lox.UnaryExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, right)
}

func (c *Checker) VisitVariableExpr(expr *lox.VariableExpr) {
	c.currType = c.getBinding(expr, expr.Name.Lexeme)
}

func (c *Checker) VisitAssignmentExpr(expr *lox.AssignmentExpr) {
	t := c.checkExpr(expr.Value)
	c.bind(expr.Name.Lexeme, t)
	c.currType = t
}

func (c *Checker) VisitLogicExpr(expr *lox.LogicExpr) {
	op := c.getBinding(expr, expr.Operator.Lexeme)
	left := c.checkExpr(expr.Left)
	right := c.checkExpr(expr.Right)
	c.checkCall(op, left, right)
}

func (c *Checker) VisitCallExpr(expr *lox.CallExpr) {
	t := c.checkExpr(expr.Callee)
	args := make([]lox.Type, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = c.checkExpr(arg)
	}
	c.checkCall(t, args...)
}

func (c *Checker) VisitFunctionExpr(expr *lox.FunctionExpr) {
	c.checkFunctionType("", expr.Params, expr.Body)
}

func (c *Checker) VisitGetExpr(expr *lox.GetExpr) {
	panic("typing.(*Checker).visitGetExpr is not implemented")
}

func (c *Checker) VisitSetExpr(expr *lox.SetExpr) {
	panic("typing.(*Checker).visitSetExpr is not implemented")
}

func (c *Checker) VisitThisExpr(expr *lox.ThisExpr) {
	panic("typing.(*Checker).visitThisExpr is not implemented")
}
