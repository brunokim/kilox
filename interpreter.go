package lox

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Callable interface {
	Arity() int
	Call(i *Interpreter, args []any) any
}

// ----

type returnSignal struct {
	value any
}

type function struct {
	name    string
	params  []Token
	body    []Stmt
	closure *Environment
	isInit  bool
}

func (f function) bind(obj object) function {
	env := f.closure.Child(staticEnvironment)
	env.Define("this", obj)
	return function{f.name, f.params, f.body, env, f.isInit}
}

func (f function) getThis() instance {
	// In a method, the only variable stored in the environment is 'this'.
	return f.closure.GetStatic(0, 0).(instance)
}

func (f function) Arity() int {
	return len(f.params)
}

func (f function) Call(i *Interpreter, args []any) (result any) {
	env := f.closure.Child(staticEnvironment)
	for i, param := range f.params {
		env.Define(param.Lexeme, args[i])
	}
	defer func() {
		if r := recover(); r != nil {
			if res, ok := r.(returnSignal); ok {
				result = res.value
			} else {
				panic(r)
			}
		}
		if f.isInit {
			result = f.getThis()
		}
	}()
	i.executeBlock(f.body, env)
	return nil
}

func (f function) String() string {
	return fmt.Sprintf("<fn %s>", f.name)
}

// ----

type loopState int

const (
	sequentialLoop loopState = iota
	breakLoop
	continueLoop
)

type loopSignal struct {
	state loopState
}

// ----

type localPosition struct {
	distance int
	index    int
}

type Interpreter struct {
	globals *Environment
	env     *Environment
	value   any
	stdout  io.Writer

	locals map[Expr]localPosition
}

func NewInterpreter() *Interpreter {
	env := builtin.Child(dynamicEnvironment)
	return &Interpreter{
		globals: env,
		env:     env,
		stdout:  os.Stdout,
		locals:  make(map[Expr]localPosition),
	}
}

func (i *Interpreter) SetStdout(w io.Writer) {
	i.stdout = w
}

func (i *Interpreter) Interpret(stmts []Stmt) (err error) {
	defer func() {
		if err_ := recover(); err_ != nil {
			if runtimeErr, ok := err_.(runtimeError); ok {
				err = runtimeErr
			} else {
				panic(err_)
			}
		}
	}()
	for _, stmt := range stmts {
		i.execute(stmt)
	}
	return nil
}

func (i *Interpreter) Debug() string {
	var b strings.Builder
	b.WriteString("locals:\n")
	for expr, pos := range i.locals {
		dist, idx := pos.distance, pos.index
		fmt.Fprintf(&b, "  '%[1]p %[1]v': {dist: %[2]d, idx: %[3]d}\n", expr, dist, idx)
	}
	b.WriteString("environment:\n")
	b.WriteString(i.env.Debug())
	return b.String()
}

// ----

func (i *Interpreter) resolve(expr Expr, depth int, index int) {
	i.locals[expr] = localPosition{depth, index}
}

func (i *Interpreter) lookupVariable(name Token, expr Expr) any {
	pos, ok := i.locals[expr]
	if ok {
		return i.env.GetStatic(pos.distance, pos.index)
	}
	return i.globals.Get(name)
}

func (i *Interpreter) execute(stmt Stmt) {
	stmt.accept(i)
}

func (i *Interpreter) executeBlock(stmts []Stmt, env *Environment) {
	defer func(prev *Environment) { i.env = prev }(i.env)
	i.env = env
	for _, stmt := range stmts {
		i.execute(stmt)
	}
}

// ----

func (i *Interpreter) visitExpressionStmt(stmt ExpressionStmt) {
	i.evaluate(stmt.Expression)
}

func (i *Interpreter) visitPrintStmt(stmt PrintStmt) {
	v := i.evaluate(stmt.Expression)
	if v == nil {
		v = "nil"
	}
	fmt.Fprintln(i.stdout, v)
}

func (i *Interpreter) visitVarStmt(stmt VarStmt) {
	var value any
	if stmt.Init != nil {
		value = i.evaluate(stmt.Init)
	}
	i.env.Define(stmt.Name.Lexeme, value)
}

