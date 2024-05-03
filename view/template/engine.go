// Package template provides templates for rendering HTML pages.
package template

import (
	"embed"
	"errors"
	"io"
	"io/fs"
)

//go:embed **/*.tmpl.html
var embedFS embed.FS

// ErrFailedToParse is an error that is returned when the templates fail to parse.
var ErrFailedToParse = errors.New("failed to parse templates")

// Templater is an interface for rendering templates.
type Templater[T any] interface {
	ExecuteTemplate(wr io.Writer, data T) error
}

// Engine is a Templater that uses embedded templates.
type Engine struct {
	fs fs.FS
}

// New creates a new template Engine.
func New(fs fs.FS) *Engine {
	return &Engine{fs}
}

// NewEmbedded creates a new template engine using the embedded templates.
func NewEmbedded() *Engine {
	return New(embedFS)
}

func getPageTemplate[T any](tfs *Engine, pagename string) PageTemplate[T] {
	globs := []string{
		"layout/base.tmpl.html",
		"layout/simple.tmpl.html",
		"component/*.tmpl.html",
		"page/" + pagename + ".tmpl.html",
	}
	return PageTemplate[T]{
		name:  pagename,
		fs:    tfs.fs,
		globs: globs,
	}
}

// HomePageData is the data needed to render the home page.
type HomePageData struct{}

// HomePage returns a handle to the template for the home page.
func (tfs *Engine) HomePage() PageTemplate[HomePageData] {
	return getPageTemplate[HomePageData](tfs, "home")
}
