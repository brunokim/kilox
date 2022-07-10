package lox

import (
	"fmt"
)

const (
	maxCallArgs = 255
)

type Parser struct {
	tokens  []Token
	current int
	errors  []parseError
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() ([]Stmt, error) {
	var stmts []Stmt
	for !p.isAtEnd() {
		stmts = append(stmts, p.declaration())
	}
	if len(p.errors) > 0 {
		return nil, errors[parseError](p.errors)
	}
	return stmts, nil
}

func (p *Parser) ParseExpression() (expr Expr, err error) {
	defer func() {
		if err_ := recover(); err_ != nil {
			if parseErr, ok := err_.(parseError); ok {
				expr, err = nil, parseErr
			} else {
				panic(err_)
			}
		}
	}()
	return p.expression(), nil
}

// ----

func (p *Parser) declaration() Stmt {
	defer func() {
		if err := recover(); err != nil {
			if parseErr, ok := err.(parseError); ok {
				p.addError(parseErr)
				p.synchronize()
			} else {
				panic(err) // Rethrow
			}
		}
	}()
	if p.match(Var) {
		return p.varDeclaration()
	}
	if p.match(Class) {
		return p.classDeclaration()
	}
	if p.check(Fun) && !p.checkNext(LeftParen) {
		p.match(Fun)
		return p.function("function")
	}
	return p.statement()
}

func (p *Parser) varDeclaration() Stmt {
	name := p.consume(Identifier, "expecting variable name")
	var init Expr
	if p.match(Equal) {
		init = p.expression()
	}
	p.consume(Semicolon, "expecting ';' after variable declaration")
	return VarStmt{name, init}
}

func (p *Parser) classDeclaration() Stmt {
	name := p.consume(Identifier, "expecting class name")
	p.consume(LeftBrace, "expecting '{' before class body")
	var methods []FunctionStmt
	for !p.isAtEnd() && !p.check(RightBrace) {
		methods = append(methods, p.function("method"))
	}
	p.consume(RightBrace, "expecting '}' after class body")
	return ClassStmt{
		Name:    name,
		Methods: methods,
	}
}

func (p *Parser) function(kind string) FunctionStmt {
	name := p.consume(Identifier, fmt.Sprintf("expecting %s name", kind))
	params := p.functionParams(kind)
	body := p.functionBody(kind)
	return FunctionStmt{name, params, body}
}

func (p *Parser) functionParams(kind string) []Token {
	p.consume(LeftParen, fmt.Sprintf("expecting '(' after %s name", kind))
	var params []Token
	if !p.check(RightParen) {
		params = append(params, p.consume(Identifier, "expecting parameter name"))
		for p.match(Comma) {
			if len(params) == maxCallArgs {
				p.addError(parseError{p.peek(), fmt.Sprintf("can't have more than %d parameters", maxCallArgs)})
			}
			params = append(params, p.consume(Identifier, "expecting parameter name"))
		}
	}
	p.consume(RightParen, "expecting ')' after params")
	return params
}

func (p *Parser) functionBody(kind string) []Stmt {
	p.consume(LeftBrace, fmt.Sprintf("expecting '{' before %s body", kind))
	return p.block()
}

func (p *Parser) statement() Stmt {
	if p.match(Print) {
		return p.printStatement()
	}
	if p.match(If) {
		return p.ifStatement()
	}
	if p.match(LeftBrace) {
		return BlockStmt{p.block()}
	}
	if p.match(While) {
		return p.whileStatement()
	}
	if p.match(For) {
		return p.forStatement()
	}
	if p.match(Break) {
		return p.breakStatement()
	}
	if p.match(Continue) {
		return p.continueStatement()
	}
	if p.match(Return) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *Parser) printStatement() PrintStmt {
	expr := p.expression()
	p.consume(Semicolon, "expecting ';' after expression")
	return PrintStmt{expr}
}

func (p *Parser) ifStatement() IfStmt {
	p.consume(LeftParen, "expecting '(' after 'if'")
	cond := p.expression()
	p.consume(RightParen, "expecting ')' after condition")
	thenStmt := p.statement()
	if p.match(Else) {
		elseStmt := p.statement()
		return IfStmt{Condition: cond, Then: thenStmt, Else: elseStmt}
	}
	return IfStmt{Condition: cond, Then: thenStmt}
}

func (p *Parser) block() []Stmt {
	var stmts []Stmt
	for !p.check(RightBrace) && !p.isAtEnd() {
		stmts = append(stmts, p.declaration())
	}
	p.consume(RightBrace, "expecting '}' after block")
	return stmts
}

func (p *Parser) whileStatement() LoopStmt {
	p.consume(LeftParen, "expecting '(' after 'while'")
	cond := p.expression()
	p.consume(RightParen, "expecting ')' after condition")
	stmt := p.statement()
	return LoopStmt{Condition: cond, Body: stmt}
}

func (p *Parser) forStatement() Stmt {
	p.consume(LeftParen, "expecting '(' after 'for'")
	// Initializer
	var init Stmt
	if p.match(Semicolon) {
		init = nil
	} else if p.match(Var) {
		init = p.varDeclaration()
	} else {
		init = p.expressionStatement()
	}
	// Condition
	var cond Expr
	if !p.check(Semicolon) {
		cond = p.expression()
	} else {
		// This is a bit weird: an empty condition in a for statement produces a 'true'
		// condition in the while statement. However, there's no token where to hang the empty literal,
		// so we use the preceding semicolon.
		// It shouldn't matter much because the token is used only to report type error messages, and I
		// don't expect typing issues with a truthy/falsey condition.
		cond = &LiteralExpr{p.previous(), true}
	}
	p.consume(Semicolon, "Expect ';' after loop condition")
	// Increment
	var inc Expr
	if !p.check(RightParen) {
		inc = p.expression()
	}
	p.consume(RightParen, "Expect ')' after loop increment")
	// Build desugared statement.
	var body Stmt
	body = LoopStmt{
		Condition: cond,
		Body:      p.statement(),
		OnLoop:    inc,
	}
	if init != nil {
		body = BlockStmt{[]Stmt{init, body}}
	}
	return body
}

func (p *Parser) breakStatement() BreakStmt {
	token := p.previous()
	p.consume(Semicolon, "expecting ';' after 'break'")
	return BreakStmt{Keyword: token}
}

func (p *Parser) continueStatement() Stmt {
	token := p.previous()
	p.consume(Semicolon, "expecting ';' after 'continue'")
	return ContinueStmt{Keyword: token}
}

func (p *Parser) returnStatement() ReturnStmt {
	token := p.previous()
	if p.match(Semicolon) {
		return ReturnStmt{Keyword: token}
	}
	expr := p.expression()
	p.consume(Semicolon, "expecting ';' after return expression")
	return ReturnStmt{Keyword: token, Result: expr}
}

func (p *Parser) expressionStatement() ExpressionStmt {
	expr := p.expression()
	p.consume(Semicolon, "expecting ';' after expression")
	return ExpressionStmt{expr}
}

// ----

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) assignment() Expr {
	expr := p.or()
	if !p.match(Equal) {
		return expr
	}
	equals := p.previous()
	switch e := expr.(type) {
	case *VariableExpr:
		value := p.assignment()
		return &AssignmentExpr{e.Name, value}
	case *GetExpr:
		value := p.assignment()
		return &SetExpr{e.Object, e.Name, value}
	default:
		msg := fmt.Sprintf("invalid target for assignment: want variable or get expression, got %T", expr)
		p.addError(parseError{equals, msg})
		p.assignment() // Keep consuming tokens after '=', but discard them.
		return nil
	}
}

