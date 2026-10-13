package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerDiskTemperatureAlertTools registers the read-only disk temperature alert tool.
func registerDiskTemperatureAlertTools(r *Registry) {
	r.tools["get_disk_temperature_alerts"] = Tool{
		Definition: mcp.Tool{
			Name: "get_disk_temperature_alerts",
			Description: "Read-only temperature alerts for disks (disk.temperature_alerts, role REPORTING_READ): level, " +
				"alert class, text, dismissed state and last occurrence. An empty list means no disk is over its limit",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"disks": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional: disk names such as 'sda' or 'nvme0n1' (default: every disk)",
					},
				},
			},
		},
		Handler: handleGetDiskTemperatureAlerts,
	}
}

// summarizeTemperatureAlerts reduces alert objects to level, klass, text, dismissed and last_occurrence,
// sorted by level then text. Internal identifiers, mail state and formatted HTML are omitted.
func summarizeTemperatureAlerts(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, alert := range raw {
		klass, _ := alert["klass"].(string)
		text, _ := alert["text"].(string)
		dismissed, _ := alert["dismissed"].(bool)
		summary = append(summary, map[string]interface{}{
			"level":           alert["level"],
			"klass":           klass,
			"text":            text,
			"dismissed":       dismissed,
			"last_occurrence": alert["last_occurrence"],
		})
	}
	sort.SliceStable(summary, func(i, j int) bool {
		leftLevel, _ := summary[i]["level"].(string)
		rightLevel, _ := summary[j]["level"].(string)
		if leftLevel != rightLevel {
			return leftLevel < rightLevel
		}
		return summary[i]["text"].(string) < summary[j]["text"].(string)
	})
	return summary
}

func handleGetDiskTemperatureAlerts(client *truenas.Client, args map[string]interface{}) (string, error) {
	names := stringList(args["disks"])
	if len(names) == 0 {
		disks, err := callClientList(client, "disk.query")
		if err != nil {
			return "", err
		}
		names = diskNamesFromQuery(disks)
	}
	raw, err := client.Call("disk.temperature_alerts", names)
	if err != nil {
		return "", err
	}
	var alerts []map[string]interface{}
	if err := json.Unmarshal(raw, &alerts); err != nil {
		return "", fmt.Errorf("failed to parse disk.temperature_alerts: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeTemperatureAlerts(alerts), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format temperature alerts: %w", err)
	}
	return string(formatted), nil
}
