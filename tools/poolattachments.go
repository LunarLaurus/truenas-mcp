package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerPoolAttachmentTools registers the read-only tool listing what uses a pool.
func registerPoolAttachmentTools(r *Registry) {
	r.tools["get_pool_attachments"] = Tool{
		Definition: mcp.Tool{
			Name: "get_pool_attachments",
			Description: "Read-only list of the services and resources attached to a pool (pool.attachments, role POOL_READ): " +
				"shares, apps, VMs and similar, with the paths or identifiers involved. Shows what an export or " +
				"destroy would disrupt",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"pool": poolNameSchema},
				"required":   []string{"pool"},
			},
		},
		Handler: handleGetPoolAttachments,
	}
}

// summarizePoolAttachments reduces pool.attachments entries to type, service and sorted attachment
// identifiers, sorted by type then service. Missing fields are tolerated.
func summarizePoolAttachments(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, attachment := range raw {
		items := []string{}
		if listed, ok := attachment["attachments"].([]interface{}); ok {
			for _, item := range listed {
				if text, ok := item.(string); ok {
					items = append(items, text)
				}
			}
		}
		sort.Strings(items)
		summary = append(summary, map[string]interface{}{
			"type":        attachment["type"],
			"service":     attachment["service"],
			"attachments": items,
		})
	}
	sort.SliceStable(summary, func(i, j int) bool {
		leftType, _ := summary[i]["type"].(string)
		rightType, _ := summary[j]["type"].(string)
		if leftType != rightType {
			return leftType < rightType
		}
		leftService, _ := summary[i]["service"].(string)
		rightService, _ := summary[j]["service"].(string)
		return leftService < rightService
	})
	return summary
}

func handleGetPoolAttachments(client *truenas.Client, args map[string]interface{}) (string, error) {
	poolID, err := resolvePoolID(client, args)
	if err != nil {
		return "", err
	}
	raw, err := client.Call("pool.attachments", poolID)
	if err != nil {
		return "", err
	}
	var attachments []map[string]interface{}
	if err := json.Unmarshal(raw, &attachments); err != nil {
		return "", fmt.Errorf("failed to parse pool.attachments: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizePoolAttachments(attachments), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format pool attachments: %w", err)
	}
	return string(formatted), nil
}
