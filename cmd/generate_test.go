package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func Test_generate(t *testing.T) {
	cmd := NewGenerateCommand() // Initialize generate command

	b := bytes.NewBufferString("")
	cmd.SetOut(b) // Set command output to buffer

	cmd.Execute() // Execute command

	out, err := ioutil.ReadAll(b) // Read buffer to bytes
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != "Generate YAML" { // Compare output with expected
		t.Fatalf("expected \"%s\" got \"%s\"", "Generate YAML", string(out))
	}
}
