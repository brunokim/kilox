package lox_test

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/brunokim/lox"
	"github.com/google/go-cmp/cmp"
)

func TestInterpreter(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*/*.lox")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		t.Log(filename)
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		text := string(bs)
		output, err := runLox(t, text)
		if err != nil {
			t.Errorf("%s: %v", filename, err)
			continue
		}
		want := extractOutput(text)
		if diff := cmp.Diff(want, output); diff != "" {
			t.Errorf("%s: (-want, +got)%s", filename, diff)
		}
	}
}

func runLox(t *testing.T, text string) (string, error) {
	var b strings.Builder
	i := lox.NewInterpreter()
	i.SetStdout(&b)

	stmts := parseStmts(t, text)
	err := i.Interpret(stmts)
	return b.String(), err
}

func extractOutput(text string) string {
	outputRE := regexp.MustCompile("(?im)// Output:(.*)$")

	var b strings.Builder
	matches := outputRE.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		b.WriteString(strings.TrimPrefix(match[1], " "))
		b.WriteRune('\n')
	}
	return b.String()
}
