package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("VAULT_TYPE", "vault")
	os.Setenv("AUTH_TYPE", "github")
	os.Setenv("GITHUB_TOKEN", "token")
	exitVal := m.Run()
	os.Unsetenv("VAULT_TYPE")
	os.Unsetenv("AUTH_TYPE")
	os.Unsetenv("GITHUB_TOKEN")
	os.Exit(exitVal)
}

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

// func Test_generate_empty(t *testing.T) {
// 	args := []string{"./fixtures/empty/"}
// 	cmd := NewGenerateCommand()
//
// 	b := bytes.NewBufferString("")
// 	cmd.SetArgs(args)
// 	cmd.SetOut(b)
// 	cmd.Execute()
// 	out, err := ioutil.ReadAll(b) // Read buffer to bytes
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	expected := "no YAML files were found in ./fixtures/empty/"
// 	// fmt.Print(strings.Contains(string(out), expected))
// 	if !strings.Contains(string(out), expected) {
// 		t.Fatalf("expected to contain: %s but got %s", expected, out)
// 	}
// }
