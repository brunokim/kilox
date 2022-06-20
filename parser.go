package lox

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

func (p *Parser) expression() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = Binary{expr, operator, right}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	return nil
}

func (p *Parser) term() Expr {
	return nil
}

func (p *Parser) factor() Expr {
	return nil
}

func (p *Parser) unary() Expr {
	return nil
}

func (p *Parser) primary() Expr {
	return nil
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
