package notificationtemplate

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/plan"
	"github.com/gzuidhof/flipper/resource"
)

//go:embed *.tmpl
var templateFS embed.FS

//nolint:gochecknoglobals // Static templates.
var templates = template.Must(template.ParseFS(templateFS, "*.tmpl"))

type planTemplateData struct {
	Cfg                             cfgmodel.GroupConfig
	State                           plan.State
	Plan                            plan.Plan
	FloatingIPsByServer             map[string]resource.FloatingIPs
	ToBeReassigned                  map[string]bool
	UnassignedFloatingIPs           resource.FloatingIPs
	FloatingIPsTargetedOutsideGroup resource.FloatingIPs
}

type stateTemplateData struct {
	Cfg                             cfgmodel.GroupConfig
	State                           plan.State
	FloatingIPsByServer             map[string]resource.FloatingIPs
	ToBeReassigned                  map[string]bool
	UnassignedFloatingIPs           resource.FloatingIPs
	FloatingIPsTargetedOutsideGroup resource.FloatingIPs
}

// RenderPlanExecution renders a plan notification in markdown format.
func RenderPlanExecution(cfg cfgmodel.GroupConfig, state plan.State, plan plan.Plan) string {
	data := planTemplateData{
		Cfg:                 cfg,
		State:               state,
		Plan:                plan,
		FloatingIPsByServer: state.FloatingIPsByServer(),
		ToBeReassigned:      plan.ToBeReassignedMap(),

		UnassignedFloatingIPs:           state.UnassignedFloatingIPs(),
		FloatingIPsTargetedOutsideGroup: state.FloatingIPsTargetedOutsideGroup(),
	}

	buf := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(buf, "plan.go.tmpl", data)
	if err != nil {
		// Should never happen as the template is hardcoded.
		slog.Error("failed to execute plan template", slog.String("error", err.Error()))
		return fmt.Sprintf("failed to execute plan template: %s", err.Error())
	}
	return buf.String()
}

// RenderState renders a state notification in markdown format.
func RenderState(cfg cfgmodel.GroupConfig, state plan.State) string {
	data := stateTemplateData{
		Cfg:                 cfg,
		State:               state,
		FloatingIPsByServer: state.FloatingIPsByServer(),
		ToBeReassigned:      nil,

		UnassignedFloatingIPs:           state.UnassignedFloatingIPs(),
		FloatingIPsTargetedOutsideGroup: state.FloatingIPsTargetedOutsideGroup(),
	}

	buf := bytes.NewBuffer(nil)
	err := templates.ExecuteTemplate(buf, "state.go.tmpl", data)
	if err != nil {
		// Should never happen as the template is hardcoded.
		slog.Error("failed to execute state template", slog.String("error", err.Error()))
		return fmt.Sprintf("failed to execute state template: %s", err.Error())
	}
	return buf.String()
}
