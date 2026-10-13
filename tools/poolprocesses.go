package tools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// maxCommandLineRunes is the longest command line reported before it is cut with an ellipsis.
const maxCommandLineRunes = 120

// poolNameSchema is the shared 'pool' argument of the tools that take a pool name.
var poolNameSchema = map[string]interface{}{
	"type":        "string",
	"description": "Pool name, for example 'tank'",
}

// registerPoolProcessTools registers the read-only tool listing processes that hold a pool open.
func registerPoolProcessTools(r *Registry) {
	r.tools["get_pool_processes"] = Tool{
		Definition: mcp.Tool{
			Name: "get_pool_processes",
			Description: "Read-only list of running processes using a pool (pool.processes, role POOL_READ), " +
				"for finding what blocks an export or keeps a pool busy. Command lines are cut to 120 characters",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"pool": poolNameSchema},
				"required":   []string{"pool"},
			},
		},
		Handler: handleGetPoolProcesses,
	}
}

// poolIDFromQuery returns the id of the pool called name from pool.query entries. The error lists the
// known pool names so a typo is easy to correct.
func poolIDFromQuery(pools []map[string]interface{}, name string) (int, error) {
	known := make([]string, 0, len(pools))
	for _, pool := range pools {
		poolName, _ := pool["name"].(string)
		known = append(known, poolName)
		if poolName != name {
			continue
		}
		if id, ok := pool["id"].(float64); ok {
			return int(id), nil
		}
	}
	sort.Strings(known)
	return 0, fmt.Errorf("pool %q not found; known pools: %s", name, strings.Join(known, ", "))
}

// resolvePoolID reads the required 'pool' argument and resolves it to a pool id through pool.query.
func resolvePoolID(client *truenas.Client, args map[string]interface{}) (int, error) {
	name, _ := args["pool"].(string)
	if name == "" {
		return 0, fmt.Errorf("the 'pool' argument is required")
	}
	pools, err := callClientList(client, "pool.query")
	if err != nil {
		return 0, err
	}
	return poolIDFromQuery(pools, name)
}

// truncateCommandLine cuts a command line to 120 runes, appending "..." when it was shortened.
func truncateCommandLine(commandLine string) string {
	runes := []rune(commandLine)
	if len(runes) <= maxCommandLineRunes {
		return commandLine
	}
	return string(runes[:maxCommandLineRunes]) + "..."
}

// summarizePoolProcesses reduces pool.processes entries to pid, name, service and a truncated command
// line, sorted by name then pid.
func summarizePoolProcesses(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, process := range raw {
		commandLine, _ := process["cmdline"].(string)
		summary = append(summary, map[string]interface{}{
			"pid":     process["pid"],
			"name":    process["name"],
			"service": process["service"],
			"cmdline": truncateCommandLine(commandLine),
		})
	}
	sort.SliceStable(summary, func(i, j int) bool {
		leftName, _ := summary[i]["name"].(string)
		rightName, _ := summary[j]["name"].(string)
		if leftName != rightName {
			return leftName < rightName
		}
		leftPid, _ := summary[i]["pid"].(float64)
		rightPid, _ := summary[j]["pid"].(float64)
		return leftPid < rightPid
	})
	return summary
}

func handleGetPoolProcesses(client *truenas.Client, args map[string]interface{}) (string, error) {
	poolID, err := resolvePoolID(client, args)
	if err != nil {
		return "", err
	}
	raw, err := client.Call("pool.processes", poolID)
	if err != nil {
		return "", err
	}
	var processes []map[string]interface{}
	if err := json.Unmarshal(raw, &processes); err != nil {
		return "", fmt.Errorf("failed to parse pool.processes: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizePoolProcesses(processes), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format pool processes: %w", err)
	}
	return string(formatted), nil
}
