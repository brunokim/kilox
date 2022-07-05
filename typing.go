package lox

import (
	"fmt"
	"strings"
)

type LoxType interface {
	fmt.Stringer
	isLoxType()
}

type LoxNil struct{}
type LoxBool struct{}
type LoxNumber struct{}
type LoxString struct{}

type LoxFunction struct {
	Params []LoxType
	Return LoxType
}

type Ref struct {
	Value LoxType
	id    int
}

type Union struct {
	Types []LoxType
}

func (LoxNil) isLoxType()      {}
func (LoxBool) isLoxType()     {}
func (LoxNumber) isLoxType()   {}
func (LoxString) isLoxType()   {}
func (LoxFunction) isLoxType() {}
func (*Ref) isLoxType()        {}
func (*Union) isLoxType()      {}

func (LoxNil) String() string    { return "Nil" }
func (LoxBool) String() string   { return "Bool" }
func (LoxNumber) String() string { return "Number" }
func (LoxString) String() string { return "String" }

func (fn LoxFunction) String() string {
	params := make([]string, len(fn.Params))
	for i, param := range fn.Params {
		params[i] = param.String()
	}
	return fmt.Sprintf("(%s) -> %v", strings.Join(params, ", "), fn.Return)
}

func (x *Ref) String() string {
	if x.Value == nil {
		return fmt.Sprintf("_%d", x.id)
	}
	return fmt.Sprintf("&%v", x.Value)
}

func (u *Union) String() string {
	types := make([]string, len(u.Types))
	for i, t := range u.Types {
		types[i] = t.String()
	}
	return strings.Join(types, "|")
}

func deref(t LoxType) LoxType {
	x, ok := t.(*Ref)
	for ok && x.Value != nil {
		x, ok = x.Value.(*Ref)
	}
	return t
}

// ----

type typeError struct {
	t1, t2 LoxType
}

func (err typeError) Error() string {
	return fmt.Sprintf("%v != %v", err.t1, err.t2)
}

func (c *TypeChecker) addError(err typeError) {
	c.errors = append(c.errors, err)
}

// ----

type typeScope map[string]LoxType

var builtinTypes = typeScope{
	"+": LoxFunction{[]LoxType{LoxNumber{}, LoxNumber{}}, LoxNumber{}},
}

type TypeChecker struct {
	i      *Interpreter
	errors []typeError
	scopes []typeScope

	currType   LoxType
	returnType LoxType

	refID int
}

func NewTypeChecker(interpreter *Interpreter) *TypeChecker {
	return &TypeChecker{
		i:      interpreter,
		scopes: []typeScope{builtinTypes, make(typeScope)},
	}
}

func (c *TypeChecker) Check(stmts []Stmt) error {
	c.checkStmts(stmts)
	if len(c.errors) > 0 {
		return errors[typeError](c.errors)
	}
	return nil
}

func (c *TypeChecker) newRef() *Ref {
	c.refID++
	return &Ref{id: c.refID}
}

// ----

func (c *TypeChecker) beginScope() {
	c.scopes = append(c.scopes, make(typeScope))
}

func (c *TypeChecker) endScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *TypeChecker) bind(name string, type_ LoxType) {
	scope := c.scopes[len(c.scopes)-1]
	if prevType, ok := scope[name]; ok {
		c.unify(prevType, type_)
	}
	scope[name] = type_
}

type typePair [2]LoxType

