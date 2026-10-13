package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerScheduledJobTools registers the read-only view of cron jobs and replication tasks.
func registerScheduledJobTools(r *Registry) {
	r.tools["query_scheduled_jobs"] = Tool{
		Definition: mcp.Tool{
			Name: "query_scheduled_jobs",
			Description: "Read-only view of recurring jobs: cron jobs (cronjob.query) and replication tasks (replication.query) " +
				"with their schedules, for lining periodic I/O bursts up with scheduled work. Cron commands and " +
				"replication credentials are deliberately not returned",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryScheduledJobs,
	}
}

// summarizeCronJobs reduces cronjob.query entries to description, user, enabled and schedule,
// sorted by description. The command is omitted because it can embed credentials.
func summarizeCronJobs(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, job := range raw {
		description, _ := job["description"].(string)
		enabled, _ := job["enabled"].(bool)
		summary = append(summary, map[string]interface{}{
			"description": description,
			"user":        job["user"],
			"enabled":     enabled,
			"schedule":    scheduleString(job["schedule"]),
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["description"].(string) < summary[j]["description"].(string)
	})
	return summary
}

// summarizeReplications reduces replication.query entries to name, route, schedule and last state,
// sorted by name. SSH credentials and other attributes are omitted.
func summarizeReplications(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, task := range raw {
		name, _ := task["name"].(string)
		auto, _ := task["auto"].(bool)
		enabled, _ := task["enabled"].(bool)
		sources, ok := task["source_datasets"].([]interface{})
		if !ok {
			sources = []interface{}{}
		}
		state := ""
		if stateMap, ok := task["state"].(map[string]interface{}); ok {
			state, _ = stateMap["state"].(string)
		}
		summary = append(summary, map[string]interface{}{
			"name":            name,
			"direction":       task["direction"],
			"transport":       task["transport"],
			"auto":            auto,
			"enabled":         enabled,
			"source_datasets": sources,
			"target_dataset":  task["target_dataset"],
			"schedule":        scheduleString(task["schedule"]),
			"state":           state,
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["name"].(string) < summary[j]["name"].(string)
	})
	return summary
}

func handleQueryScheduledJobs(client *truenas.Client, args map[string]interface{}) (string, error) {
	cron, err := callClientList(client, "cronjob.query")
	if err != nil {
		return "", err
	}
	replication, err := callClientList(client, "replication.query")
	if err != nil {
		return "", err
	}
	response := map[string]interface{}{
		"cron_jobs":    summarizeCronJobs(cron),
		"replications": summarizeReplications(replication),
	}
	if rsync, problem := optionalClientList(client, "rsynctask.query"); problem != "" {
		response["rsync_tasks"] = "unavailable: " + problem
	} else {
		response["rsync_tasks"] = summarizeRsyncTasks(rsync)
	}
	if cloud, problem := optionalClientList(client, "cloudsync.query"); problem != "" {
		response["cloudsync_tasks"] = "unavailable: " + problem
	} else {
		response["cloudsync_tasks"] = summarizeCloudSyncTasks(cloud)
	}
	formatted, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format scheduled jobs: %w", err)
	}
	return string(formatted), nil
}
