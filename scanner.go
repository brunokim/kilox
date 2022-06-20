package lox

import (
	"fmt"
	"strconv"
)

var keywords = map[string]TokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"for":    For,
	"fun":    Fun,
	"if":     If,
	"nil":    Nil,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}

type Scanner struct {
	source  string
	tokens  []Token
	start   int
	current int
	line    int
	errors  []scanErr
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: source,
		line:   1,
	}
}

func (s *Scanner) ScanTokens() []Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}
	s.tokens = append(s.tokens, Token{Eof, "", nil, s.line})
	return s.tokens
}

func (s *Scanner) scanToken() {
	ch := s.advance()
	switch ch {
	// One-character token
	case '(':
		s.addToken(LeftParen)
	case ')':
		s.addToken(RightParen)
	case '{':
		s.addToken(LeftBrace)
	case '}':
		s.addToken(RightBrace)
	case ',':
		s.addToken(Comma)
	case '.':
		s.addToken(Dot)
	case '-':
		s.addToken(Minus)
	case '+':
		s.addToken(Plus)
	case ';':
		s.addToken(Semicolon)
	case '*':
		s.addToken(Star)
	// One or two character tokens
	case '!':
		tokenType := Bang
		if s.match('=') {
			tokenType = BangEqual
		}
		s.addToken(tokenType)
	case '=':
		tokenType := Equal
		if s.match('=') {
			tokenType = EqualEqual
		}
		s.addToken(tokenType)
	case '<':
		tokenType := Less
		if s.match('=') {
			tokenType = LessEqual
		}
		s.addToken(tokenType)
	case '>':
		tokenType := Greater
		if s.match('=') {
			tokenType = GreaterEqual
		}
		s.addToken(tokenType)
	// Slash
	case '/':
		if s.match('/') {
			// Line comment
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(Slash)
		}
	// Whitespace
	case ' ', '\r', '\t':
		// Do nothing
	case '\n':
		s.line++
	// String start
	case '"':
		s.readString()
	default:
		if isDigit(ch) {
			s.readNumber()
		} else if isAlpha(ch) {
			s.readIdentifier()
		} else {
			s.addError(s.line, fmt.Sprintf("unexpected character: %c", ch))
		}
	}
}

// ----

func (s *Scanner) readIdentifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}
	text := s.source[s.start:s.current]
	tokenType, ok := keywords[text]
	if !ok {
		tokenType = Identifier
	}
	s.addToken(tokenType)
}

func (s *Scanner) readNumber() {
	for isDigit(s.peek()) {
		s.advance()
	}
	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance() // Consume the '.'
		for isDigit(s.peek()) {
			s.advance()
		}
	}

	text := s.source[s.start:s.current]
	f, _ := strconv.ParseFloat(text, 64)
	s.addLiteralToken(Number, f)
}

func (s *Scanner) readString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	if s.isAtEnd() {
		s.addError(s.line, fmt.Sprintf("unterminated string"))
	}
	s.advance()                                // Consume the final '"'.
	value := s.source[s.start+1 : s.current-1] // Trim the surrounding quotes.
	s.addLiteralToken(String, value)
}

// ----

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() rune {
	s.current++
	return rune(s.source[s.current-1])
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if rune(s.source[s.current]) != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}
	return rune(s.source[s.current])
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return rune(s.source[s.current+1])
}

func (s *Scanner) addToken(tokenType TokenType) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, Token{tokenType, text, nil, s.line})
}

func (s *Scanner) addLiteralToken(tokenType TokenType, literal interface{}) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, Token{tokenType, text, literal, s.line})
}

// ----

type scanErr struct {
	line int
	msg  string
}

func (s *Scanner) addError(line int, msg string) {
	s.errors = append(s.errors, scanErr{line, msg})
}

// ----

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_'
}

func isAlphaNumeric(ch rune) bool {
	return isDigit(ch) || isAlpha(ch)
}
