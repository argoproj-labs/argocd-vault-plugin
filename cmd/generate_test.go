package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/vault"
)

var roleid, secretid string
var cluster *vault.TestCluster
var client *api.Client

func TestMain(t *testing.T) {
	cluster, roleid, secretid = helpers.CreateTestAppRoleVault(t)
	os.Setenv("AVP_TYPE", "vault")
	os.Setenv("VAULT_ADDR", cluster.Cores[0].Client.Address())
	os.Setenv("AVP_AUTH_TYPE", "approle")
	os.Setenv("AVP_SECRET_ID", secretid)
	os.Setenv("AVP_ROLE_ID", roleid)
	os.Setenv("VAULT_SKIP_VERIFY", "true")

	t.Run("will throw an error expecting arguments", func(t *testing.T) {
		args := []string{}
		cmd := NewGenerateCommand()

		c := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetErr(c)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.Execute()
		out, err := ioutil.ReadAll(c) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := "<path> argument required to generate manifests"
		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}
	})

	t.Run("will return that couldn't find yamls", func(t *testing.T) {
		args := []string{"./fixtures/input/empty/"}
		cmd := NewGenerateCommand()

		b := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetErr(b)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.Execute()
		out, err := ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := "no YAML or JSON files were found in ./fixtures/input/empty/"
		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}
	})

	t.Run("returns error for empty manifests", func(t *testing.T) {
		// From path
		args := []string{"../fixtures/input/empty/file.yaml"}
		cmd := NewGenerateCommand()

		b := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetErr(b)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.Execute()
		out, err := ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := ""
		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}

		// From stdin
		args = []string{"-"}
		stdin := bytes.NewBufferString("")
		inputBuf, err := ioutil.ReadFile("../fixtures/input/empty/file.yaml")
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(inputBuf)

		b = bytes.NewBufferString("")
		cmd.SetIn(stdin)
		cmd.SetArgs(args)
		cmd.SetErr(b)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.Execute()
		out, err = ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}
	})

	t.Run("will replace templates from local vault", func(t *testing.T) {
		args := []string{"../fixtures/input/nonempty"}
		cmd := NewGenerateCommand()

		b := bytes.NewBufferString("")
		e := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetOut(b)
		cmd.SetErr(e)
		cmd.Execute()
		out, err := ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}
		stderr, err := ioutil.ReadAll(e) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		buf, err := ioutil.ReadFile("../fixtures/output/all.yaml")
		if err != nil {
			t.Fatal(err)
		}

		expected := string(buf)
		if string(out) != expected {
			t.Fatalf("expected %s\n\nbut got\n\n%s\nerr: %s", expected, string(out), string(stderr))
		}
	})

	t.Run("will ignore templates with avp.kubernetes.io/ignore set to True", func(t *testing.T) {
		args := []string{"../fixtures/input/nonempty/ignored-secret.yaml"}
		cmd := NewGenerateCommand()

		b := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetOut(b)
		cmd.Execute()
		out, err := ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		buf, err := ioutil.ReadFile("../fixtures/output/ignored-secret.yaml")
		if err != nil {
			t.Fatal(err)
		}

		expected := string(buf)
		if string(out) != expected {
			t.Fatalf("expected %s but got %s", expected, string(out))
		}
	})

	t.Run("will read from STDIN", func(t *testing.T) {
		stdin := bytes.NewBufferString("")
		inputBuf, err := ioutil.ReadFile("../fixtures/input/nonempty/full.yaml")
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(inputBuf)

		args := []string{"-"}
		cmd := NewGenerateCommand()

		stdout := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetOut(stdout)
		cmd.SetIn(stdin)
		cmd.Execute()
		out, err := ioutil.ReadAll(stdout) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		buf, err := ioutil.ReadFile("../fixtures/output/stdin-full.yaml")
		if err != nil {
			t.Fatal(err)
		}

		expected := string(buf)
		if string(out) != expected {
			t.Fatalf("expected %s but got %s", expected, string(out))
		}
	})

	t.Run("will return invalid yaml error from STDIN", func(t *testing.T) {
		stdin := bytes.NewBufferString("")
		inputBuf, err := ioutil.ReadFile("../fixtures/input/invalid.yaml")
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(inputBuf)

		args := []string{"-"}
		cmd := NewGenerateCommand()

		stderr := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetErr(stderr)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.SetIn(stdin)
		cmd.Execute()
		out, err := ioutil.ReadAll(stderr) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := "Error: error converting YAML to JSON: yaml: line 18: did not find expected key"
		if strings.TrimSpace(string(out)) != expected {
			t.Fatalf("expected %s but got %s", expected, string(out))
		}
	})

	t.Run("will return that path validation env is not valid", func(t *testing.T) {
		args := []string{"../fixtures/input/nonempty"}
		cmd := NewGenerateCommand()

		// set specific env and register cleanup func
		os.Setenv("AVP_PATH_VALIDATION", `^\/(?!\/)(.*?)`)
		t.Cleanup(func() {
			os.Unsetenv("AVP_PATH_VALIDATION")
		})

		b := bytes.NewBufferString("")
		cmd.SetArgs(args)
		cmd.SetErr(b)
		cmd.SetOut(bytes.NewBufferString(""))
		cmd.Execute()
		out, err := ioutil.ReadAll(b) // Read buffer to bytes
		if err != nil {
			t.Fatal(err)
		}

		expected := "^\\/(?!\\/)(.*?) is not a valid regular expression: error parsing regexp: invalid or unsupported Perl syntax: `(?!`"
		if !strings.Contains(string(out), expected) {
			t.Fatalf("expected to contain: %s but got %s", expected, out)
		}
	})

	os.Unsetenv("AVP_TYPE")
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("AVP_AUTH_TYPE")
	os.Unsetenv("AVP_SECRET_ID")
	os.Unsetenv("AVP_ROLE_ID")
	os.Unsetenv("VAULT_SKIP_VERIFY")
	os.Unsetenv("AVP_PATH_VALIDATION")
}
