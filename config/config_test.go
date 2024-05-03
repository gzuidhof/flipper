package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	t.Parallel()

	c, err := Init("")
	assert.NoError(t, err)

	// Check that the default values are set
	assert.Equal(t, c.Version, 1)
	assert.Equal(t, c.Server.ShutdownTimeout, time.Second*5)

	// Check that this returns the defaults.
	assert.Equal(t, c, Default())
}

func TestInitConfigWithFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("valid", func(t *testing.T) {
		file := dir + "/config.yaml"

		yamlContent := []byte(`
server:
  shutdown_timeout: 10m
`)
		err := os.WriteFile(file, yamlContent, 0o600)
		require.NoError(t, err)

		c, err := Init(file)
		assert.NoError(t, err)

		// Check that the default values are set
		assert.Equal(t, c.Server.ShutdownTimeout, time.Minute*10)
	})

	t.Run("invalid", func(t *testing.T) {
		file := dir + "/config.yaml"

		yamlContent := []byte(`
server:
  shutdown_timeout: invalid
`)
		err := os.WriteFile(file, yamlContent, 0o600)
		require.NoError(t, err)

		_, err = Init(file)
		assert.Error(t, err)
	})
}
