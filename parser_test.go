package lox_test

import (
	"testing"

	"github.com/brunokim/lox"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func literal(v interface{}) lox.LiteralExpr {
	return lox.LiteralExpr{Value: v}
}

func TestParserExpression(t *testing.T) {
	tests := []struct {
		text string
		want lox.Expr
	}{
		{"10", literal(10.0)},
		{"10.25", literal(10.25)},
		{"false", literal(false)},
		{"true", literal(true)},
		{"nil", literal(nil)},
		{`"abc def"`, literal("abc def")},
		{"x", lox.VariableExpr{Name: token(lox.Identifier, "x")}},
		{"-1", lox.UnaryExpr{Operator: token(lox.Minus, "-"), Right: literal(1.0)}},
		{"(1)", lox.GroupingExpr{Expression: literal(1.0)}},
		{"2-1", lox.BinaryExpr{
			Left:     literal(2.0),
			Operator: token(lox.Minus, "-"),
			Right:    literal(1.0),
		}},
		{"3-2-1", lox.BinaryExpr{
			Left: lox.BinaryExpr{
				Left:     literal(3.0),
				Operator: token(lox.Minus, "-"),
				Right:    literal(2.0),
			},
			Operator: token(lox.Minus, "-"),
			Right:    literal(1.0),
		}},
		{"3*2-1", lox.BinaryExpr{
			Left: lox.BinaryExpr{
				Left:     literal(3.0),
				Operator: token(lox.Star, "*"),
				Right:    literal(2.0),
			},
			Operator: token(lox.Minus, "-"),
			Right:    literal(1.0),
		}},
		{"3-2*1", lox.BinaryExpr{
			Left:     literal(3.0),
			Operator: token(lox.Minus, "-"),
			Right: lox.BinaryExpr{
				Left:     literal(2.0),
				Operator: token(lox.Star, "*"),
				Right:    literal(1.0),
			},
		}},
		{"4*3-2*1", lox.BinaryExpr{
			Left: lox.BinaryExpr{
				Left:     literal(4.0),
				Operator: token(lox.Star, "*"),
				Right:    literal(3.0),
			},
			Operator: token(lox.Minus, "-"),
			Right: lox.BinaryExpr{
				Left:     literal(2.0),
				Operator: token(lox.Star, "*"),
				Right:    literal(1.0),
			},
		}},
		{"4*(3-2)*1", lox.BinaryExpr{
			Left: lox.BinaryExpr{
				Left:     literal(4.0),
				Operator: token(lox.Star, "*"),
				Right: lox.GroupingExpr{
					Expression: lox.BinaryExpr{
						Left:     literal(3.0),
						Operator: token(lox.Minus, "-"),
						Right:    literal(2.0),
					},
				},
			},
			Operator: token(lox.Star, "*"),
			Right:    literal(1.0),
		}},
		{"1 != 2 > 3 + 4 / !!5", lox.BinaryExpr{
			Left:     literal(1.0),
			Operator: token(lox.BangEqual, "!="),
			Right: lox.BinaryExpr{
				Left:     literal(2.0),
				Operator: token(lox.Greater, ">"),
				Right: lox.BinaryExpr{
					Left:     literal(3.0),
					Operator: token(lox.Plus, "+"),
					Right: lox.BinaryExpr{
						Left:     literal(4.0),
						Operator: token(lox.Slash, "/"),
						Right: lox.UnaryExpr{
							Operator: token(lox.Bang, "!"),
							Right: lox.UnaryExpr{
								Operator: token(lox.Bang, "!"),
								Right:    literal(5.0),
							},
						},
					},
				},
			},
		}},
	}

	for _, test := range tests {
		got := parseExpr(t, test.text)
		opt := cmpopts.IgnoreFields(lox.Token{}, "Line")
		if diff := cmp.Diff(test.want, got, opt); diff != "" {
			t.Errorf("%s: (-want, +got)%s", test.text, diff)
		}
	}
}

// -----

func parseStmts(t *testing.T, text string) []lox.Stmt {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		t.Fatalf("%q: want nil, got err: %v", text, err)
	}
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
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
