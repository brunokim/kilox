package lox

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Callable interface {
	Arity() int
	Call(i *Interpreter, args []interface{}) interface{}
}

var globals = NewEnvironment()

func init() {
	globals.Define("clock", clockFunc{})
}

// ----

type clockFunc struct{}

func (f clockFunc) Arity() int { return 0 }
func (f clockFunc) Call(i *Interpreter, args []interface{}) interface{} {
	return float64(time.Now().UnixMicro()) / 1e6
}
func (f clockFunc) String() string { return "<native fn clock>" }

// ----

type returnSignal struct {
	value interface{}
}

type function struct {
	name    string
	params  []Token
	body    []Stmt
	closure *Environment
}

func (f function) Arity() int {
	return len(f.params)
}

func (f function) Call(i *Interpreter, args []interface{}) (result interface{}) {
	env := f.closure.Child()
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

type Interpreter struct {
	topLevel *Environment
	env      *Environment
	value    interface{}
	stdout   io.Writer

	locals map[Expr]int
}

func NewInterpreter() *Interpreter {
	env := globals.Child()
	return &Interpreter{
		topLevel: env,
		env:      env,
		stdout:   os.Stdout,
		locals:   make(map[Expr]int),
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

func (i *Interpreter) resolve(expr Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) lookupVariable(name Token, expr Expr) interface{} {
	distance, ok := i.locals[expr]
	if ok {
		return i.env.GetAt(distance, name.Lexeme)
	}
	return i.topLevel.Get(name)
}

// ----

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
	var value interface{}
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
	i.executeBlock(stmt.Statements, i.env.Child())
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
	f := function{name, stmt.Params, stmt.Body, i.env}
	i.env.Define(name, f)
}

func (i *Interpreter) visitReturnStmt(stmt ReturnStmt) {
	var value interface{}
	if stmt.Result != nil {
		value = i.evaluate(stmt.Result)
	}
	panic(returnSignal{value})
}

// ----

func (i *Interpreter) evaluate(expr Expr) interface{} {
	expr.accept(i)
	return i.value
}

func (i *Interpreter) visitBinaryExpr(expr BinaryExpr) {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)
	i.value = operate2(expr.Operator, left, right)
}

func (i *Interpreter) visitGroupingExpr(expr GroupingExpr) {
	i.evaluate(expr.Expression)
}

func (i *Interpreter) visitLiteralExpr(expr LiteralExpr) {
	i.value = expr.Value
}

func (i *Interpreter) visitUnaryExpr(expr UnaryExpr) {
	right := i.evaluate(expr.Right)
	i.value = operate1(expr.Operator, right)
}

func (i *Interpreter) visitVariableExpr(expr VariableExpr) {
	//i.value = i.env.Get(expr.Name)
	i.value = i.lookupVariable(expr.Name, expr)
}

func (i *Interpreter) visitAssignmentExpr(expr AssignmentExpr) {
	value := i.evaluate(expr.Value)
	//i.env.Set(expr.Name, value)
	distance, ok := i.locals[expr]
	if ok {
		i.env.SetAt(distance, expr.Name, value)
	} else {
		i.topLevel.Set(expr.Name, value)
	}
}

func (i *Interpreter) visitLogicExpr(expr LogicExpr) {
	left := i.evaluate(expr.Left)
	if expr.Operator.TokenType == Or && isTruthy(left) {
		return
	}
	if expr.Operator.TokenType == And && !isTruthy(left) {
		return
	}
	i.evaluate(expr.Right)
}

func (i *Interpreter) visitCallExpr(expr CallExpr) {
	callee := i.evaluate(expr.Callee)
	args := make([]interface{}, len(expr.Args))
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

func (i *Interpreter) visitFunctionExpr(expr FunctionExpr) {
	i.value = function{"anonymous", expr.Params, expr.Body, i.env}
}

// ----

func operate2(token Token, left, right interface{}) interface{} {
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

func operate1(token Token, right interface{}) interface{} {
	switch token.TokenType {
	case Bang:
		return !isTruthy(right)
	case Minus:
		num := checkNumberOperand(token, right)
		return -num
	}
	panic(fmt.Errorf("compiler error: unimplemented unary operator %s", token.TokenType))
}

func isTruthy(v interface{}) bool {
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
	return fmt.Sprintf("operator %s in line %d: %s", err.token.Lexeme, err.token.Line, err.msg)
}

func checkNumberOperand(token Token, right interface{}) float64 {
	b, ok := right.(float64)
	if ok {
		return b
	}
	panic(runtimeError{token, "operand must be number"})
}

func checkNumberOperands(token Token, left, right interface{}) (float64, float64) {
	a, ok1 := left.(float64)
	b, ok2 := right.(float64)
	if ok1 && ok2 {
		return a, b
	}
	panic(runtimeError{token, "operands must be numbers"})
}
