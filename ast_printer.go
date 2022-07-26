package lox

import (
	"fmt"
	"strings"
)

type astPrinter struct {
	indentation int
	str         *strings.Builder
}

func newASTPrinter() *astPrinter {
	return &astPrinter{
		str: new(strings.Builder),
	}
}

type writeStyle int

const (
	singleLine writeStyle = iota
	multiLine
)

// ----

func PrintStmts(stmts ...Stmt) string {
	p := newASTPrinter()
	for _, stmt := range stmts {
		stmt.Accept(p)
		p.str.WriteRune('\n')
	}
	return p.str.String()
}

func PrintExpr(expr Expr) string {
	p := newASTPrinter()
	expr.Accept(p)
	return p.str.String()
}

func PrintType(t Type) string {
	p := newASTPrinter()
	t.Accept(p)
	return p.str.String()
}

// ---- Expr

func (p *astPrinter) visitBinaryExpr(expr *BinaryExpr) {
	p.parenthesize(multiLine, expr.Operator, expr.Left, expr.Right)
}

func (p *astPrinter) visitGroupingExpr(expr *GroupingExpr) {
	p.parenthesize(singleLine, "group", expr.Expression)
}

func (p *astPrinter) visitLiteralExpr(expr *LiteralExpr) {
	value := "nil"
	if expr.Value != nil {
		value = fmt.Sprintf("%v", expr.Value)
	}
	p.printStuff(value)
}

func (p *astPrinter) visitUnaryExpr(expr *UnaryExpr) {
	p.parenthesize(singleLine, expr.Operator, expr.Right)
}

func (p *astPrinter) visitVariableExpr(expr *VariableExpr) {
	p.printStuff(expr.Name)
}

func (p *astPrinter) visitAssignmentExpr(expr *AssignmentExpr) {
	p.parenthesize(singleLine, "assign", expr.Name, expr.Value)
}

func (p *astPrinter) visitLogicExpr(expr *LogicExpr) {
	p.parenthesize(multiLine, expr.Operator, expr.Left, expr.Right)
}

func (p *astPrinter) visitCallExpr(expr *CallExpr) {
	parts := []any{expr.Callee}
	parts = append(parts, moveArray[Expr](expr.Args...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *astPrinter) visitFunctionExpr(expr *FunctionExpr) {
	parts := []any{"fun", expr.Params}
	parts = append(parts, moveArray[Stmt](expr.Body...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *astPrinter) visitGetExpr(expr *GetExpr) {
	p.parenthesize(singleLine, "get", expr.Object, expr.Name)
}

func (p *astPrinter) visitSetExpr(expr *SetExpr) {
	p.parenthesize(singleLine, "set", expr.Object, expr.Name, expr.Value)
}

func (p *astPrinter) visitThisExpr(expr *ThisExpr) {
	p.str.WriteString("this")
}

// ---- Stmt

func (p *astPrinter) visitExpressionStmt(stmt ExpressionStmt) {
	p.parenthesize(singleLine, "expr", stmt.Expression)
}

func (p *astPrinter) visitPrintStmt(stmt PrintStmt) {
	p.parenthesize(singleLine, "print", stmt.Expression)
}

func (p *astPrinter) visitVarStmt(stmt VarStmt) {
	if stmt.Init == nil {
		p.parenthesize(singleLine, "var", stmt.Name)
	} else {
		p.parenthesize(singleLine, "var", stmt.Name, stmt.Init)
	}
}

func (p *astPrinter) visitIfStmt(stmt IfStmt) {
	if stmt.Else == nil {
		p.parenthesize(multiLine, "if", stmt.Condition, stmt.Then)
	} else {
		p.parenthesize(multiLine, "if", stmt.Condition, stmt.Then, stmt.Else)
	}
}

func (p *astPrinter) visitBlockStmt(stmt BlockStmt) {
	parts := []any{"block"}
	parts = append(parts, moveArray[Stmt](stmt.Statements...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *astPrinter) visitLoopStmt(stmt LoopStmt) {
	if stmt.OnLoop == nil {
		p.parenthesize(multiLine, "loop", stmt.Condition, stmt.Body)
	} else {
		p.parenthesize(multiLine, "loop", stmt.Condition, stmt.Body, stmt.OnLoop)
	}
}

func (p *astPrinter) visitBreakStmt(stmt BreakStmt) {
	p.str.WriteString("break")
}

func (p *astPrinter) visitContinueStmt(stmt ContinueStmt) {
	p.str.WriteString("continue")
}

func (p *astPrinter) visitFunctionStmt(stmt FunctionStmt) {
	parts := []any{"defun", stmt.Name, stmt.Params}
	parts = append(parts, moveArray[Stmt](stmt.Body...)...)
	p.parenthesize(multiLine, parts...)
}

func (p *astPrinter) visitReturnStmt(stmt ReturnStmt) {
	p.parenthesize(singleLine, "return", stmt.Result)
}

func (p *astPrinter) visitClassStmt(stmt ClassStmt) {
	panic("lox.(*ASTPrinter).visitClassStmt is not implemented")
}

// ---- Type

func (p *astPrinter) visitNilType(t NilType) {
	p.str.WriteString("Nil")
}

func (p *astPrinter) visitBoolType(t BoolType) {
	p.str.WriteString("Bool")
}

func (p *astPrinter) visitNumberType(t NumberType) {
	p.str.WriteString("Number")
}

func (p *astPrinter) visitStringType(t StringType) {
	p.str.WriteString("String")
}

func (p *astPrinter) visitFunctionType(t FunctionType) {
	p.parenthesize(singleLine, "Fun", t.Params, t.Return)
}

func (p *astPrinter) visitRefType(x *RefType) {
	if x.Value == nil {
		fmt.Fprintf(p.str, "_%d", x.id)
	} else {
		p.str.WriteRune('&')
		p.printStuff(x.Value)
	}
}

// ----

func (p *astPrinter) indent() {
	for i := 0; i < p.indentation; i++ {
		p.str.WriteString("  ")
	}
}

func (p *astPrinter) parenthesize(style writeStyle, parts ...any) {
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

func (p *astPrinter) printStuff(x any) {
	switch stuff := x.(type) {
	case Expr:
		stuff.Accept(p)
	case Stmt:
		stuff.Accept(p)
	case Type:
		stuff.Accept(p)
	case []Token:
		parts := moveArray[Token](stuff...)
		p.parenthesize(singleLine, parts...)
	case []Type:
		parts := moveArray[Type](stuff...)
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
