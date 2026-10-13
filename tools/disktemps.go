package tools

import (
	"encoding/json"
	"fmt"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// defaultTemperatureDays is the documented default window of disk.temperature_agg.
const defaultTemperatureDays = 7

// registerDiskTemperatureTools registers the read-only disk temperature tool.
func registerDiskTemperatureTools(r *Registry) {
	r.tools["get_disk_temperatures"] = Tool{
		Definition: mcp.Tool{
			Name: "get_disk_temperatures",
			Description: "Read-only disk temperatures (disk.temperatures, role REPORTING_READ). " +
				"Reads each disk directly, so it also covers disks missing from the reporting graphs. " +
				"Set 'days' to add min/max/avg over that window (disk.temperature_agg).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"disks": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional: disk names such as 'sda' or 'nvme0n1' (default: every disk)",
					},
					"include_thresholds": map[string]interface{}{
						"type":        "boolean",
						"description": "Optional: include the temperature thresholds the disk reports (default: false)",
					},
					"days": map[string]interface{}{
						"type":        "integer",
						"description": "Optional: also return min/max/avg over this many days",
					},
				},
			},
		},
		Handler: handleGetDiskTemperatures,
	}
}

// diskNamesFromQuery returns the non-empty "name" of every disk.query entry.
func diskNamesFromQuery(disks []map[string]interface{}) []string {
	names := make([]string, 0, len(disks))
	for _, disk := range disks {
		if name, ok := disk["name"].(string); ok && name != "" {
			names = append(names, name)
		}
	}
	return names
}

// diskTemperatureParams builds the disk.temperatures call parameters (name, include_thresholds).
func diskTemperatureParams(names []string, includeThresholds bool) []interface{} {
	return []interface{}{names, includeThresholds}
}

// diskTemperatureAggParams builds the disk.temperature_agg call parameters (names, days).
// A days value below 1 selects the documented default of 7.
func diskTemperatureAggParams(names []string, days int) []interface{} {
	if days < 1 {
		days = defaultTemperatureDays
	}
	return []interface{}{names, days}
}

func handleGetDiskTemperatures(client *truenas.Client, args map[string]interface{}) (string, error) {
	var names []string
	if requested, ok := args["disks"].([]interface{}); ok {
		for _, entry := range requested {
			if name, ok := entry.(string); ok && name != "" {
				names = append(names, name)
			}
		}
	}
	if len(names) == 0 {
		result, err := client.Call("disk.query", []interface{}{}, map[string]interface{}{})
		if err != nil {
			return "", err
		}
		var disks []map[string]interface{}
		if err := json.Unmarshal(result, &disks); err != nil {
			return "", fmt.Errorf("failed to parse disks: %w", err)
		}
		names = diskNamesFromQuery(disks)
	}

	includeThresholds, _ := args["include_thresholds"].(bool)
	current, err := client.Call("disk.temperatures", diskTemperatureParams(names, includeThresholds)...)
	if err != nil {
		return "", err
	}

	response := map[string]interface{}{"current": json.RawMessage(current)}
	if days, ok := args["days"].(float64); ok && days > 0 {
		aggregate, err := client.Call("disk.temperature_agg", diskTemperatureAggParams(names, int(days))...)
		if err != nil {
			response["aggregate_error"] = err.Error()
		} else {
			response["aggregate"] = json.RawMessage(aggregate)
		}
	}

	formatted, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}
