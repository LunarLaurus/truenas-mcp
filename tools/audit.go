package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

const (
	defaultAuditLimit = 25
	maxAuditLimit     = 200
)

// auditServices are the services audit.query accepts.
var auditServices = map[string]bool{"MIDDLEWARE": true, "SMB": true, "SUDO": true, "SYSTEM": true}

// registerAuditTools registers the read-only audit log tool.
func registerAuditTools(r *Registry) {
	r.tools["query_audit"] = Tool{
		Definition: mcp.Tool{
			Name: "query_audit",
			Description: "Read-only audit log (audit.query, role SYSTEM_AUDIT_READ), newest first: timestamp, username, " +
				"address, event and success. Event payloads are deliberately not returned because they can contain secrets",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"MIDDLEWARE", "SMB", "SUDO", "SYSTEM"},
						"description": "Optional: audit service to read (default: MIDDLEWARE)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Optional: number of entries to return (default 25, maximum 200)",
					},
				},
			},
		},
		Handler: handleQueryAudit,
	}
}

// auditQueryParams builds the audit.query call parameters from the tool arguments: one service, newest
// first, with the limit defaulting to 25 and clamped to 200.
func auditQueryParams(args map[string]interface{}) ([]interface{}, error) {
	service := "MIDDLEWARE"
	if requested, ok := args["service"].(string); ok && requested != "" {
		service = strings.ToUpper(requested)
	}
	if !auditServices[service] {
		return nil, fmt.Errorf("unknown audit service %q: expected MIDDLEWARE, SMB, SUDO or SYSTEM", service)
	}

	limit := defaultAuditLimit
	if requested, ok := args["limit"].(float64); ok && requested >= 1 {
		limit = int(requested)
		if limit > maxAuditLimit {
			limit = maxAuditLimit
		}
	}

	return []interface{}{map[string]interface{}{
		"services":      []interface{}{service},
		"query-filters": []interface{}{},
		"query-options": map[string]interface{}{
			"order_by": []interface{}{"-message_timestamp"},
			"limit":    limit,
		},
	}}, nil
}

// summarizeAudit reduces audit entries to timestamp, username, address, event and success, keeping the
// input order. event_data, service_data and identifiers are omitted because they can contain secrets.
func summarizeAudit(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, entry := range raw {
		username, _ := entry["username"].(string)
		address, _ := entry["address"].(string)
		event, _ := entry["event"].(string)
		success, _ := entry["success"].(bool)
		summary = append(summary, map[string]interface{}{
			"timestamp": entry["timestamp"],
			"username":  username,
			"address":   address,
			"event":     event,
			"success":   success,
		})
	}
	return summary
}

func handleQueryAudit(client *truenas.Client, args map[string]interface{}) (string, error) {
	params, err := auditQueryParams(args)
	if err != nil {
		return "", err
	}
	raw, err := client.Call("audit.query", params...)
	if err != nil {
		return "", err
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(raw, &entries); err != nil {
		return "", fmt.Errorf("failed to parse audit.query: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeAudit(entries), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format audit log: %w", err)
	}
	return string(formatted), nil
}
