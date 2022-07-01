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
func (f clockFunc) String() string { return "<fn clock>" }

// ----

type loopState int

const (
	sequentialLoop loopState = iota
	breakLoop
	continueLoop
)

// ----

type Interpreter struct {
	env    *Environment
	value  interface{}
	stdout io.Writer

	loopState loopState
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env:       globals.Child(),
		stdout:    os.Stdout,
		loopState: sequentialLoop,
	}
}

func (i *Interpreter) SetStdout(w io.Writer) {
	i.stdout = w
}

func (i *Interpreter) Interpret(stmts []Stmt) (err error) {
	defer func() {
		if err_ := recover(); err_ != nil {
			runtimeErr, ok := err_.(runtimeError)
			if !ok {
				panic(err_) // Rethrow
			}
			i.loopState = sequentialLoop
			err = runtimeErr
		}
	}()
	for _, stmt := range stmts {
		i.execute(stmt)
	}
	return nil
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
		if i.loopState == breakLoop || i.loopState == continueLoop {
			break
		}
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

func (i *Interpreter) visitWhileStmt(stmt WhileStmt) {
	for isTruthy(i.evaluate(stmt.Condition)) && i.loopState != breakLoop {
		i.loopState = sequentialLoop
		i.execute(stmt.Body)
	}
	i.loopState = sequentialLoop
}

func (i *Interpreter) visitBreakStmt(stmt BreakStmt) {
	i.loopState = breakLoop
}

func (i *Interpreter) visitContinueStmt(stmt ContinueStmt) {
	i.loopState = continueLoop
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
	i.value = i.env.Get(expr.Name)
}

func (i *Interpreter) visitAssignmentExpr(expr AssignmentExpr) {
	i.evaluate(expr.Value)
	switch e := expr.Target.(type) {
	case VariableExpr:
		i.env.Set(e.Name, i.value)
	default:
		panic(fmt.Errorf("compiler error: unhandled assignment to %T (%v)", expr.Target, expr.Target))
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
