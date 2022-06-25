package lox

import (
	"fmt"
	"io"
	"os"
)

type Interpreter struct {
	env    *Environment
	value  interface{}
	stdout io.Writer
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env:    NewEnvironment(),
		stdout: os.Stdout,
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

func (i *Interpreter) visitExpressionStmt(stmt ExpressionStmt) {
	i.evaluate(stmt.Expression)
}

func (i *Interpreter) visitPrintStmt(stmt PrintStmt) {
	v := i.evaluate(stmt.Expression)
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
