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

func TestInterpreterSuite(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*/*.lox")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
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

func TestInterpreter(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"var a = 10; print a; print a;", "10\n10\n"},
		{"var a = 10; var a = 20; print a;", "20\n"},
		{"var a = 10; var b = a + 5; print b;", "15\n"},
		{"var a; var b; a = b = 10; print a + b;", "20\n"},
		{"var a; if (a) a = 10; else a = 20; print a;", "20\n"},
		{`var a; var b;
          if (!a)
            if (b)
              b = 20;
            else
              b = 40;
          else
            a = 2;
          if (!a)
            a = 4;
          print a + b;`,
			"44\n"},
		{`print 10 and 0 and nil and "a";`, "nil\n"},
		{`print nil or false or "b" or false;`, "b\n"},
		{`print nil or true and "b" or false;`, "b\n"},
		{`var a = 10; {var a = 20; print a;} print a;`, "20\n10\n"},
		{`var a; var b;
          if (!a) {
            if (b) {
              a = 20;
            } else {
              a = 10;
            }
          }
          print a;`,
			"10\n"},
		{"var a = 1; { var a = a + 2; print a; }", "3\n"},
		{"var a = 5; while (a >= 0) { print a; a = a - 1; }", "5\n4\n3\n2\n1\n0\n"},
		{"for (var i = 0; i < 5; i = i + 1) print i;", "0\n1\n2\n3\n4\n"},
		{`var a = 0;
          var temp;

          for (var b = 1; a < 1000; b = temp + b) {
              print a;
              temp = a;
              a = b;
          }`, `0
1
1
2
3
5
8
13
21
34
55
89
144
233
377
610
987
`},
		{`
        var x = 0;
        while (x < 10) {
            var start = x;
            while (true) {
                x = x + 1;
                print x;
                if (x - start > 3) {
                    break;
                }
            }
            if (x > 6) {
                continue;
            }
            x = x - 3;
        }`, `1
2
3
4
2
3
4
5
3
4
5
6
4
5
6
7
8
9
10
11
`},
		{`for (var i=0; i < 10; i = i+1) {
		    if (i < 4) {
		        continue;
		    }
		    print i;
		    if (i > 6) {
		        break;
		    }
		}`, "4\n5\n6\n7\n"},
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
