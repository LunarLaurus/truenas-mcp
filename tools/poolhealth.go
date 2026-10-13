package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerPoolHealthTools registers the compact pool health view.
func registerPoolHealthTools(r *Registry) {
	r.tools["get_pool_health"] = Tool{
		Definition: mcp.Tool{
			Name: "get_pool_health",
			Description: "Compact pool health from pool.query: status, used percent, fragmentation, dedup table size, " +
				"last scrub or resilver, per-vdev error counts, and whether the pool has a dedicated log (SLOG), cache, " +
				"special or dedup vdev. query_pools returns the full raw objects",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pool": map[string]interface{}{"type": "string", "description": "Optional: only this pool"},
				},
			},
		},
		Handler: handleGetPoolHealth,
	}
}

// numberOrNil returns the value when it is a JSON number, otherwise nil.
func numberOrNil(value interface{}) interface{} {
	if number, ok := value.(float64); ok {
		return number
	}
	return nil
}

// vdevCount returns how many entries a topology section holds.
func vdevCount(topology map[string]interface{}, section string) int {
	list, _ := topology[section].([]interface{})
	return len(list)
}

// summarizePoolHealth reduces pool.query entries to the fields that answer "is this pool healthy and
// how is it laid out", sorted by name. Fields the pool does not report are omitted or nil.
func summarizePoolHealth(pools []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(pools))
	for _, pool := range pools {
		name, _ := pool["name"].(string)
		entry := map[string]interface{}{
			"name":              name,
			"status":            pool["status"],
			"healthy":           pool["healthy"],
			"status_code":       pool["status_code"],
			"fragmentation":     pool["fragmentation"],
			"dedup_table_bytes": pool["dedup_table_size"],
			"dedup_table_quota": pool["dedup_table_quota"],
			"used_percent":      nil,
			"scan":              nil,
		}
		if detail, ok := pool["status_detail"].(string); ok && detail != "" {
			entry["status_detail"] = detail
		}
		if autotrim, ok := pool["autotrim"].(map[string]interface{}); ok {
			entry["autotrim"] = autotrim["value"]
		}

		size, _ := pool["size"].(float64)
		allocated, _ := pool["allocated"].(float64)
		if size > 0 {
			entry["used_percent"] = math.Round(allocated/size*1000) / 10
		}

		if scan, ok := pool["scan"].(map[string]interface{}); ok {
			percent, _ := scan["percentage"].(float64)
			scanEntry := map[string]interface{}{
				"function": scan["function"],
				"state":    scan["state"],
				"errors":   scan["errors"],
				"percent":  math.Round(math.Min(percent, 100)*10) / 10,
			}
			if ended, ok := scan["end_time"].(map[string]interface{}); ok {
				scanEntry["ended_ms"] = ended["$date"]
			}
			entry["scan"] = scanEntry
		}

		topology, _ := pool["topology"].(map[string]interface{})
		dataVdevs := []map[string]interface{}{}
		data, _ := topology["data"].([]interface{})
		for _, raw := range data {
			vdev, _ := raw.(map[string]interface{})
			children, _ := vdev["children"].([]interface{})
			disks := len(children)
			if disks == 0 {
				disks = 1
			}
			stats, _ := vdev["stats"].(map[string]interface{})
			dataVdevs = append(dataVdevs, map[string]interface{}{
				"type":            vdev["type"],
				"disks":           disks,
				"read_errors":     numberOrNil(stats["read_errors"]),
				"write_errors":    numberOrNil(stats["write_errors"]),
				"checksum_errors": numberOrNil(stats["checksum_errors"]),
			})
		}
		entry["data_vdevs"] = dataVdevs
		entry["has_log"] = vdevCount(topology, "log") > 0
		entry["has_cache"] = vdevCount(topology, "cache") > 0
		entry["has_special"] = vdevCount(topology, "special") > 0
		entry["has_dedup_vdev"] = vdevCount(topology, "dedup") > 0

		summary = append(summary, entry)
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["name"].(string) < summary[j]["name"].(string)
	})
	return summary
}

func handleGetPoolHealth(client *truenas.Client, args map[string]interface{}) (string, error) {
	pools, err := callClientList(client, "pool.query")
	if err != nil {
		return "", err
	}
	if want, _ := args["pool"].(string); want != "" {
		filtered := make([]map[string]interface{}, 0, 1)
		for _, pool := range pools {
			if pool["name"] == want {
				filtered = append(filtered, pool)
			}
		}
		pools = filtered
	}
	formatted, err := json.MarshalIndent(summarizePoolHealth(pools), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format pool health: %w", err)
	}
	return string(formatted), nil
}
