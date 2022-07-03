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
			Name:  token(lox.Identifier, "a"),
			Value: literal(1.0),
		}},
		{"a = b = 10", lox.AssignmentExpr{
			Name: token(lox.Identifier, "a"),
			Value: lox.AssignmentExpr{
				Name:  token(lox.Identifier, "b"),
				Value: literal(10.0),
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
					Name:  token(lox.Identifier, "b"),
					Value: literal(10.0),
				}},
			},
		}},
		{"if (a) b = 10; else a = 5;", []lox.Stmt{
			lox.IfStmt{
				Condition: variableExpr("a"),
				Then: lox.ExpressionStmt{lox.AssignmentExpr{
					Name:  token(lox.Identifier, "b"),
					Value: literal(10.0),
				}},
				Else: lox.ExpressionStmt{lox.AssignmentExpr{
					Name:  token(lox.Identifier, "a"),
					Value: literal(5.0),
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
			lox.LoopStmt{
				Condition: variableExpr("a"),
				Body: lox.ExpressionStmt{lox.AssignmentExpr{
					Name: token(lox.Identifier, "a"),
					Value: lox.BinaryExpr{
						Left:     variableExpr("a"),
						Operator: token(lox.Minus, "-"),
						Right:    literal(1.0),
					},
				}},
			},
		}},
		{"for (;;) print 42;", []lox.Stmt{
			lox.LoopStmt{
				Condition: literal(true),
				Body:      lox.PrintStmt{literal(42.0)},
			},
		}},
		{"for (var i = 0;;) print i;", []lox.Stmt{
			lox.BlockStmt{[]lox.Stmt{
				lox.VarStmt{Name: token(lox.Identifier, "i"), Init: literal(0.0)},
				lox.LoopStmt{
					Condition: literal(true),
					Body:      lox.PrintStmt{variableExpr("i")},
				},
			}},
		}},
		{"for (; i > 0;) print i;", []lox.Stmt{
			lox.LoopStmt{
				Condition: lox.BinaryExpr{
					Left:     variableExpr("i"),
					Operator: token(lox.Greater, ">"),
					Right:    literal(0.0),
				},
				Body: lox.PrintStmt{variableExpr("i")},
			},
		}},
		{"for (;; i = i+1) print i;", []lox.Stmt{
			lox.LoopStmt{
				Condition: literal(true),
				Body:      lox.PrintStmt{variableExpr("i")},
				OnLoop: lox.AssignmentExpr{
					Name: token(lox.Identifier, "i"),
					Value: lox.BinaryExpr{
						Left:     variableExpr("i"),
						Operator: token(lox.Plus, "+"),
						Right:    literal(1.0),
					},
				},
			},
		}},
		{"for (;; inc) { if (a) continue; continue; }", []lox.Stmt{
			lox.LoopStmt{
				Condition: literal(true),
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
          print "fin";`, `line 1 at '==': expecting ';' after variable declaration
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
		{"f(1, 2;", "line 1 at ';': expecting ')' after arguments"},
		{"f(1, 2,);", "line 1 at ')': expecting expression"},
		{`f(
            "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "0a", "0b", "0c", "0d", "0e", "0f",
            "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "1a", "1b", "1c", "1d", "1e", "1f",
            "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "2a", "2b", "2c", "2d", "2e", "2f",
            "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "3a", "3b", "3c", "3d", "3e", "3f",
            "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4a", "4b", "4c", "4d", "4e", "4f",
            "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "5a", "5b", "5c", "5d", "5e", "5f",
            "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "6a", "6b", "6c", "6d", "6e", "6f",
            "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "7a", "7b", "7c", "7d", "7e", "7f",
            "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "8a", "8b", "8c", "8d", "8e", "8f",
            "90", "91", "92", "93", "94", "95", "96", "97", "98", "99", "9a", "9b", "9c", "9d", "9e", "9f",
            "a0", "a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "aa", "ab", "ac", "ad", "ae", "af",
            "b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9", "ba", "bb", "bc", "bd", "be", "bf",
            "c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "ca", "cb", "cc", "cd", "ce", "cf",
            "d0", "d1", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "da", "db", "dc", "dd", "de", "df",
            "e0", "e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "ea", "eb", "ec", "ed", "ee", "ef",
            "f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "fa", "fb", "fc", "fd", "fe", "ff",
            "after", "limits");`,
			`line 17 at '"ff"': can't have more than 255 arguments`},
		{"fun 100(){}", "line 1 at '100': expecting function name"},
		{"fun f{}", "line 1 at '{': expecting '(' after function name"},
		{"fun f(a,b{}", "line 1 at '{': expecting ')' after params"},
		{"fun f(a,b,){}", "line 1 at ')': expecting parameter name"},
		{"fun f(a,1){};", "line 1 at '1': expecting parameter name"},
		{"fun f(a,b) a;", "line 1 at 'a': expecting '{' before function body"},
		{"fun f(a,b) {a;", "line 1 at end: expecting '}' after block"},
		{`fun f(
            p00, p01, p02, p03, p04, p05, p06, p07, p08, p09, p0a, p0b, p0c, p0d, p0e, p0f,
            p10, p11, p12, p13, p14, p15, p16, p17, p18, p19, p1a, p1b, p1c, p1d, p1e, p1f,
            p20, p21, p22, p23, p24, p25, p26, p27, p28, p29, p2a, p2b, p2c, p2d, p2e, p2f,
            p30, p31, p32, p33, p34, p35, p36, p37, p38, p39, p3a, p3b, p3c, p3d, p3e, p3f,
            p40, p41, p42, p43, p44, p45, p46, p47, p48, p49, p4a, p4b, p4c, p4d, p4e, p4f,
            p50, p51, p52, p53, p54, p55, p56, p57, p58, p59, p5a, p5b, p5c, p5d, p5e, p5f,
            p60, p61, p62, p63, p64, p65, p66, p67, p68, p69, p6a, p6b, p6c, p6d, p6e, p6f,
            p70, p71, p72, p73, p74, p75, p76, p77, p78, p79, p7a, p7b, p7c, p7d, p7e, p7f,
            p80, p81, p82, p83, p84, p85, p86, p87, p88, p89, p8a, p8b, p8c, p8d, p8e, p8f,
            p90, p91, p92, p93, p94, p95, p96, p97, p98, p99, p9a, p9b, p9c, p9d, p9e, p9f,
            pa0, pa1, pa2, pa3, pa4, pa5, pa6, pa7, pa8, pa9, paa, pab, pac, pad, pae, paf,
            pb0, pb1, pb2, pb3, pb4, pb5, pb6, pb7, pb8, pb9, pba, pbb, pbc, pbd, pbe, pbf,
            pc0, pc1, pc2, pc3, pc4, pc5, pc6, pc7, pc8, pc9, pca, pcb, pcc, pcd, pce, pcf,
            pd0, pd1, pd2, pd3, pd4, pd5, pd6, pd7, pd8, pd9, pda, pdb, pdc, pdd, pde, pdf,
            pe0, pe1, pe2, pe3, pe4, pe5, pe6, pe7, pe8, pe9, pea, peb, pec, ped, pee, pef,
            pf0, pf1, pf2, pf3, pf4, pf5, pf6, pf7, pf8, pf9, pfa, pfb, pfc, pfd, pfe, pff,
            after, limits){}`,
			"line 17 at 'pff': can't have more than 255 parameters"},
		{"return;", "line 1 at 'return': 'return' can only be used within functions"},
		{"if (true)\n  return false;", "line 2 at 'return': 'return' can only be used within functions"},
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
