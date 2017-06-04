package httpdoc

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mercari/go-httpdoc/static"
)

// Generate writes documentation into the given file. Generation is skipped
// if EnvHTTPDoc is empty. If directory does not exist or any, it returns error.
func (d *Document) Generate(path string) error {

	// Only generate documentation when EnvHttpDoc has non-empty value
	if os.Getenv(EnvHTTPDoc) == "" {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return d.generate(f)
}

func (d *Document) generate(w io.Writer) error {
	buf, err := static.Asset("tmpl/doc.md.tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New("httpdoc").Funcs(funcMap()).Parse(string(buf))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, d); err != nil {
		return err
	}
	return nil
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"stripslash": func(s string) string {
			return strings.Replace(s, "/", "", -1)
		},
	}
}
