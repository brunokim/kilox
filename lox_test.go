package lox_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/brunokim/lox"
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
		c := lox.NewTypeChecker(i)
		err := c.Check(stmts)
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	i.SetStdout(&b)
	err = i.Interpret(stmts)
	return b.String(), err
}

func parse(t *testing.T, text string) ([]lox.Stmt, error) {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		t.Fatalf("%q: want nil, got err: %v", text, err)
	}
	p := lox.NewParser(tokens)
	return p.Parse()
}

func parseStmts(t *testing.T, text string) []lox.Stmt {
	stmts, err := parse(t, text)
	if err != nil {
		t.Fatalf("%q: want nil, got err: %v", text, err)
	}
	return stmts
}

func parseExpr(t *testing.T, text string) lox.Expr {
	stmts := parseStmts(t, text+";")
	if len(stmts) != 1 {
		t.Log(stmts)
		t.Fatalf("%q: expecting a single statement, got %d", text, len(stmts))
	}
	exprStmt, ok := stmts[0].(lox.ExpressionStmt)
	if !ok {
		t.Log(stmts[0])
		t.Fatalf("%q: expecting an expression statement, got %T", text, stmts[0])
	}
	return exprStmt.Expression
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