func (c *TypeChecker) unify(t1, t2 LoxType) {
	t1, t2 = deref(t1), deref(t2)
	stack := []typePair{{t1, t2}}
	for len(stack) > 0 {
		n := len(stack)
		var top typePair
		top, stack = stack[n-1], stack[:n-1]
		t1, t2 = deref(top[0]), deref(top[1])
		// Handle case where one or both of the types is a *Ref.
		x1, isRef1 := t1.(*Ref)
		x2, isRef2 := t2.(*Ref)
		if isRef1 || isRef2 {
			if isRef1 && !isRef2 {
				x1.Value = t2
			} else if !isRef1 && isRef2 {
				x2.Value = t1
			} else if x1.id < x2.id {
				x2.Value = x1
			} else {
				x1.Value = x2
			}
			continue
		}
		// Handle "atomic" types: nil, bool, number, string.
		_, isNil1 := t1.(LoxNil)
		_, isNil2 := t2.(LoxNil)
		_, isBool1 := t1.(LoxBool)
		_, isBool2 := t2.(LoxBool)
		_, isNumber1 := t1.(LoxNumber)
		_, isNumber2 := t2.(LoxNumber)
		_, isString1 := t1.(LoxString)
		_, isString2 := t2.(LoxString)
		if (isNil1 && isNil2) ||
			(isBool1 && isBool2) ||
			(isNumber1 && isNumber2) ||
			(isString1 && isString2) {
			continue
		}
		isAtomic1 := isNil1 || isBool1 || isNumber1 || isString1
		isAtomic2 := isNil2 || isBool2 || isNumber2 || isString2
		if isAtomic1 || isAtomic2 {
			c.addError(typeError{t1, t2})
			continue
		}
		// Handle function types
		fn1, isFn1 := t1.(LoxFunction)
		fn2, isFn2 := t2.(LoxFunction)
		if isFn1 && isFn2 {
			if len(fn1.Params) != len(fn2.Params) {
				c.addError(typeError{t1, t2})
				continue
			}
			// Include in reverse order so that after pop'ping traversal happens in-order.
			stack = append(stack, typePair{fn1.Return, fn2.Return})
			for i := len(fn1.Params) - 1; i >= 0; i-- {
				a1, a2 := fn1.Params[i], fn2.Params[i]
				stack = append(stack, typePair{a1, a2})
			}
			continue
		}
		if isFn1 || isFn2 {
			c.addError(typeError{t1, t2})
			continue
		}
		// Unhandled type
		c.addError(typeError{t1, t2})
	}
}

func (c *TypeChecker) getBinding(name string) LoxType {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		scope := c.scopes[i]
		if t, ok := scope[name]; ok {
			return t
		}
	}
	panic(fmt.Sprintf("compiler error: variable %q not found, shouldn't happen after resolver", name))
}

// ----

func (c *TypeChecker) checkExpr(expr Expr) LoxType {
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

func (c *TypeChecker) checkFunction(params []Token, body []Stmt) LoxType {
	defer func(old LoxType) { c.returnType = old }(c.returnType)
	c.returnType = c.newRef()

	c.beginScope()
	refs := make([]LoxType, len(params))
	for i, param := range params {
		refs[i] = c.newRef()
		c.bind(param.Lexeme, refs[i])
	}
	c.checkStmts(body)
	c.endScope()

	t := LoxFunction{
		Params: refs,
		Return: c.returnType,
	}
	c.currType = t
	return t
}

func (c *TypeChecker) checkCall(callee LoxType, args ...LoxType) LoxType {
	result := c.newRef()
	callType := LoxFunction{
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
	var t LoxType
	if stmt.Init != nil {
		t = c.checkExpr(stmt.Init)
	} else {
		t = c.newRef()
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
	funcType := c.checkFunction(stmt.Params, stmt.Body)
	c.bind(stmt.Name.Lexeme, funcType)
}

func (c *TypeChecker) visitReturnStmt(stmt ReturnStmt) {
	if stmt.Result == nil {
		c.returnType = LoxNil{}
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
		c.currType = LoxBool{}
	case float64:
		c.currType = LoxNumber{}
	case string:
		c.currType = LoxString{}
	default:
		if expr.Value == nil {
			c.currType = LoxNil{}
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
	args := make([]LoxType, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = c.checkExpr(arg)
	}
	c.checkCall(t, args...)
}

func (c *TypeChecker) visitFunctionExpr(expr FunctionExpr) {
	c.checkFunction(expr.Params, expr.Body)
}
