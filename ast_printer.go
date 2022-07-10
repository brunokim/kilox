package lox

import (
	"fmt"
	"strings"
)

type ASTPrinter struct {
	indentation int
	str         *strings.Builder
}

type writeStyle int

const (
	singleLine writeStyle = iota
	multiLine
)

// ----

func (p *ASTPrinter) PrintStmts(stmts []Stmt) string {
	p.str = new(strings.Builder)
	for _, stmt := range stmts {
		stmt.accept(p)
		p.str.WriteRune('\n')
	}
	return p.str.String()
}

func (p *ASTPrinter) PrintExpr(expr Expr) string {
	p.str = new(strings.Builder)
	expr.accept(p)
	return p.str.String()
}

func (p *ASTPrinter) indent() {
	for i := 0; i < p.indentation; i++ {
		p.str.WriteString("  ")
	}
}

// ----

func (p *ASTPrinter) visitBinaryExpr(expr BinaryExpr) {
	p.parenthesize(multiLine, expr.Operator, expr.Left, expr.Right)
}

func (p *ASTPrinter) visitGroupingExpr(expr GroupingExpr) {
	p.parenthesize(singleLine, "group", expr.Expression)
}

func (p *ASTPrinter) visitLiteralExpr(expr LiteralExpr) {
	value := "nil"
	if expr.Value != nil {
		value = fmt.Sprintf("%v", expr.Value)
	}
	p.printStuff(value)
}

func (p *ASTPrinter) visitUnaryExpr(expr UnaryExpr) {
	p.parenthesize(singleLine, expr.Operator, expr.Right)
}

func (p *ASTPrinter) visitVariableExpr(expr VariableExpr) {
	p.printStuff(expr.Name)
}

func (p *ASTPrinter) visitAssignmentExpr(expr AssignmentExpr) {
	p.parenthesize(singleLine, "set", expr.Name, expr.Value)
}

func (p *ASTPrinter) visitLogicExpr(expr LogicExpr) {
	p.parenthesize(multiLine, expr.Operator, expr.Left, expr.Right)
}

func (p *ASTPrinter) visitCallExpr(expr CallExpr) {
	parts := []any{expr.Callee}
	parts = append(parts, moveArray[Expr](expr.Args...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *ASTPrinter) visitFunctionExpr(expr FunctionExpr) {
	parts := []any{"fun", expr.Params}
	parts = append(parts, moveArray[Stmt](expr.Body...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *ASTPrinter) visitGetExpr(expr GetExpr) {
	panic("lox.(*ASTPrinter).visitGetExpr is not implemented")
}

// ----

func (p *ASTPrinter) visitExpressionStmt(stmt ExpressionStmt) {
	p.parenthesize(singleLine, "expr", stmt.Expression)
}

func (p *ASTPrinter) visitPrintStmt(stmt PrintStmt) {
	p.parenthesize(singleLine, "print", stmt.Expression)
}

func (p *ASTPrinter) visitVarStmt(stmt VarStmt) {
	if stmt.Init == nil {
		p.parenthesize(singleLine, "var", stmt.Name)
	} else {
		p.parenthesize(singleLine, "var", stmt.Name, stmt.Init)
	}
}

func (p *ASTPrinter) visitIfStmt(stmt IfStmt) {
	if stmt.Else == nil {
		p.parenthesize(multiLine, "if", stmt.Condition, stmt.Then)
	} else {
		p.parenthesize(multiLine, "if", stmt.Condition, stmt.Then, stmt.Else)
	}
}

func (p *ASTPrinter) visitBlockStmt(stmt BlockStmt) {
	parts := []any{"block"}
	parts = append(parts, moveArray[Stmt](stmt.Statements...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *ASTPrinter) visitLoopStmt(stmt LoopStmt) {
	if stmt.OnLoop == nil {
		p.parenthesize(multiLine, "loop", stmt.Condition, stmt.Body)
	} else {
		p.parenthesize(multiLine, "loop", stmt.Condition, stmt.Body, stmt.OnLoop)
	}
}

func (p *ASTPrinter) visitBreakStmt(stmt BreakStmt) {
	p.str.WriteString("break")
}

func (p *ASTPrinter) visitContinueStmt(stmt ContinueStmt) {
	p.str.WriteString("continue")
}

func (p *ASTPrinter) visitFunctionStmt(stmt FunctionStmt) {
	parts := []any{"defun", stmt.Name, stmt.Params}
	parts = append(parts, moveArray[Stmt](stmt.Body...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *ASTPrinter) visitReturnStmt(stmt ReturnStmt) {
	p.parenthesize(singleLine, "return", stmt.Result)
}

func (p *ASTPrinter) visitClassStmt(stmt ClassStmt) {
	panic("lox.(*ASTPrinter).visitClassStmt is not implemented")
}

// ----

func (p *ASTPrinter) parenthesize(style writeStyle, parts ...any) {
	if len(parts) <= 2 {
		style = singleLine
	}
	p.str.WriteRune('(')
	if style == singleLine {
		for i, part := range parts {
			p.printStuff(part)
			if i < len(parts)-1 {
				p.str.WriteRune(' ')
			}
		}
	} else {
		p.indentation++
		for i, part := range parts {
			if i > 0 {
				p.indent()
			}
			p.printStuff(part)
			if i < len(parts)-1 {
				p.str.WriteRune('\n')
			}
		}
		p.indentation--
	}
	p.str.WriteRune(')')
}

func (p *ASTPrinter) printStuff(x any) {
	switch stuff := x.(type) {
	case Expr:
		stuff.accept(p)
	case Stmt:
		stuff.accept(p)
	case []Token:
		parts := moveArray[Token](stuff...)
		p.parenthesize(singleLine, parts...)
	case Token:
		p.str.WriteString(stuff.Lexeme)
	case string:
		p.str.WriteString(stuff)
	}
}

func moveArray[T any](objs ...T) []any {
	arr := make([]any, len(objs))
	for i, obj := range objs {
		arr[i] = obj
	}
	return arr
}
