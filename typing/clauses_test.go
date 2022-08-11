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
			clauses_(clause_(func_(types_(), bref_(nil_)))),
		},
		{
			"fun answer() { return 42; }",
			clauses_(clause_(func_(types_(), bref_(num_)))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = "test";
              return a;
            }`),
			clauses_(clause_(
				func_(types_(), bref_(bref_(str_))))),
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
