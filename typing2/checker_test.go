package typing2_test

import (
	"testing"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/typing2"

	"github.com/google/go-cmp/cmp"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		text string
		want []lox.ClauseTerm
	}{
		{
			"fun foo() {}",
			clauses(
				clause(
					functor("type", atom("foo"), functor("fun", list(), var_("Ret"))),
					functor("=", atom("nil"), var_("Ret")))),
		},
		{
			"fun ans() { return 42; }",
			clauses(
				clause(
					functor("type", atom("ans"), functor("fun", list(), var_("Ret"))),
					functor("=", atom("number"), var_("Ret")))),
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			s := lox.NewScanner(test.text)
			tokens, err := s.ScanTokens()
			if err != nil {
				t.Fatalf("scanner: %v", err)
			}
			p := lox.NewParser(tokens)
			stmts, err := p.Parse()
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			c := typing2.NewChecker()
			clauses, err := c.Build(stmts)
			if err != nil {
				t.Errorf("got err: %v", err)
				return
			}
			if diff := cmp.Diff(clauses, test.want); diff != "" {
				t.Errorf("(-want,+got)\n%s", diff)
			}
		})
	}
}
