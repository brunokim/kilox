package typing_test

import (
	"testing"

	"github.com/brunokim/lox/typing"

	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestBuildClauses(t *testing.T) {
	tests := []struct {
		text string
		want []typing.TypeClause
	}{
		{
			"fun foo() {}",
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(1), nil_))),
		},
		{
			"fun answer() { return 42; }",
			clauses_(clause_("answer", func_(types_(), refi_(1)),
				binding_(refi_(1), num_))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = "test";
              return a;
            }`),
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				// a: brefi_(2, str_)
				binding_(refi_(1), brefi_(2, str_)))),
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			stmts := parse(t, test.text)
			clauses, err := typing.BuildClauses(stmts)
			if err != nil {
				t.Errorf("got err: %v", err)
				return
			}
			if diff := cmp.Diff(test.want, clauses, ignoreTypeFields); diff != "" {
				t.Errorf("(-want,+got):\n%s", diff)
			}
		})
	}
}
