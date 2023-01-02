package lox_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/brunokim/kilox"
	"github.com/brunokim/kilox/typing"
)

func runLox(text string, experiments map[string]bool) (string, error) {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		return "", err
	}
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		return "", err
	}
	i := lox.NewInterpreter()
	r := lox.NewResolver(i)
	err = r.Resolve(stmts)
	if err != nil {
		return "", err
	}
	if experiments["typing"] {
		c := typing.NewChecker()
		_, err := c.Check(stmts)
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	i.SetStdout(&b)
	err = i.Interpret(stmts)
	return b.String(), err
}

func parser(t *testing.T, text string) *lox.Parser {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		t.Fatalf("parser(%q): %v", text, err)
	}
	return lox.NewParser(tokens)
}

func parseStmts(t *testing.T, text string) []lox.Stmt {
	p := parser(t, text)
	stmts, err := p.Parse()
	if err != nil {
		t.Fatalf("parseStmts(%q): %v", text, err)
	}
	return stmts
}

func parseExpr(t *testing.T, text string) lox.Expr {
	p := parser(t, text)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("parseExpr(%q): %v", text, err)
	}
	return expr
}

// ---- scanner test

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

// ---- parser test

func literal(tokenType lox.TokenType, v any) *lox.LiteralExpr {
	return &lox.LiteralExpr{Token: literalToken(tokenType, "", v), Value: v}
}

func number(x float64) *lox.LiteralExpr {
	return &lox.LiteralExpr{Token: literalToken(lox.Number, "", x), Value: x}
}

func boolean(x bool) *lox.LiteralExpr {
	if x {
		return &lox.LiteralExpr{Token: token(lox.True, ""), Value: true}
	}
	return &lox.LiteralExpr{Token: token(lox.False, ""), Value: false}
}

func variableExpr(name string) *lox.VariableExpr {
	return &lox.VariableExpr{Name: token(lox.Identifier, name)}
}

// ---- interpreter test

func extractExpected(text string) (string, string) {
	wantOutput := extractComment(text, "output")
	wantError := extractComment(text, "error")
	return wantOutput, wantError
}

func extractComment(text, pattern string) string {
	commentRE := regexp.MustCompile("(?im)// " + pattern + ":(.*)$")

	var b strings.Builder
	matches := commentRE.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		b.WriteString(strings.TrimPrefix(match[1], " "))
		b.WriteRune('\n')
	}
	return b.String()
}

func extractExperiments(text string) map[string]bool {
	expStr := extractComment(text, "experiments")
	exps := make(map[string]bool)
	for _, exp := range strings.Split(expStr, ",") {
		exp = strings.TrimSpace(exp)
		if exp == "" {
			continue
		}
		// Last setting wins.
		switch exp[0] {
		case '+':
			exps[exp[1:]] = true
		case '-':
			exps[exp[1:]] = false
		default:
			exps[exp] = true
		}
	}
	return exps
}
