package notificationtemplate

import (
	"testing"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/plan"
	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	// Just a smoke test that will catch any template errors.
	out := RenderPlanExecution(cfgmodel.GroupConfig{}, plan.State{}, plan.Plan{})
	assert.NotContains(t, out, "fail")

	out = RenderState(cfgmodel.GroupConfig{}, plan.State{})
	assert.NotContains(t, out, "fail")
}
