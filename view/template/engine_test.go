package template

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddedFSContainsTemplates(t *testing.T) {
	files := make([]string, 0)

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	})

	assert.NoError(t, err)
	assert.Greater(t, len(files), 0)
}

func TestEmbedTemplates(t *testing.T) {
	embedT := NewEmbedded()

	homePageTemplate := embedT.HomePage()

	sink := bytes.NewBuffer([]byte{})

	// Test that the template can be executed without error
	err := homePageTemplate.ExecuteTemplate(sink, HomePageData{})
	assert.NoError(t, err)
}
