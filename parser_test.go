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
		{"a = 1", lox.AssignmentExpr{
			Target: lox.VariableExpr{token(lox.Identifier, "a")},
			Value:  literal(1.0),
		}},
		{"a = b = 10", lox.AssignmentExpr{
			Target: lox.VariableExpr{token(lox.Identifier, "a")},
			Value: lox.AssignmentExpr{
				Target: lox.VariableExpr{token(lox.Identifier, "b")},
				Value:  literal(10.0),
			},
		}},
		{"a and b or c or d and e", lox.LogicExpr{
			Left: lox.LogicExpr{
				Left: lox.LogicExpr{
					Left:     lox.VariableExpr{token(lox.Identifier, "a")},
					Operator: token(lox.And, "and"),
					Right:    lox.VariableExpr{token(lox.Identifier, "b")},
				},
				Operator: token(lox.Or, "or"),
				Right:    lox.VariableExpr{token(lox.Identifier, "c")},
			},
			Operator: token(lox.Or, "or"),
			Right: lox.LogicExpr{
				Left:     lox.VariableExpr{token(lox.Identifier, "d")},
				Operator: token(lox.And, "and"),
				Right:    lox.VariableExpr{token(lox.Identifier, "e")},
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

func TestParserStatements(t *testing.T) {
	tests := []struct {
		text string
		want []lox.Stmt
	}{
		{"a+2;", []lox.Stmt{lox.ExpressionStmt{lox.BinaryExpr{
			Left:     lox.VariableExpr{token(lox.Identifier, "a")},
			Operator: token(lox.Plus, "+"),
			Right:    literal(2.0),
		}}}},
		{"print a; print b;", []lox.Stmt{
			lox.PrintStmt{lox.VariableExpr{token(lox.Identifier, "a")}},
			lox.PrintStmt{lox.VariableExpr{token(lox.Identifier, "b")}},
		}},
		{"var a; var b = false;", []lox.Stmt{
			lox.VarStmt{Name: token(lox.Identifier, "a")},
			lox.VarStmt{Name: token(lox.Identifier, "b"), Init: literal(false)},
		}},
		{"if (a) b = 10;", []lox.Stmt{
			lox.IfStmt{
				Condition: lox.VariableExpr{Name: token(lox.Identifier, "a")},
				Then: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: lox.VariableExpr{token(lox.Identifier, "b")},
					Value:  literal(10.0),
				}},
			},
		}},
		{"if (a) b = 10; else a = 5;", []lox.Stmt{
			lox.IfStmt{
				Condition: lox.VariableExpr{Name: token(lox.Identifier, "a")},
				Then: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: lox.VariableExpr{token(lox.Identifier, "b")},
					Value:  literal(10.0),
				}},
				Else: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: lox.VariableExpr{token(lox.Identifier, "a")},
					Value:  literal(5.0),
				}},
			},
		}},
		{"1; {2; {3; 4; {}} 5;} {6;}", []lox.Stmt{
			lox.ExpressionStmt{literal(1.0)},
			lox.BlockStmt{[]lox.Stmt{
				lox.ExpressionStmt{literal(2.0)},
				lox.BlockStmt{[]lox.Stmt{
					lox.ExpressionStmt{literal(3.0)},
					lox.ExpressionStmt{literal(4.0)},
					lox.BlockStmt{},
				}},
				lox.ExpressionStmt{literal(5.0)},
			}},
			lox.BlockStmt{[]lox.Stmt{
				lox.ExpressionStmt{literal(6.0)},
			}},
		}},
	}

	for _, test := range tests {
		got := parseStmts(t, test.text)
		opt := cmpopts.IgnoreFields(lox.Token{}, "Line")
		if diff := cmp.Diff(test.want, got, opt); diff != "" {
			t.Errorf("%s: (-want, +got)%s", test.text, diff)
		}
	}
}

func TestParserError(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"!)", "line 1 at ')': expecting expression"},
		{"(1", "line 1 at end: expecting ')' after expression"},
		{"var 1 = 2;", "line 1 at '1': expecting variable name"},
		{"1 + 2", "line 1 at end: expecting ';' after expression"},
		{"print 1 + 2", "line 1 at end: expecting ';' after expression"},
		{"var a", "line 1 at end: expecting ';' after variable declaration"},
		{"var a = 1", "line 1 at end: expecting ';' after variable declaration"},
		{`var a == 2
          print 1, !false ;
          (a + b
          print "fin";`, `multiple errors:
  line 1 at '==': expecting ';' after variable declaration
  line 2 at ',': expecting ';' after expression
  line 4 at 'print': expecting ')' after expression`},
		{"(a) = 1;", "line 1 at '=': invalid target for assignment: want variable, got lox.GroupingExpr"},
		{"if a then b = 10;", "line 1 at 'a': expecting '(' after 'if'"},
		{"if (a then b = 10;", "line 1 at 'then': expecting ')' after expression"},
		{"if (a) then b = 10;", "line 1 at 'b': expecting ';' after expression"},
		{"if (a) b = 10; else", "line 1 at end: expecting expression"},
		{"if (a) var b = 10;", "line 1 at 'var': expecting expression"},
	}

	for _, test := range tests {
		stmts, err := parse(t, test.text)
		if err == nil {
			t.Fatalf("%q: want err, got stmts: %v", test.text, stmts)
		}
		if diff := cmp.Diff(test.want, err.Error()); diff != "" {
			t.Errorf("%q: (-want, +got)%s", test.text, diff)
		}
	}
}

// -----

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