func (i *Interpreter) visitIfStmt(stmt IfStmt) {
	cond := i.evaluate(stmt.Condition)
	if isTruthy(cond) {
		i.execute(stmt.Then)
	} else if stmt.Else != nil {
		i.execute(stmt.Else)
	}
}

func (i *Interpreter) visitBlockStmt(stmt BlockStmt) {
	i.executeBlock(stmt.Statements, i.env.Child(staticEnvironment))
}

func (i *Interpreter) visitLoopStmt(stmt LoopStmt) {
	for isTruthy(i.evaluate(stmt.Condition)) {
		state := i.runLoopBody(stmt.Body)
		if state == breakLoop {
			break
		}
		if stmt.OnLoop != nil {
			i.evaluate(stmt.OnLoop)
		}
	}
}

func (i *Interpreter) runLoopBody(stmt Stmt) (s loopState) {
	defer func() {
		if r := recover(); r != nil {
			if signal, ok := r.(loopSignal); ok {
				s = signal.state
			} else {
				panic(r)
			}
		}
	}()
	i.execute(stmt)
	return sequentialLoop
}

func (i *Interpreter) visitBreakStmt(stmt BreakStmt) {
	panic(loopSignal{breakLoop})
}

func (i *Interpreter) visitContinueStmt(stmt ContinueStmt) {
	panic(loopSignal{continueLoop})
}

func (i *Interpreter) visitFunctionStmt(stmt FunctionStmt) {
	name := stmt.Name.Lexeme
	isInit := false
	f := function{name, stmt.Params, stmt.Body, i.env, isInit}
	i.env.Define(name, f)
}

func (i *Interpreter) visitReturnStmt(stmt ReturnStmt) {
	var value any
	if stmt.Result != nil {
		value = i.evaluate(stmt.Result)
	}
	panic(returnSignal{value})
}

func (i *Interpreter) visitClassStmt(stmt ClassStmt) {
	className := stmt.Name.Lexeme
	cl := newClass(newMetaClass(className))
	for _, method := range stmt.StaticMethods {
		methodName := method.Name.Lexeme
		isInit := false
		cl.meta.methods[methodName] = function{methodName, method.Params, method.Body, i.env, isInit}
	}
	for _, decl := range stmt.StaticVars {
		var value any = nil
		if decl.Init != nil {
			value = i.evaluate(decl.Init)
		}
		cl.static.set(decl.Name, value)
	}
	for _, decl := range stmt.Vars {
		fieldInit := fieldInitializer{name: decl.Name}
		if decl.Init != nil {
			fieldInit.value = i.evaluate(decl.Init)
		}
		cl.fieldInits = append(cl.fieldInits, fieldInit)
	}
	classEnv := i.env.Child(staticEnvironment)
	for _, method := range stmt.Methods {
		methodName := method.Name.Lexeme
		isInit := (methodName == "init")
		cl.instanceBehavior.methods[methodName] = function{methodName, method.Params, method.Body, classEnv, isInit}
	}
	i.env.Define(className, cl)
}

// ----

func (i *Interpreter) evaluate(expr Expr) any {
	expr.accept(i)
	return i.value
}

func (i *Interpreter) visitBinaryExpr(expr *BinaryExpr) {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)
	i.value = operate2(expr.Operator, left, right)
}

func (i *Interpreter) visitGroupingExpr(expr *GroupingExpr) {
	i.evaluate(expr.Expression)
}

func (i *Interpreter) visitLiteralExpr(expr *LiteralExpr) {
	i.value = expr.Value
}

func (i *Interpreter) visitUnaryExpr(expr *UnaryExpr) {
	right := i.evaluate(expr.Right)
	i.value = operate1(expr.Operator, right)
}

func (i *Interpreter) visitVariableExpr(expr *VariableExpr) {
	i.value = i.lookupVariable(expr.Name, expr)
}

func (i *Interpreter) visitAssignmentExpr(expr *AssignmentExpr) {
	value := i.evaluate(expr.Value)
	pos, ok := i.locals[expr]
	if ok {
		i.env.SetStatic(pos.distance, pos.index, value)
	} else {
		i.globals.Set(expr.Name, value)
	}
}

