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

func variableExpr(name string) lox.VariableExpr {
	return lox.VariableExpr{Name: token(lox.Identifier, name)}
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
		{"x", variableExpr("x")},
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
			Target: variableExpr("a"),
			Value:  literal(1.0),
		}},
		{"a = b = 10", lox.AssignmentExpr{
			Target: variableExpr("a"),
			Value: lox.AssignmentExpr{
				Target: variableExpr("b"),
				Value:  literal(10.0),
			},
		}},
		{"a and b or c or d and e", lox.LogicExpr{
			Left: lox.LogicExpr{
				Left: lox.LogicExpr{
					Left:     variableExpr("a"),
					Operator: token(lox.And, "and"),
					Right:    variableExpr("b"),
				},
				Operator: token(lox.Or, "or"),
				Right:    variableExpr("c"),
			},
			Operator: token(lox.Or, "or"),
			Right: lox.LogicExpr{
				Left:     variableExpr("d"),
				Operator: token(lox.And, "and"),
				Right:    variableExpr("e"),
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
			Left:     variableExpr("a"),
			Operator: token(lox.Plus, "+"),
			Right:    literal(2.0),
		}}}},
		{"print a; print b;", []lox.Stmt{
			lox.PrintStmt{variableExpr("a")},
			lox.PrintStmt{variableExpr("b")},
		}},
		{"var a; var b = false;", []lox.Stmt{
			lox.VarStmt{Name: token(lox.Identifier, "a")},
			lox.VarStmt{Name: token(lox.Identifier, "b"), Init: literal(false)},
		}},
		{"if (a) b = 10;", []lox.Stmt{
			lox.IfStmt{
				Condition: variableExpr("a"),
				Then: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: variableExpr("b"),
					Value:  literal(10.0),
				}},
			},
		}},
		{"if (a) b = 10; else a = 5;", []lox.Stmt{
			lox.IfStmt{
				Condition: variableExpr("a"),
				Then: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: variableExpr("b"),
					Value:  literal(10.0),
				}},
				Else: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: variableExpr("a"),
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
		{"while (a) a = a - 1;", []lox.Stmt{
			lox.WhileStmt{
				Condition: variableExpr("a"),
				Body: lox.ExpressionStmt{lox.AssignmentExpr{
					Target: variableExpr("a"),
					Value: lox.BinaryExpr{
						Left:     variableExpr("a"),
						Operator: token(lox.Minus, "-"),
						Right:    literal(1.0),
					},
				}},
			},
		}},
		{"for (;;) print 42;", []lox.Stmt{
			lox.WhileStmt{
				Condition: literal(true),
				Body:      lox.PrintStmt{literal(42.0)},
			},
		}},
		{"for (var i = 0;;) print i;", []lox.Stmt{
			lox.BlockStmt{[]lox.Stmt{
				lox.VarStmt{Name: token(lox.Identifier, "i"), Init: literal(0.0)},
				lox.WhileStmt{
					Condition: literal(true),
					Body:      lox.PrintStmt{variableExpr("i")},
				},
			}},
		}},
		{"for (; i > 0;) print i;", []lox.Stmt{
			lox.WhileStmt{
				Condition: lox.BinaryExpr{
					Left:     variableExpr("i"),
					Operator: token(lox.Greater, ">"),
					Right:    literal(0.0),
				},
				Body: lox.PrintStmt{variableExpr("i")},
			},
		}},
		{"for (;; i = i+1) print i;", []lox.Stmt{
			lox.WhileStmt{
				Condition: literal(true),
				Body: lox.BlockStmt{[]lox.Stmt{
					lox.PrintStmt{variableExpr("i")},
					lox.ExpressionStmt{lox.AssignmentExpr{
						Target: variableExpr("i"),
						Value: lox.BinaryExpr{
							Left:     variableExpr("i"),
							Operator: token(lox.Plus, "+"),
							Right:    literal(1.0),
						},
					}},
				}},
			},
		}},
		{"for (;; inc) { if (a) continue; continue; }", []lox.Stmt{
			lox.WhileStmt{
				Condition: literal(true),
				Body: lox.BlockStmt{[]lox.Stmt{
					lox.BlockStmt{[]lox.Stmt{
						lox.IfStmt{
							Condition: variableExpr("a"),
							Then: lox.BlockStmt{[]lox.Stmt{
								lox.ExpressionStmt{variableExpr("inc")},
								lox.ContinueStmt{token(lox.Continue, "continue")},
							}},
						},
						lox.BlockStmt{[]lox.Stmt{
							lox.ExpressionStmt{variableExpr("inc")},
							lox.ContinueStmt{token(lox.Continue, "continue")},
						}},
					}},
					lox.ExpressionStmt{variableExpr("inc")},
				}},
			},
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
		{"if (a then b = 10;", "line 1 at 'then': expecting ')' after condition"},
		{"if (a) then b = 10;", "line 1 at 'b': expecting ';' after expression"},
		{"if (a) b = 10; else", "line 1 at end: expecting expression"},
		{"if (a) var b = 10;", "line 1 at 'var': expecting expression"},
		{"{1;}; 2;", "line 1 at ';': expecting expression"},
		{"{1; 2;", "line 1 at end: expecting '}' after block"},
		{"while a do a = a - 1;", "line 1 at 'a': expecting '(' after 'while'"},
		{"while (a do a = a - 1;", "line 1 at 'do': expecting ')' after condition"},
		{"while (a) do a = a - 1;", "line 1 at 'a': expecting ';' after expression"},
		{"var a; break;", "line 1 at 'break': 'break' can only be used within loops"},
		{"{var a; continue;}", "line 1 at 'continue': 'continue' can only be used within loops"},
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
