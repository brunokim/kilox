package lox

import (
	"fmt"
)

// expression ::= equality ;
// equality   ::= comparison (("!="|"==") comparison)* ;
// comparison ::= term ((">"|"<"|">="|"<=") term)* ;
// term       ::= factor (("-"|"+") factor)* ;
// factor     ::= unary (("/"|"*") unary)* ;
// unary      ::= ("!"|"-") unary
//              | primary
//              ;
// primary    ::= number | string | "true" | "false" | "nil"
//              | "(" expression ")"
//              ;

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() (expr Expr) {
	defer func() {
		if err := recover(); err != nil {
			err, ok := err.(parseError)
			if !ok {
				// Rethrow
				panic(err)
			}
			expr = nil
		}
	}()
	expr = p.expression()
	return
}

// ----

func (p *Parser) expression() Expr {
	return p.equality()
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
	if p.match(LeftParen) {
		expr := p.expression()
		p.consume(RightParen, "expect ')' after expression")
		return GroupingExpr{expr}
	}
	panic(p.err(p.peek(), "expect expression"))
}

// ----

type parseError struct {
	token Token
	msg   string
}

func (err parseError) Error() string {
	return fmt.Sprintf("%v - %s", err.token, err.msg)
}

func (p *Parser) consume(t TokenType, msg string) Token {
	if p.check(t) {
		return p.advance()
	}
	panic(p.err(p.peek(), msg))
}

func (p *Parser) err(token Token, msg string) error {
	if token.TokenType == EOF {
		fmt.Printf("line %d at end: %s\n", token.Line, msg)
	} else {
		fmt.Printf("line %d at '%s': %s\n", token.Line, token.Lexeme, msg)
	}
	return parseError{token, msg}
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
