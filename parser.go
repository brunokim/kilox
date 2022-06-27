package lox

import (
	"fmt"
)

// program     ::= declaration* eof;
// declaration ::= varDecl
//               | statement
//               ;
// varDecl     ::= "var" identifier ( "=" expression )? ";" ;
// statement   ::= exprStmt
//               | printStmt
//               | ifStmt
//               ;
// exprStmt    ::= expression ";" ;
// printStmt   ::= "print" expression ";" ;
// ifStmt      ::= "if" "(" expression ")" statement ("else" statement)? ;
// expression  ::= assignment ;
// assignment  ::= identifier "=" assignment ;
//               | logic_or
//               ;
// logic_or    ::= logic_and ("or" logic_and)* ;
// logic_and   ::= equality ("and" equality)* ;
// equality    ::= comparison (("!="|"==") comparison)* ;
// comparison  ::= term ((">"|"<"|">="|"<=") term)* ;
// term        ::= factor (("-"|"+") factor)* ;
// factor      ::= unary (("/"|"*") unary)* ;
// unary       ::= ("!"|"-") unary
//               | primary
//               ;
// primary     ::= number | string | "true" | "false" | "nil"
//               | "(" expression ")"
//               | identifier
//               ;

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
	p.consume(RightParen, "expecting ')' after expression")
	thenStmt := p.statement()
	if p.match(Else) {
		elseStmt := p.statement()
		return IfStmt{Condition: cond, Then: thenStmt, Else: elseStmt}
	}
	return IfStmt{Condition: cond, Then: thenStmt}
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