func (p *Parser) or() Expr {
	expr := p.and()
	for p.match(Or) {
		operator := p.previous()
		right := p.and()
		expr = &LogicExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) and() Expr {
	expr := p.equality()
	for p.match(And) {
		operator := p.previous()
		right := p.equality()
		expr = &LogicExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = &BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()
	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous()
		right := p.term()
		expr = &BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()
	for p.match(Minus, Plus) {
		operator := p.previous()
		right := p.factor()
		expr = &BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()
	for p.match(Slash, Star) {
		operator := p.previous()
		right := p.unary()
		expr = &BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) unary() Expr {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right := p.unary()
		return &UnaryExpr{operator, right}
	}
	return p.call()
}

func (p *Parser) call() Expr {
	expr := p.primary()
	for {
		if p.match(LeftParen) {
			expr = p.finishCall(expr)
		} else if p.match(Dot) {
			name := p.consume(Identifier, "expecting property name after '.'")
			expr = &GetExpr{expr, name}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) finishCall(callee Expr) Expr {
	var args []Expr
	if !p.check(RightParen) {
		args = append(args, p.expression())
		for p.match(Comma) {
			if len(args) == maxCallArgs {
				p.addError(parseError{p.peek(), fmt.Sprintf("can't have more than %d arguments", maxCallArgs)})
			}
			args = append(args, p.expression())
		}
	}
	paren := p.consume(RightParen, "expecting ')' after arguments")
	return &CallExpr{callee, paren, args}
}

func (p *Parser) primary() Expr {
	if p.match(False) {
		return &LiteralExpr{p.previous(), false}
	}
	if p.match(True) {
		return &LiteralExpr{p.previous(), true}
	}
	if p.match(Nil) {
		return &LiteralExpr{p.previous(), nil}
	}
	if p.match(Number, String) {
		return &LiteralExpr{p.previous(), p.previous().Literal}
	}
	if p.match(Identifier) {
		return &VariableExpr{p.previous()}
	}
	if p.match(LeftParen) {
		expr := p.expression()
		p.consume(RightParen, "expecting ')' after expression")
		return &GroupingExpr{expr}
	}
	if p.match(Fun) {
		return p.anonymousFunction()
	}
	panic(parseError{p.peek(), "expecting expression"})
}

func (p *Parser) anonymousFunction() *FunctionExpr {
	kind := "anonymous function"
	keyword := p.previous()
	params := p.functionParams(kind)
	body := p.functionBody(kind)

	return &FunctionExpr{
		Keyword: keyword,
		Params:  params,
		Body:    body,
	}
}

// ----

type parseError struct {
	token Token
	msg   string
}

func (err parseError) Error() string {
	if err.token.TokenType == EOF {
		return fmt.Sprintf("line %d at end: %s", err.token.Line, err.msg)
	}
	return fmt.Sprintf("line %d at '%s': %s", err.token.Line, err.token.Lexeme, err.msg)
}

func (p *Parser) addError(err parseError) {
	p.errors = append(p.errors, err)
}

func (p *Parser) consume(t TokenType, msg string) Token {
	if p.check(t) {
		return p.advance()
	}
	panic(parseError{p.peek(), msg})
}

func (p *Parser) synchronize() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().TokenType == Semicolon {
			return
		}
		switch p.peek().TokenType {
		case Class, For, Fun, If, Print, Return, Var, While:
			return
		}
		p.advance()
	}
}

// ----

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == t
}

func (p *Parser) checkNext(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekNext().TokenType == t
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == EOF
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) peekNext() Token {
	if p.current+1 >= len(p.tokens) {
		return Token{}
	}
	return p.tokens[p.current+1]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}
