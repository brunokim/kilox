package lox_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

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
		if _, ok := experiments["typing"]; !ok {
			experiments["typing"] = true // Enable typing, if not specified.
		}
		output, err := runLox(text, experiments)
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
