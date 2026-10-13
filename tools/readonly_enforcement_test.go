package tools

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/truenas/truenas-mcp/mcp"
)

// writeMethodOwners maps every TrueNAS API method that changes state to the tool that calls it. A new
// write method anywhere in the tool sources fails TestEveryWriteMethodHasAMutatingOwner until it is added
// here with an owning tool that is also in mutatingTools.
var writeMethodOwners = map[string]string{
	"alert.dismiss":                   "dismiss_alert",
	"alert.restore":                   "restore_alert",
	"app.create":                      "install_app",
	"app.delete":                      "delete_app",
	"app.start":                       "start_app",
	"app.stop":                        "stop_app",
	"app.update":                      "update_app",
	"app.upgrade":                     "upgrade_app",
	"boot.environment.destroy":        "delete_boot_environment",
	"directoryservices.cache_refresh": "refresh_directory_cache",
	"directoryservices.leave":         "leave_directory_service",
	"directoryservices.update":        "configure_directory_service",
	"pool.dataset.create":             "create_dataset",
	"pool.scrub.create":               "create_scrub_schedule",
	"pool.scrub.delete":               "delete_scrub_schedule",
	"pool.scrub.run":                  "run_scrub",
	"sharing.nfs.create":              "create_nfs_share",
	"sharing.smb.create":              "create_smb_share",
	"system.reboot":                   "system_reboot",
	"update.download":                 "download_update",
	"update.run":                      "apply_update",
}

var apiMethodShape = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// writeVerbPrefixes are the leading words of an API method's last segment that mark it as state-changing.
var writeVerbPrefixes = []string{
	"create", "update", "delete", "destroy", "run", "dismiss", "restore", "start", "stop", "upgrade",
	"reboot", "shutdown", "download", "leave", "cache_refresh", "set_", "sync", "wipe", "promote",
	"rollback", "lock", "unlock", "abort", "install", "import", "rename", "convert", "apply", "clear",
	"reset", "join", "remove", "add", "enable", "disable",
}

// readOnlyLookalikes are methods whose last segment starts with a write verb but only read.
var readOnlyLookalikes = map[string]bool{
	"app.upgrade_summary":       true,
	"update.available_versions": true,
}

func isWriteMethod(method string) bool {
	if readOnlyLookalikes[method] {
		return false
	}
	last := method[strings.LastIndex(method, ".")+1:]
	for _, prefix := range writeVerbPrefixes {
		if strings.HasPrefix(last, prefix) {
			return true
		}
	}
	return false
}

// apiMethodLiterals returns every string literal in the non-test tool sources that has the shape of a
// TrueNAS API method name (dotted, lower case), however it is passed to the client.
func apiMethodLiterals(t *testing.T) map[string]bool {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err == nil && apiMethodShape.MatchString(value) && !strings.HasSuffix(value, ".com") {
				found[value] = true
			}
			return true
		})
	}
	return found
}

func TestEveryWriteMethodHasAMutatingOwner(t *testing.T) {
	found := map[string]bool{}
	for method := range apiMethodLiterals(t) {
		if isWriteMethod(method) {
			found[method] = true
		}
	}

	var unowned, stale []string
	for method := range found {
		if _, ok := writeMethodOwners[method]; !ok {
			unowned = append(unowned, method)
		}
	}
	for method := range writeMethodOwners {
		if !found[method] {
			stale = append(stale, method)
		}
	}
	sort.Strings(unowned)
	sort.Strings(stale)
	if len(unowned) > 0 {
		t.Errorf("write methods with no owner in writeMethodOwners (add the owning tool here and to mutatingTools): %v", unowned)
	}
	if len(stale) > 0 {
		t.Errorf("writeMethodOwners lists methods no longer called anywhere: %v", stale)
	}

	registry := newTestRegistry()
	for method, owner := range writeMethodOwners {
		if _, registered := registry.tools[owner]; !registered {
			t.Errorf("%s: owner %q is not a registered tool", method, owner)
		}
		if !mutatingTools[owner] {
			t.Errorf("%s: owner %q is missing from mutatingTools", method, owner)
		}
	}
}

func TestEveryRegisteredToolIsClassifiedExactlyOnce(t *testing.T) {
	for name := range newTestRegistry().tools {
		inMutating, inReadOnly := mutatingTools[name], readOnlyTools[name]
		if !inMutating && !inReadOnly {
			t.Errorf("tool %q is in neither mutatingTools nor readOnlyTools: classify it", name)
		}
		if inMutating && inReadOnly {
			t.Errorf("tool %q is in both mutatingTools and readOnlyTools", name)
		}
	}
	for name := range readOnlyTools {
		if _, registered := newTestRegistry().tools[name]; !registered {
			t.Errorf("readOnlyTools lists %q, which is not a registered tool", name)
		}
	}
}

func TestReadOnlyModeIsDefaultDenyForUnclassifiedTools(t *testing.T) {
	r := newTestRegistry()
	r.tools["frobnicate_pool"] = Tool{
		Definition: mcp.Tool{Name: "frobnicate_pool"},
		Handler:    nil,
	}
	r.SetReadOnly(true)

	for _, tool := range r.ListTools() {
		if tool.Name == "frobnicate_pool" {
			t.Error("an unclassified tool must be hidden in read-only mode")
		}
	}
	if _, err := r.CallTool("frobnicate_pool", nil); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Errorf("CallTool on an unclassified tool in read-only mode = %v, want a read-only refusal", err)
	}
}

func TestReadOnlyModeExposesExactlyTheAllowlist(t *testing.T) {
	r := newTestRegistry()
	r.SetReadOnly(true)

	listed := map[string]bool{}
	for _, tool := range r.ListTools() {
		listed[tool.Name] = true
	}
	for name := range readOnlyTools {
		if !listed[name] {
			t.Errorf("read-only tool %q is not exposed in read-only mode", name)
		}
	}
	if len(listed) != len(readOnlyTools) {
		t.Errorf("read-only mode lists %d tools, allowlist has %d", len(listed), len(readOnlyTools))
	}
}
