package template

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"

	"github.com/Masterminds/sprig/v3"
)

// PageTemplate is a template for a specific page.
type PageTemplate[T any] struct {
	name  string
	fs    fs.FS
	globs []string
}

// ExecuteTemplate executes a template with the given data.
func (p PageTemplate[T]) ExecuteTemplate(wr io.Writer, data T) error {
	t, err := template.New(p.name).
		Funcs(sprig.FuncMap()).
		ParseFS(p.fs, p.globs...)
	if err != nil {
		return fmt.Errorf("failed to parse templates for page %s: %w", p.name, err)
	}

	pageTemplate := t.Lookup(p.name + ".tmpl.html")
	if pageTemplate == nil {
		return fmt.Errorf("failed to find template for page %s", p.name)
	}

	err = pageTemplate.Execute(wr, data)
	if err != nil {
		return fmt.Errorf("failed to execute template %s: %w", p.name, err)
	}
	return nil
}
