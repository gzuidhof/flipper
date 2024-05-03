package template

import (
	"fmt"
	"net/http"
)

// RenderTemplateHTML sets the Content-Type header to text/html and executes the given template.
func RenderTemplateHTML[T any](template Templater[T], w http.ResponseWriter, data T) error {
	w.Header().Set("Content-Type", "text/html")

	err := template.ExecuteTemplate(w, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	return nil
}
