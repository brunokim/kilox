package lox_test

import (
	"strings"
	"testing"

	"github.com/brunokim/lox"
	"github.com/google/go-cmp/cmp"
)

func TestInterpreter(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"var a = 10; print a;", "10\n"},
		{"var a = 10; print a; print a;", "10\n10\n"},
		{"var a = 10; var a = 20; print a;", "20\n"},
		{"var a = 10; var b = a + 5; print b;", "15\n"},
	}

	for _, test := range tests {
		var b strings.Builder
		i := lox.NewInterpreter()
		i.SetStdout(&b)

		stmts := parseStmts(t, test.text)
		err := i.Interpret(stmts)
		if err != nil {
			t.Fatalf("%q: want nil, got err: %v", test.text, err)
		}
		if diff := cmp.Diff(test.want, b.String()); diff != "" {
			t.Errorf("%q: (-want, +got)%s", test.text, diff)
		}
	}
}
