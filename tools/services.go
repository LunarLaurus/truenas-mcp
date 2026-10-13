package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerServiceTools registers the read-only service status view.
func registerServiceTools(r *Registry) {
	r.tools["query_services"] = Tool{
		Definition: mcp.Tool{
			Name:        "query_services",
			Description: "Read-only service status from service.query (role SERVICE_READ): state, start-on-boot flag and process count per service",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"running_only": map[string]interface{}{
						"type":        "boolean",
						"description": "Optional: only services in the RUNNING state (default: false)",
					},
				},
			},
		},
		Handler: handleQueryServices,
	}
}

// summarizeServices reduces service.query entries to name, state, enable and pid count, sorted by name.
// The API documents the name as either "service" or "name", so both are accepted.
func summarizeServices(raw []map[string]interface{}, runningOnly bool) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, entry := range raw {
		name, _ := entry["service"].(string)
		if name == "" {
			name, _ = entry["name"].(string)
		}
		if name == "" {
			continue
		}
		state, _ := entry["state"].(string)
		if runningOnly && state != "RUNNING" {
			continue
		}
		pids, _ := entry["pids"].([]interface{})
		summary = append(summary, map[string]interface{}{
			"service":   name,
			"state":     state,
			"enable":    entry["enable"],
			"pid_count": len(pids),
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["service"].(string) < summary[j]["service"].(string)
	})
	return summary
}

func handleQueryServices(client *truenas.Client, args map[string]interface{}) (string, error) {
	services, err := callClientList(client, "service.query")
	if err != nil {
		return "", err
	}
	runningOnly, _ := args["running_only"].(bool)

	formatted, err := json.MarshalIndent(summarizeServices(services, runningOnly), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format services: %w", err)
	}
	return string(formatted), nil
}
