package lox

import (
	"fmt"
	"strings"
)

type ASTPrinter struct {
	parts []string
}

func (p *ASTPrinter) Print(expr Expr) string {
	p.parts = nil
	expr.accept(p)
	return strings.Join(p.parts, "")
}

func (p *ASTPrinter) visitBinaryExpr(expr BinaryExpr) {
	p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *ASTPrinter) visitGroupingExpr(expr GroupingExpr) {
	p.parenthesize("group", expr.Expression)
}

func (p *ASTPrinter) visitLiteralExpr(expr LiteralExpr) {
	var part string
	if expr.Value == nil {
		part = "nil"
	} else {
		part = fmt.Sprintf("%v", expr.Value)
	}
	p.parts = append(p.parts, part)
}

func (p *ASTPrinter) visitUnaryExpr(expr UnaryExpr) {
	p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p *ASTPrinter) visitVariableExpr(expr VariableExpr) {
	p.parts = append(p.parts, expr.Name.Lexeme)
}

func (p *ASTPrinter) visitAssignmentExpr(expr AssignmentExpr) {
	p.parenthesize("set", expr.Target, expr.Value)
}

func (p *ASTPrinter) visitLogicExpr(expr LogicExpr) {
	p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *ASTPrinter) visitCallExpr(expr CallExpr) {
	exprs := []Expr{expr.Callee}
	exprs = append(exprs, expr.Args...)
	p.parenthesize("call", exprs...)
}

func (p *ASTPrinter) parenthesize(name string, exprs ...Expr) {
	p.parts = append(p.parts, "(", name)
	for _, expr := range exprs {
		p.parts = append(p.parts, " ")
		expr.accept(p)
	}
	p.parts = append(p.parts, ")")
}
