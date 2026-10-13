package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerReportingExporterTools registers the read-only reporting exporter view.
func registerReportingExporterTools(r *Registry) {
	r.tools["query_reporting_exporters"] = Tool{
		Definition: mcp.Tool{
			Name: "query_reporting_exporters",
			Description: "Read-only reporting exporters (reporting.exporters.query, role REPORTING_READ): name, enabled, type " +
				"and destination address, to see whether metrics are already pushed to an external collector. " +
				"Exporter attributes are not returned because they can hold credentials",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryReportingExporters,
	}
}

// summarizeRsyncTasks reduces rsynctask.query entries to description, path, remote host, direction,
// enabled state and schedule, sorted by description. Remote passwords and SSH credentials are omitted.
func summarizeRsyncTasks(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, task := range raw {
		description, _ := task["desc"].(string)
		enabled, _ := task["enabled"].(bool)
		summary = append(summary, map[string]interface{}{
			"description": description,
			"path":        task["path"],
			"remote_host": task["remotehost"],
			"direction":   task["direction"],
			"enabled":     enabled,
			"schedule":    scheduleString(task["schedule"]),
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["description"].(string) < summary[j]["description"].(string)
	})
	return summary
}

// summarizeCloudSyncTasks reduces cloudsync.query entries to an allowlist of fields plus the last job
// state, sorted by description. Provider credentials are never copied.
func summarizeCloudSyncTasks(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, task := range raw {
		description, _ := task["description"].(string)
		enabled, _ := task["enabled"].(bool)
		lastState := ""
		if job, ok := task["job"].(map[string]interface{}); ok {
			lastState, _ = job["state"].(string)
		}
		summary = append(summary, map[string]interface{}{
			"description":   description,
			"direction":     task["direction"],
			"transfer_mode": task["transfer_mode"],
			"path":          task["path"],
			"enabled":       enabled,
			"schedule":      scheduleString(task["schedule"]),
			"last_state":    lastState,
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["description"].(string) < summary[j]["description"].(string)
	})
	return summary
}

// summarizeReportingExporters reduces reporting.exporters.query entries to name, enabled, exporter type
// and destination address. The attributes object is never echoed because it can hold credentials.
func summarizeReportingExporters(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, exporter := range raw {
		name, _ := exporter["name"].(string)
		enabled, _ := exporter["enabled"].(bool)
		attributes, _ := exporter["attributes"].(map[string]interface{})
		exporterType, _ := attributes["exporter_type"].(string)
		destination, _ := attributes["destination_ip"].(string)
		summary = append(summary, map[string]interface{}{
			"name":        name,
			"enabled":     enabled,
			"type":        exporterType,
			"destination": destination,
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["name"].(string) < summary[j]["name"].(string)
	})
	return summary
}

// markUICertificate sets is_ui_certificate on every certificate summary: true only for the one whose id
// matches the certificate the web UI is configured to use.
func markUICertificate(certificates []map[string]interface{}, uiCertificateID int) []map[string]interface{} {
	for _, certificate := range certificates {
		id, _ := certificate["id"].(float64)
		certificate["is_ui_certificate"] = int(id) == uiCertificateID
	}
	return certificates
}

// optionalClientList calls a query method and returns its rows, or the error text when the call fails,
// so one unavailable namespace does not hide the others.
func optionalClientList(client *truenas.Client, method string) ([]map[string]interface{}, string) {
	rows, err := callClientList(client, method)
	if err != nil {
		return nil, err.Error()
	}
	return rows, ""
}

func handleQueryReportingExporters(client *truenas.Client, args map[string]interface{}) (string, error) {
	exporters, err := callClientList(client, "reporting.exporters.query")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeReportingExporters(exporters), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format reporting exporters: %w", err)
	}
	return string(formatted), nil
}
