package httpdoc

import (
	"io/ioutil"
	"os"
	"testing"
)

func setEnv(t *testing.T, k, v string) func() {
	preV := os.Getenv(k)
	if err := os.Setenv(k, v); err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := os.Setenv(k, preV); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDocument_Generate(t *testing.T) {
	resetF := setEnv(t, EnvHTTPDoc, "test")
	defer resetF()

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	doc := &Document{}
	if err := doc.Generate(f.Name()); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if fi.Size() == 0 {
		t.Fatalf("expect doc to be generated")
	}
}

func TestDocument_Generate_noEnv(t *testing.T) {
	resetF := setEnv(t, EnvHTTPDoc, "")
	defer resetF()

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	doc := &Document{}
	if err := doc.Generate(f.Name()); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if fi.Size() > 0 {
		t.Fatalf("expect doc not to be generated")
	}
}

func TestFuncMap(t *testing.T) {
	m := funcMap()
	lower := m["lower"].(func(s string) string)
	if got, want := lower("DOC"), "doc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	stripslash := m["stripslash"].(func(s string) string)
	if got, want := stripslash("/v2/user/contact"), "v2usercontact"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
