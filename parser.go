package lox

import (
	"fmt"
	"strings"
)

// program     ::= declaration* eof;
// declaration ::= varDecl
//               | statement
//               ;
// varDecl     ::= "var" identifier ( "=" expression )? ";" ;
// statement   ::= exprStmt
//               | printStmt
//               ;
// exprStmt    ::= expression ";" ;
// printStmt   ::= "print" expression ";" ;
// expression  ::= equality ;
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
		return nil, parseErrors(p.errors)
	}
	return stmts, nil
}

// ----

func (p *Parser) declaration() Stmt {
	defer func() {
		if err := recover(); err != nil {
			runtimeErr, ok := err.(parseError)
			if !ok {
				panic(err) // Rethrow
			}
			p.errors = append(p.errors, runtimeErr)
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
	return p.expressionStatement()
}

func (p *Parser) printStatement() PrintStmt {
	expr := p.expression()
	p.consume(Semicolon, "expected ';' after expression")
	return PrintStmt{expr}
}

func (p *Parser) expressionStatement() ExpressionStmt {
	expr := p.expression()
	p.consume(Semicolon, "expected ';' after expression")
	return ExpressionStmt{expr}
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
	if p.match(Identifier) {
		return VariableExpr{p.previous()}
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

type parseErrors []parseError

func (errs parseErrors) Error() string {
	if len(errs) == 1 {
		return errs[0].Error()
	}
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = "  " + err.Error()
	}
	return fmt.Sprintf("multiple errors:\n%s", strings.Join(msgs, "\n"))
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
