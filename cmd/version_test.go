package cmd

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/version"
)

func TestVersion(t *testing.T) {
	t.Run("will show version", func(t *testing.T) {
		args := []string{}
		cmd := NewVersionCommand()

		version.Version = "v0.0.1"
		version.BuildDate = "1970-01-01T:00:00:00Z"
		version.CommitSHA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

		c := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetOut(c)
		cmd.Execute()
		out, err := ioutil.ReadAll(c) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := "argocd-vault-plugin v0.0.1 (AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA) BuildDate: 1970-01-01T:00:00:00Z"
		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}
	})
}
