package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerDiskDetailTools registers the read-only disk usage view.
func registerDiskDetailTools(r *Registry) {
	r.tools["get_disk_usage"] = Tool{
		Definition: mcp.Tool{
			Name: "get_disk_usage",
			Description: "Read-only disk inventory from disk.details: name, model, size, type, bus, RPM, enclosure slot and " +
				"the imported pool it belongs to (or an exported pool it was last part of), so unused disks are easy to spot. Serial numbers, passwords and " +
				"key identifiers are not returned",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"USED", "UNUSED", "BOTH"},
						"description": "Optional: which disks to list (default: BOTH)",
					},
				},
			},
		},
		Handler: handleGetDiskUsage,
	}
}

// trimmedString returns the string with surrounding spaces removed (Sun OEM drives pad identifiers), or the
// value unchanged when it is not a string.
func trimmedString(value interface{}) interface{} {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return value
}

// diskDetailEntries flattens the disk.details result into a list of disk objects. The method may return a
// flat list or an object with "used" and "unused" lists.
func diskDetailEntries(raw interface{}) []map[string]interface{} {
	var entries []map[string]interface{}
	collect := func(list []interface{}) {
		for _, item := range list {
			if disk, ok := item.(map[string]interface{}); ok {
				entries = append(entries, disk)
			}
		}
	}
	switch typed := raw.(type) {
	case []interface{}:
		collect(typed)
	case map[string]interface{}:
		for _, key := range []string{"used", "unused"} {
			if list, ok := typed[key].([]interface{}); ok {
				collect(list)
			}
		}
	}
	return entries
}

// summarizeDiskDetails reduces disk.details entries to an allowlist of fields, sorted by name.
// in_use is true when the disk reports a pool. Serials, passwords and KMIP identifiers are omitted.
func summarizeDiskDetails(raw interface{}) []map[string]interface{} {
	entries := diskDetailEntries(raw)
	summary := make([]map[string]interface{}, 0, len(entries))
	for _, disk := range entries {
		name, _ := disk["devname"].(string)
		if name == "" {
			name, _ = disk["name"].(string)
		}
		imported, _ := disk["imported_zpool"].(string)
		entry := map[string]interface{}{
			"name":          name,
			"model":         trimmedString(disk["model"]),
			"type":          disk["type"],
			"bus":           disk["bus"],
			"rpm":           numberOrNil(disk["rotationrate"]),
			"pool":          disk["imported_zpool"],
			"exported_pool": disk["exported_zpool"],
			"in_use":        imported != "",
			"slot":          nil,
		}
		if size, ok := disk["size"].(float64); ok {
			entry["size_gb"] = math.Round(size/1024/1024/1024*10) / 10
		}
		if enclosure, ok := disk["enclosure"].(map[string]interface{}); ok {
			number, numberOK := enclosure["number"].(float64)
			slot, slotOK := enclosure["slot"].(float64)
			if numberOK && slotOK {
				entry["slot"] = fmt.Sprintf("%d:%d", int(number), int(slot))
			}
		}
		summary = append(summary, entry)
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["name"].(string) < summary[j]["name"].(string)
	})
	return summary
}

func handleGetDiskUsage(client *truenas.Client, args map[string]interface{}) (string, error) {
	diskType := "BOTH"
	if requested, ok := args["type"].(string); ok && (requested == "USED" || requested == "UNUSED" || requested == "BOTH") {
		diskType = requested
	}
	result, err := client.Call("disk.details", map[string]interface{}{"type": diskType})
	if err != nil {
		return "", err
	}
	var raw interface{}
	if err := json.Unmarshal(result, &raw); err != nil {
		return "", fmt.Errorf("failed to parse disk.details: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeDiskDetails(raw), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format disk usage: %w", err)
	}
	return string(formatted), nil
}
