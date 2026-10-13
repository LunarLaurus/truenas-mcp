package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerSnapshotTaskTools registers the read-only periodic snapshot task view.
func registerSnapshotTaskTools(r *Registry) {
	r.tools["query_snapshot_tasks"] = Tool{
		Definition: mcp.Tool{
			Name: "query_snapshot_tasks",
			Description: "Read-only periodic snapshot tasks from pool.snapshottask.query: dataset, recursion, " +
				"cron-style schedule, retention and enabled state. Useful for lining periodic I/O bursts up with snapshot schedules",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQuerySnapshotTasks,
	}
}

// summarizeSnapshotTasks reduces pool.snapshottask.query entries to dataset, recursion, retention,
// a "minute hour dom month dow" schedule string and an exclude count, sorted by dataset.
func summarizeSnapshotTasks(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, task := range raw {
		dataset, _ := task["dataset"].(string)
		if dataset == "" {
			continue
		}
		recursive, _ := task["recursive"].(bool)
		enabled, _ := task["enabled"].(bool)
		exclude, _ := task["exclude"].([]interface{})

		retention := ""
		if value, ok := task["lifetime_value"].(float64); ok {
			if unit, ok := task["lifetime_unit"].(string); ok {
				retention = fmt.Sprintf("%d %s", int(value), unit)
			}
		}

		summary = append(summary, map[string]interface{}{
			"dataset":       dataset,
			"recursive":     recursive,
			"enabled":       enabled,
			"allow_empty":   task["allow_empty"],
			"retention":     retention,
			"schedule":      scheduleString(task["schedule"]),
			"exclude_count": len(exclude),
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["dataset"].(string) < summary[j]["dataset"].(string)
	})
	return summary
}

// scheduleString renders a TrueNAS schedule object as "minute hour dom month dow", or "" when absent.
func scheduleString(value interface{}) string {
	schedule, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}
	fields := make([]string, 0, 5)
	for _, key := range []string{"minute", "hour", "dom", "month", "dow"} {
		text, _ := schedule[key].(string)
		if text == "" {
			return ""
		}
		fields = append(fields, text)
	}
	return fmt.Sprintf("%s %s %s %s %s", fields[0], fields[1], fields[2], fields[3], fields[4])
}

func handleQuerySnapshotTasks(client *truenas.Client, args map[string]interface{}) (string, error) {
	tasks, err := callClientList(client, "pool.snapshottask.query")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeSnapshotTasks(tasks), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format snapshot tasks: %w", err)
	}
	return string(formatted), nil
}
