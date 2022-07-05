package lox_test

import (
	"testing"

	"github.com/brunokim/lox"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func token(tokenType lox.TokenType, text string) lox.Token {
	return literalToken(tokenType, text, nil)
}

func literalToken(tokenType lox.TokenType, text string, literal any) lox.Token {
	return lox.Token{
		TokenType: tokenType,
		Lexeme:    text,
		Literal:   literal,
	}
}

func TestScanner(t *testing.T) {
	tests := []struct {
		text string
		want []lox.Token
	}{
		{"a", []lox.Token{token(lox.Identifier, "a"), token(lox.EOF, "")}},
		{" a", []lox.Token{token(lox.Identifier, "a"), token(lox.EOF, "")}},
		{" a ", []lox.Token{token(lox.Identifier, "a"), token(lox.EOF, "")}},
		{"a + 1", []lox.Token{
			token(lox.Identifier, "a"),
			token(lox.Plus, "+"),
			literalToken(lox.Number, "1", 1.0),
			token(lox.EOF, "")}},
		{"!(x and false)", []lox.Token{
			token(lox.Bang, "!"),
			token(lox.LeftParen, "("),
			token(lox.Identifier, "x"),
			token(lox.And, "and"),
			token(lox.False, "false"),
			token(lox.RightParen, ")"),
			token(lox.EOF, ""),
		}},
		{`"abc \" def \\ ghi"`, []lox.Token{
			literalToken(lox.String, `"abc \" def \\ ghi"`, `abc " def \ ghi`),
			token(lox.EOF, ""),
		}},
		{"\"abc\ndef\"", []lox.Token{
			literalToken(lox.String, "\"abc\ndef\"", "abc\ndef"),
			token(lox.EOF, ""),
		}},
	}

	for _, test := range tests {
		s := lox.NewScanner(test.text)
		tokens, err := s.ScanTokens()
		if err != nil {
			t.Fatalf("want nil, got err: %v", err)
		}
		opts := cmp.Options{cmpopts.IgnoreFields(lox.Token{}, "Line")}
		if d := cmp.Diff(test.want, tokens, opts); d != "" {
			t.Errorf("(-want, +got)%s", d)
		}
	}
}

func TestScannerError(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"a % b", "line 1: unexpected character: %"},
		{"a \n% \nb", "line 2: unexpected character: %"},
		{`"unterminated`, "line 1: unterminated string"},
		{`"unterminated`, "line 1: unterminated string"},
		{`"unterminated

        `, "line 3: unterminated string"},
		{`"invalid \n escape"`, "line 1: invalid escaped character 'n' in string"},
		{`a ^ b
         "str\b
         `, `line 1: unexpected character: ^
line 2: invalid escaped character 'b' in string
line 3: unterminated string`},
	}

	for _, test := range tests {
		s := lox.NewScanner(test.text)
		tokens, err := s.ScanTokens()
		if err == nil {
			t.Fatalf("want err, got tokens %v for text %q", tokens, test.text)
		}
		if d := cmp.Diff(test.want, err.Error()); d != "" {
			t.Errorf("(-want, +got)%s", d)
		}
	}
}
