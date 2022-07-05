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
		wantOutput, wantErr := extractExpected(text)
		experiments := extractExperiments(text)
		output, err := runLox(text, experiments...)
		errMsg := ""
		if err != nil {
			errMsg = err.Error() + "\n"
		}
		if diff := cmp.Diff(wantErr, errMsg); diff != "" {
			t.Errorf("%s: errors: (-want, +got)%s", filename, diff)
		}
		if diff := cmp.Diff(wantOutput, output); diff != "" {
			t.Errorf("%s: (-want, +got)%s", filename, diff)
		}
	}
}

func runLox(text string, experiments ...string) (string, error) {
	enabled := make(map[string]bool)
	for _, exp := range experiments {
		enabled[exp] = true
	}
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		return "", err
	}
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		return "", err
	}
	i := lox.NewInterpreter()
	r := lox.NewResolver(i)
	err = r.Resolve(stmts)
	if err != nil {
		return "", err
	}
	if enabled["typing"] {
		c := lox.NewTypeChecker(i)
		err := c.Check(stmts)
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	i.SetStdout(&b)
	err = i.Interpret(stmts)
	return b.String(), err
}

func extractExpected(text string) (string, string) {
	wantOutput := extractComment(text, "output")
	wantError := extractComment(text, "error")
	return wantOutput, wantError
}

func extractComment(text, pattern string) string {
	commentRE := regexp.MustCompile("(?im)// " + pattern + ":(.*)$")

	var b strings.Builder
	matches := commentRE.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		b.WriteString(strings.TrimPrefix(match[1], " "))
		b.WriteRune('\n')
	}
	return b.String()
}

func extractExperiments(text string) []string {
	expStr := extractComment(text, "experiments")
	exps := strings.Split(expStr, ",")
	for i, exp := range exps {
		exps[i] = strings.TrimSpace(exp)
	}
	return exps
}
