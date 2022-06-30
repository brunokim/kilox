package lox

import (
	"fmt"
)

// program      ::= declaration* eof;
// declaration  ::= varDecl
//                | statement
//                ;
// varDecl      ::= "var" identifier ( "=" expression )? ";" ;
// statement    ::= exprStmt
//                | printStmt
//                | ifStmt
//                | block
//                | whileStmt
//                | forStmt
//                | breakStmt
//                | continueStmt
//                ;
// exprStmt     ::= expression ";" ;
// printStmt    ::= "print" expression ";" ;
// ifStmt       ::= "if" "(" expression ")" statement ("else" statement)? ;
// block        ::= "{" declaration* "}" ;
// whileStmt    ::= "while" "(" expression ")" statement ;
// forStmt      ::= "for" "(" forInit expression? ";" expression? ")" statement ;
// breakStmt    ::= "break" ";" ;
// continueStmt ::= "continue" ";" ;
// forInit      ::= varDecl | exprStmt | ";" ;
// expression   ::= assignment ;
// assignment   ::= identifier "=" assignment ;
//                | logic_or
//                ;
// logic_or     ::= logic_and ("or" logic_and)* ;
// logic_and    ::= equality ("and" equality)* ;
// equality     ::= comparison (("!="|"==") comparison)* ;
// comparison   ::= term ((">"|"<"|">="|"<=") term)* ;
// term         ::= factor (("-"|"+") factor)* ;
// factor       ::= unary (("/"|"*") unary)* ;
// unary        ::= ("!"|"-") unary
//                | primary
//                ;
// primary      ::= number | string | "true" | "false" | "nil"
//                | "(" expression ")"
//                | identifier
//                ;

type Parser struct {
	tokens  []Token
	current int
	errors  []parseError
	loopEnv *loopEnvironment
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
			parseErr, ok := err_.(parseError)
			if !ok {
				panic(err_) // Rethrow
			}
			expr = nil
			err = parseErr
		}
	}()
	return p.expression(), nil
}

// ----

type loopType int

const (
	whileLoop loopType = iota
	forLoop
)

type loopEnvironment struct {
	enclosing *loopEnvironment
	loopType  loopType
	inc       Expr
}

// ----

func (p *Parser) declaration() Stmt {
	defer func() {
		if err := recover(); err != nil {
			parseErr, ok := err.(parseError)
			if !ok {
				panic(err) // Rethrow
			}
			p.addError(parseErr)
			p.synchronize()
		}
	}()
	if p.match(Var) {
		return p.varDeclaration()
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

func (p *Parser) whileStatement() WhileStmt {
	p.loopEnv = &loopEnvironment{
		loopType:  whileLoop,
		enclosing: p.loopEnv,
	}
	p.consume(LeftParen, "expecting '(' after 'while'")
	cond := p.expression()
	p.consume(RightParen, "expecting ')' after condition")
	stmt := p.statement()
	p.loopEnv = p.loopEnv.enclosing
	return WhileStmt{Condition: cond, Body: stmt}
}

func (p *Parser) forStatement() Stmt {
	p.loopEnv = &loopEnvironment{
		loopType:  forLoop,
		enclosing: p.loopEnv,
	}
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
		cond = LiteralExpr{true}
	}
	p.consume(Semicolon, "Expect ';' after loop condition")
	// Increment
	var inc Expr
	if !p.check(RightParen) {
		inc = p.expression()
	}
	p.loopEnv.inc = inc
	p.consume(RightParen, "Expect ')' after loop increment")
	// Build desugared statement.
	body := p.statement()
	if inc != nil {
		body = BlockStmt{[]Stmt{body, ExpressionStmt{inc}}}
	}
	body = WhileStmt{cond, body}
	if init != nil {
		body = BlockStmt{[]Stmt{init, body}}
	}
	p.loopEnv = p.loopEnv.enclosing
	return body
}

func (p *Parser) breakStatement() BreakStmt {
	token := p.previous()
	if p.loopEnv == nil {
		p.addError(parseError{token, "'break' can only be used within loops"})
	}
	p.consume(Semicolon, "expecting ';' after 'break'")
	return BreakStmt{token}
}

func (p *Parser) continueStatement() Stmt {
	token := p.previous()
	if p.loopEnv == nil {
		p.addError(parseError{token, "'continue' can only be used within loops"})
	}
	p.consume(Semicolon, "expecting ';' after 'continue'")
	var stmt Stmt = ContinueStmt{token}
	if p.loopEnv != nil && p.loopEnv.inc != nil {
		stmt = BlockStmt{[]Stmt{
			ExpressionStmt{p.loopEnv.inc},
			stmt,
		}}
	}
	return stmt
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
	_, isVar := expr.(VariableExpr)
	if !isVar {
		msg := fmt.Sprintf("invalid target for assignment: want variable, got %T", expr)
		p.addError(parseError{p.previous(), msg})
		p.assignment() // Keep consuming tokens after '=', but discard them.
		return nil
	}
	value := p.assignment()
	return AssignmentExpr{expr, value}
}

func (p *Parser) or() Expr {
	expr := p.and()
	for p.match(Or) {
		operator := p.previous()
		right := p.and()
		expr = LogicExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) and() Expr {
	expr := p.equality()
	for p.match(And) {
		operator := p.previous()
		right := p.equality()
		expr = LogicExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()
	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous()
		right := p.term()
		expr = BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()
	for p.match(Minus, Plus) {
		operator := p.previous()
		right := p.factor()
		expr = BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()
	for p.match(Slash, Star) {
		operator := p.previous()
		right := p.unary()
		expr = BinaryExpr{expr, operator, right}
	}
	return expr
}

func (p *Parser) unary() Expr {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right := p.unary()
		return UnaryExpr{operator, right}
	}
	return p.primary()
}

func (p *Parser) primary() Expr {
	if p.match(False) {
		return LiteralExpr{false}
	}
	if p.match(True) {
		return LiteralExpr{true}
	}
	if p.match(Nil) {
		return LiteralExpr{nil}
	}
	if p.match(Number, String) {
		return LiteralExpr{p.previous().Literal}
	}
	if p.match(Identifier) {
		return VariableExpr{p.previous()}
	}
	if p.match(LeftParen) {
		expr := p.expression()
		p.consume(RightParen, "expecting ')' after expression")
		return GroupingExpr{expr}
	}
	panic(parseError{p.peek(), "expecting expression"})
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

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}
