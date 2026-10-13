package tools

import (
	"strings"
	"testing"

	"github.com/truenas/truenas-mcp/mcp"
)

func newTestRegistry() *Registry {
	return NewRegistry(nil, nil)
}

func listedNames(tools []mcp.Tool) map[string]bool {
	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}
	return names
}

func TestEveryRegisteredToolIsClassifiedAsMutatingOrNot(t *testing.T) {
	for name := range mutatingTools {
		if _, registered := newTestRegistry().tools[name]; !registered {
			t.Errorf("mutatingTools lists %q, which is not a registered tool", name)
		}
	}
}

func TestMutatingVerbsAreAllClassified(t *testing.T) {
	verbs := []string{
		"create_", "delete_", "update_", "install_", "upgrade_", "start_", "stop_", "run_",
		"apply_", "download_", "configure_", "leave_", "dismiss_", "restore_", "refresh_", "system_reboot",
	}
	readTools := map[string]bool{"update_status": true}
	for name := range newTestRegistry().tools {
		for _, verb := range verbs {
			if strings.HasPrefix(name, verb) && !readTools[name] && !mutatingTools[name] {
				t.Errorf("tool %q looks mutating but is missing from mutatingTools", name)
			}
		}
	}
}

func TestReadOnlyModeHidesMutatingTools(t *testing.T) {
	r := newTestRegistry()
	r.SetReadOnly(true)

	names := listedNames(r.ListTools())
	for name := range mutatingTools {
		if names[name] {
			t.Errorf("read-only mode still lists mutating tool %q", name)
		}
	}
	for _, name := range []string{"system_info", "query_pools", "query_datasets", "get_arc_metrics", "query_disks"} {
		if !names[name] {
			t.Errorf("read-only mode must keep read tool %q", name)
		}
	}
}

func TestReadOnlyModeRefusesMutatingCalls(t *testing.T) {
	r := newTestRegistry()
	r.SetReadOnly(true)

	for name := range mutatingTools {
		_, err := r.CallTool(name, map[string]interface{}{})
		if err == nil || !strings.Contains(err.Error(), "read-only") {
			t.Errorf("CallTool(%q) in read-only mode = %v, want a read-only refusal", name, err)
		}
	}
}

func TestDefaultModeListsMutatingTools(t *testing.T) {
	names := listedNames(newTestRegistry().ListTools())
	for name := range mutatingTools {
		if !names[name] {
			t.Errorf("default mode must list %q", name)
		}
	}
}
