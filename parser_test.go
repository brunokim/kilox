package lox_test

import (
	"testing"

	"github.com/brunokim/lox"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParserExpression(t *testing.T) {
	tests := []struct {
		text string
		want lox.Expr
	}{
		{"10", number(10)},
		{"10.25", number(10.25)},
		{"false", boolean(false)},
		{"true", boolean(true)},
		{"nil", literal(lox.Nil, nil)},
		{"this", &lox.ThisExpr{token(lox.This, "this")}},
		{`"abc def"`, literal(lox.String, "abc def")},
		{"x", variableExpr("x")},
		{"-1", &lox.UnaryExpr{Operator: token(lox.Minus, "-"), Right: number(1)}},
		{"(1)", &lox.GroupingExpr{Expression: number(1)}},
		{"2-1", &lox.BinaryExpr{
			Left:     number(2),
			Operator: token(lox.Minus, "-"),
			Right:    number(1),
		}},
		{"3-2-1", &lox.BinaryExpr{
			Left: &lox.BinaryExpr{
				Left:     number(3),
				Operator: token(lox.Minus, "-"),
				Right:    number(2),
			},
			Operator: token(lox.Minus, "-"),
			Right:    number(1),
		}},
		{"3*2-1", &lox.BinaryExpr{
			Left: &lox.BinaryExpr{
				Left:     number(3),
				Operator: token(lox.Star, "*"),
				Right:    number(2),
			},
			Operator: token(lox.Minus, "-"),
			Right:    number(1),
		}},
		{"3-2*1", &lox.BinaryExpr{
			Left:     number(3),
			Operator: token(lox.Minus, "-"),
			Right: &lox.BinaryExpr{
				Left:     number(2),
				Operator: token(lox.Star, "*"),
				Right:    number(1),
			},
		}},
		{"4*3-2*1", &lox.BinaryExpr{
			Left: &lox.BinaryExpr{
				Left:     number(4),
				Operator: token(lox.Star, "*"),
				Right:    number(3),
			},
			Operator: token(lox.Minus, "-"),
			Right: &lox.BinaryExpr{
				Left:     number(2),
				Operator: token(lox.Star, "*"),
				Right:    number(1),
			},
		}},
		{"4*(3-2)*1", &lox.BinaryExpr{
			Left: &lox.BinaryExpr{
				Left:     number(4),
				Operator: token(lox.Star, "*"),
				Right: &lox.GroupingExpr{
					Expression: &lox.BinaryExpr{
						Left:     number(3),
						Operator: token(lox.Minus, "-"),
						Right:    number(2),
					},
				},
			},
			Operator: token(lox.Star, "*"),
			Right:    number(1),
		}},
		{"1 != 2 > 3 + 4 / !!5", &lox.BinaryExpr{
			Left:     number(1),
			Operator: token(lox.BangEqual, "!="),
			Right: &lox.BinaryExpr{
				Left:     number(2),
				Operator: token(lox.Greater, ">"),
				Right: &lox.BinaryExpr{
					Left:     number(3),
					Operator: token(lox.Plus, "+"),
					Right: &lox.BinaryExpr{
						Left:     number(4),
						Operator: token(lox.Slash, "/"),
						Right: &lox.UnaryExpr{
							Operator: token(lox.Bang, "!"),
							Right: &lox.UnaryExpr{
								Operator: token(lox.Bang, "!"),
								Right:    number(5),
							},
						},
					},
				},
			},
		}},
		{"a = 1", &lox.AssignmentExpr{
			Name:  token(lox.Identifier, "a"),
			Value: number(1),
		}},
		{"a = b = 10", &lox.AssignmentExpr{
			Name: token(lox.Identifier, "a"),
			Value: &lox.AssignmentExpr{
				Name:  token(lox.Identifier, "b"),
				Value: number(10),
			},
		}},
		{"a and b or c or d and e", &lox.LogicExpr{
			Left: &lox.LogicExpr{
				Left: &lox.LogicExpr{
					Left:     variableExpr("a"),
					Operator: token(lox.And, "and"),
					Right:    variableExpr("b"),
				},
				Operator: token(lox.Or, "or"),
				Right:    variableExpr("c"),
			},
			Operator: token(lox.Or, "or"),
			Right: &lox.LogicExpr{
				Left:     variableExpr("d"),
				Operator: token(lox.And, "and"),
				Right:    variableExpr("e"),
			},
		}},
		{"foo(1).bar.baz().qux", &lox.GetExpr{
			Object: &lox.CallExpr{
				Callee: &lox.GetExpr{
					Object: &lox.GetExpr{
						Object: &lox.CallExpr{
							Callee: variableExpr("foo"),
							Args:   []lox.Expr{number(1)},
							Paren:  token(lox.RightParen, ")"),
						},
						Name: token(lox.Identifier, "bar"),
					},
					Name: token(lox.Identifier, "baz"),
				},
				Paren: token(lox.RightParen, ")"),
			},
			Name: token(lox.Identifier, "qux"),
		}},
		{"foo(a).bar = 10", &lox.SetExpr{
			Object: &lox.CallExpr{
				Callee: variableExpr("foo"),
				Args:   []lox.Expr{variableExpr("a")},
				Paren:  token(lox.RightParen, ")"),
			},
			Name:  token(lox.Identifier, "bar"),
			Value: number(10),
		}},
	}

	for _, test := range tests {
		got := parseExpr(t, test.text)
		opts := cmp.Options{
			cmpopts.IgnoreFields(lox.Token{}, "Line"),
			cmpopts.IgnoreFields(lox.LiteralExpr{}, "Token.Lexeme"),
		}
		if diff := cmp.Diff(test.want, got, opts); diff != "" {
			t.Errorf("%s: (-want, +got)%s", test.text, diff)
		}
	}
}

func TestParserStatements(t *testing.T) {
	tests := []struct {
		text string
		want []lox.Stmt
	}{
		{"a+2;", []lox.Stmt{lox.ExpressionStmt{&lox.BinaryExpr{
			Left:     variableExpr("a"),
			Operator: token(lox.Plus, "+"),
			Right:    number(2),
		}}}},
		{"print a; print b;", []lox.Stmt{
			lox.PrintStmt{variableExpr("a")},
			lox.PrintStmt{variableExpr("b")},
		}},
		{"var a; var b = false;", []lox.Stmt{
			lox.VarStmt{Name: token(lox.Identifier, "a")},
			lox.VarStmt{Name: token(lox.Identifier, "b"), Init: boolean(false)},
		}},
		{"if (a) b = 10;", []lox.Stmt{
			lox.IfStmt{
				Condition: variableExpr("a"),
				Then: lox.ExpressionStmt{&lox.AssignmentExpr{
					Name:  token(lox.Identifier, "b"),
					Value: number(10),
				}},
			},
		}},
		{"if (a) b = 10; else a = 5;", []lox.Stmt{
			lox.IfStmt{
				Condition: variableExpr("a"),
				Then: lox.ExpressionStmt{&lox.AssignmentExpr{
					Name:  token(lox.Identifier, "b"),
					Value: number(10),
				}},
				Else: lox.ExpressionStmt{&lox.AssignmentExpr{
					Name:  token(lox.Identifier, "a"),
					Value: number(5),
				}},
			},
		}},
		{"1; {2; {3; 4; {}} 5;} {6;}", []lox.Stmt{
			lox.ExpressionStmt{number(1)},
			lox.BlockStmt{[]lox.Stmt{
				lox.ExpressionStmt{number(2)},
				lox.BlockStmt{[]lox.Stmt{
					lox.ExpressionStmt{number(3)},
					lox.ExpressionStmt{number(4)},
					lox.BlockStmt{},
				}},
				lox.ExpressionStmt{number(5)},
			}},
			lox.BlockStmt{[]lox.Stmt{
				lox.ExpressionStmt{number(6)},
			}},
		}},
		{"while (a) a = a - 1;", []lox.Stmt{
			lox.LoopStmt{
				Condition: variableExpr("a"),
				Body: lox.ExpressionStmt{&lox.AssignmentExpr{
					Name: token(lox.Identifier, "a"),
					Value: &lox.BinaryExpr{
						Left:     variableExpr("a"),
						Operator: token(lox.Minus, "-"),
						Right:    number(1),
					},
				}},
			},
		}},
		{"for (;;) print 42;", []lox.Stmt{
			lox.LoopStmt{
				Condition: &lox.LiteralExpr{token(lox.Semicolon, ";"), true},
				Body:      lox.PrintStmt{number(42)},
			},
		}},
		{"for (var i = 0;;) print i;", []lox.Stmt{
			lox.BlockStmt{[]lox.Stmt{
				lox.VarStmt{Name: token(lox.Identifier, "i"), Init: number(0)},
				lox.LoopStmt{
					Condition: &lox.LiteralExpr{token(lox.Semicolon, ";"), true},
					Body:      lox.PrintStmt{variableExpr("i")},
				},
			}},
		}},
		{"for (; i > 0;) print i;", []lox.Stmt{
			lox.LoopStmt{
				Condition: &lox.BinaryExpr{
					Left:     variableExpr("i"),
					Operator: token(lox.Greater, ">"),
					Right:    number(0),
				},
				Body: lox.PrintStmt{variableExpr("i")},
			},
		}},
		{"for (;; i = i+1) print i;", []lox.Stmt{
			lox.LoopStmt{
				Condition: &lox.LiteralExpr{token(lox.Semicolon, ";"), true},
				Body:      lox.PrintStmt{variableExpr("i")},
				OnLoop: &lox.AssignmentExpr{
					Name: token(lox.Identifier, "i"),
					Value: &lox.BinaryExpr{
						Left:     variableExpr("i"),
						Operator: token(lox.Plus, "+"),
						Right:    number(1),
					},
				},
			},
		}},
		{"for (;; inc) { if (a) continue; continue; }", []lox.Stmt{
			lox.LoopStmt{
				Condition: &lox.LiteralExpr{token(lox.Semicolon, ";"), true},
				Body: lox.BlockStmt{[]lox.Stmt{
					lox.IfStmt{
						Condition: variableExpr("a"),
						Then:      lox.ContinueStmt{token(lox.Continue, "continue")},
					},
					lox.ContinueStmt{token(lox.Continue, "continue")},
				}},
				OnLoop: variableExpr("inc"),
			},
		}},
		{"class Empty {}", []lox.Stmt{
			lox.ClassStmt{
				Name: token(lox.Identifier, "Empty"),
			},
		}},
		{"class BigThinker { answer() {print 42;} }", []lox.Stmt{
			lox.ClassStmt{
				Name: token(lox.Identifier, "BigThinker"),
				Methods: []lox.FunctionStmt{
					{
						Name: token(lox.Identifier, "answer"),
						Body: []lox.Stmt{
							lox.PrintStmt{number(42)},
						},
					},
				},
			},
		}},
		{"class Foo { bar(x){} baz(y, z){} }", []lox.Stmt{
			lox.ClassStmt{
				Name: token(lox.Identifier, "Foo"),
				Methods: []lox.FunctionStmt{
					{
						Name:   token(lox.Identifier, "bar"),
						Params: []lox.Token{token(lox.Identifier, "x")},
					},
					{
						Name: token(lox.Identifier, "baz"),
						Params: []lox.Token{
							token(lox.Identifier, "y"),
							token(lox.Identifier, "z"),
						},
					},
				},
			},
		}},
		{"class Foo { f(){} class g(){} class h(){} i(){} }", []lox.Stmt{
			lox.ClassStmt{
				Name: token(lox.Identifier, "Foo"),
				Methods: []lox.FunctionStmt{
					{Name: token(lox.Identifier, "f")},
					{Name: token(lox.Identifier, "i")},
				},
				StaticMethods: []lox.FunctionStmt{
					{Name: token(lox.Identifier, "g")},
					{Name: token(lox.Identifier, "h")},
				},
			},
		}},
		{"class Foo { f(){} class var g; class h(){} var i; }", []lox.Stmt{
			lox.ClassStmt{
				Name: token(lox.Identifier, "Foo"),
				Methods: []lox.FunctionStmt{
					{Name: token(lox.Identifier, "f")},
				},
				StaticMethods: []lox.FunctionStmt{
					{Name: token(lox.Identifier, "h")},
				},
				Vars: []lox.VarStmt{
					{Name: token(lox.Identifier, "i")},
				},
				StaticVars: []lox.VarStmt{
					{Name: token(lox.Identifier, "g")},
				},
			},
		}},
	}

	for _, test := range tests {
		got := parseStmts(t, test.text)
		opts := cmp.Options{
			cmpopts.IgnoreFields(lox.Token{}, "Line"),
			cmpopts.IgnoreFields(lox.LiteralExpr{}, "Token.Lexeme"),
		}
		if diff := cmp.Diff(test.want, got, opts); diff != "" {
			t.Errorf("%s: (-want, +got)%s", test.text, diff)
		}
	}
}
