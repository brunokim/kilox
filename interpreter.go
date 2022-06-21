package lox

import (
	"fmt"
)

type Interpreter struct {
	value interface{}
}

func (i *Interpreter) Interpret(expr Expr) (v interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(runtimeErr); !ok {
				panic(err) // Rethrow
			}
			v = nil
		}
	}()
	v = i.evaluate(expr)
	return
}

// ----

func (i *Interpreter) evaluate(expr Expr) interface{} {
	expr.accept(i)
	return i.value
}

func (i *Interpreter) visitBinaryExpr(expr Binary) {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)
	i.value = operate2(expr.Operator, left, right)
}

func (i *Interpreter) visitGroupingExpr(expr Grouping) {
	i.evaluate(expr.Expression)
}

func (i *Interpreter) visitLiteralExpr(expr Literal) {
	i.value = expr.Value
}

func (i *Interpreter) visitUnaryExpr(expr Unary) {
	right := i.evaluate(expr.Right)
	i.value = operate1(expr.Operator, right)
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
		panic(newRuntimeErr(token, "operands must be two numbers or two strings"))
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

type runtimeErr struct {
	token Token
	msg   string
}

func (err runtimeErr) Error() string {
	return fmt.Sprintf("operator %s in line %d: %s", err.token.Lexeme, err.token.Line, err.msg)
}

func newRuntimeErr(token Token, msg string) error {
	err := runtimeErr{token, msg}
	fmt.Println(err)
	return err
}

func checkNumberOperand(token Token, right interface{}) float64 {
	b, ok := right.(float64)
	if ok {
		return b
	}
	panic(newRuntimeErr(token, "operand must be number"))
}

func checkNumberOperands(token Token, left, right interface{}) (float64, float64) {
	a, ok1 := left.(float64)
	b, ok2 := right.(float64)
	if ok1 && ok2 {
		return a, b
	}
	panic(newRuntimeErr(token, "operands must be numbers"))
}