func (i *Interpreter) visitLogicExpr(expr *LogicExpr) {
	left := i.evaluate(expr.Left)
	if expr.Operator.TokenType == Or && isTruthy(left) {
		return
	}
	if expr.Operator.TokenType == And && !isTruthy(left) {
		return
	}
	i.evaluate(expr.Right)
}

func (i *Interpreter) visitCallExpr(expr *CallExpr) {
	callee := i.evaluate(expr.Callee)
	args := make([]any, len(expr.Args))
	for index, arg := range expr.Args {
		args[index] = i.evaluate(arg)
	}
	f, ok := callee.(Callable)
	if !ok {
		panic(runtimeError{expr.Paren, fmt.Sprintf("value %v (%T) is not callable", callee, callee)})
	}
	if f.Arity() != len(args) {
		panic(runtimeError{expr.Paren, fmt.Sprintf("expecting %d arguments but got %d", f.Arity(), len(args))})
	}
	i.value = f.Call(i, args)
}

func (i *Interpreter) visitFunctionExpr(expr *FunctionExpr) {
	isInit := false
	i.value = function{"anonymous", expr.Params, expr.Body, i.env, isInit}
}

func (i *Interpreter) visitGetExpr(expr *GetExpr) {
	obj := i.evaluate(expr.Object)
	is, ok := obj.(object)
	if !ok {
		panic(runtimeError{expr.Name, fmt.Sprintf("want an object for property access, got %[1]T (%[1]v)", obj)})
	}
	i.value = is.get(expr.Name)
}

func (i *Interpreter) visitSetExpr(expr *SetExpr) {
	obj := i.evaluate(expr.Object)
	is, ok := obj.(object)
	if !ok {
		panic(runtimeError{expr.Name, fmt.Sprintf("want an object for field access, got %[1]T (%[1]v)", obj)})
	}
	value := i.evaluate(expr.Value)
	is.set(expr.Name, value)
	i.value = value
}

func (i *Interpreter) visitThisExpr(expr *ThisExpr) {
	i.value = i.lookupVariable(expr.Keyword, expr)
}

// ----

func operate2(token Token, left, right any) any {
	switch token.TokenType {
	case BangEqual:
		return left != right
	case EqualEqual:
		return left == right
	case Greater:
		a, b := checkNumberOperands(token, left, right)
		return a > b
	case GreaterEqual:
		a, b := checkNumberOperands(token, left, right)
		return a >= b
	case Less:
		a, b := checkNumberOperands(token, left, right)
		return a < b
	case LessEqual:
		a, b := checkNumberOperands(token, left, right)
		return a <= b
	case Minus:
		a, b := checkNumberOperands(token, left, right)
		return a - b
	case Plus:
		aNum, ok1 := left.(float64)
		bNum, ok2 := right.(float64)
		if ok1 && ok2 {
			return aNum + bNum
		}
		aStr, ok3 := left.(string)
		bStr, ok4 := right.(string)
		if ok3 && ok4 {
			return aStr + bStr
		}
		panic(runtimeError{token, "operands must be two numbers or two strings"})
	case Slash:
		a, b := checkNumberOperands(token, left, right)
		return a / b
	case Star:
		a, b := checkNumberOperands(token, left, right)
		return a * b
	}
	panic(fmt.Errorf("compiler error: unimplemented binary operator %s", token.TokenType))
}

func operate1(token Token, right any) any {
	switch token.TokenType {
	case Bang:
		return !isTruthy(right)
	case Minus:
		num := checkNumberOperand(token, right)
		return -num
	}
	panic(fmt.Errorf("compiler error: unimplemented unary operator %s", token.TokenType))
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

// -----

type runtimeError struct {
	token Token
	msg   string
}

func (err runtimeError) Error() string {
	return fmt.Sprintf("token '%s' in line %d: %s", err.token.Lexeme, err.token.Line, err.msg)
}

func checkNumberOperand(token Token, right any) float64 {
	b, ok := right.(float64)
	if ok {
		return b
	}
	panic(runtimeError{token, "operand must be number"})
}

func checkNumberOperands(token Token, left, right any) (float64, float64) {
	a, ok1 := left.(float64)
	b, ok2 := right.(float64)
	if ok1 && ok2 {
		return a, b
	}
	panic(runtimeError{token, "operands must be numbers"})
}
