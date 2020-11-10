package cmd

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_generate_noargs(t *testing.T) {
	args := []string{}
	cmd := NewGenerateCommand()

	b := bytes.NewBufferString("")
	cmd.SetArgs(args)
	cmd.SetOut(b)
	cmd.Execute()
	out, err := ioutil.ReadAll(b) // Read buffer to bytes
	if err != nil {
		t.Fatal(err)
	}

	expected := "<path> argument required to generate manifests"

	if !strings.Contains(string(out), expected) {
		t.Fatalf("expected to contain: %s but got %s", expected, out)
	}
}

func Test_generate_empty(t *testing.T) {
	args := []string{"./fixtures/empty/"}
	cmd := NewGenerateCommand()

	b := bytes.NewBufferString("")
	cmd.SetArgs(args)
	cmd.SetOut(b)
	cmd.Execute()
	out, err := ioutil.ReadAll(b) // Read buffer to bytes
	if err != nil {
		t.Fatal(err)
	}

	expected := "No YAML files were found in ./fixtures/empty/"

	if !strings.Contains(string(out), expected) {
		t.Fatalf("expected to contain: %s but got %s", expected, out)
	}
}
