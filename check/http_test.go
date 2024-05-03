package check_test

import (
	"context"
	"net"
	"testing"

	"github.com/gzuidhof/flipper/check"
	"github.com/gzuidhof/flipper/checker"
	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ checker.Check[check.HTTPCheckResult] = (*check.HTTPCheck)(nil)

func getIPForHost(t *testing.T, host string) string {
	t.Helper()
	ip, err := net.LookupHost(host)
	require.NoError(t, err)

	return ip[0]
}

func TestHTTPCheck(t *testing.T) {
	t.Parallel()

	exampleComIP := getIPForHost(t, "example.com")

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		cfg := cfgmodel.HealthCheckConfig{
			Type: "some-invalid-value",
		}

		h := check.NewHTTPCheck(cfg, "0.0.0.0")
		result := h.Check(context.Background())
		assert.Error(t, result.Error)
	})

	t.Run("https", func(t *testing.T) {
		t.Parallel()
		for _, host := range []string{"example.com", "www.example.com", "www.example.com"} {
			cfg := cfgmodel.HealthCheckConfig{
				Type: "https",
				Host: host,
				Path: "/",
			}

			h := check.NewHTTPCheck(cfg, exampleComIP)
			result := h.Check(context.Background())
			assert.NoError(t, result.Error)
		}
	})
}
