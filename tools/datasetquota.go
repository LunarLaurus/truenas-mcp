package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// datasetQuotaFields are the pool.dataset.get_quota result fields reported. Absent fields (for example
// quota when none is set) are left out rather than invented.
var datasetQuotaFields = []string{
	"quota_type", "id", "name", "used_bytes", "quota", "obj_used", "obj_quota", "refquota",
}

// datasetQuotaTypes are the quota types pool.dataset.get_quota accepts.
var datasetQuotaTypes = map[string]bool{"USER": true, "GROUP": true, "DATASET": true, "PROJECT": true}

// registerDatasetQuotaTools registers the read-only dataset quota tool.
func registerDatasetQuotaTools(r *Registry) {
	r.tools["get_dataset_quota"] = Tool{
		Definition: mcp.Tool{
			Name: "get_dataset_quota",
			Description: "Read-only dataset quotas (pool.dataset.get_quota, role DATASET_READ): used bytes, quota and " +
				"object counts for the dataset itself, or per user, group or project",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dataset": map[string]interface{}{
						"type":        "string",
						"description": "Dataset path, for example 'Internal/Vmware-Dataset'",
					},
					"quota_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"USER", "GROUP", "DATASET", "PROJECT"},
						"description": "Optional: which quotas to read (default: DATASET)",
					},
				},
				"required": []string{"dataset"},
			},
		},
		Handler: handleGetDatasetQuota,
	}
}

// datasetQuotaParams builds the pool.dataset.get_quota call parameters (dataset, quota_type, filters,
// options) from the tool arguments, defaulting quota_type to DATASET.
func datasetQuotaParams(args map[string]interface{}) ([]interface{}, error) {
	dataset, _ := args["dataset"].(string)
	if dataset == "" {
		return nil, fmt.Errorf("the 'dataset' argument is required")
	}
	quotaType := "DATASET"
	if requested, ok := args["quota_type"].(string); ok && requested != "" {
		quotaType = strings.ToUpper(requested)
	}
	if !datasetQuotaTypes[quotaType] {
		return nil, fmt.Errorf("unknown quota_type %q: expected USER, GROUP, DATASET or PROJECT", quotaType)
	}
	return []interface{}{dataset, quotaType, []interface{}{}, map[string]interface{}{}}, nil
}

// summarizeDatasetQuota keeps the documented quota fields that are present on each entry.
func summarizeDatasetQuota(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, quota := range raw {
		entry := make(map[string]interface{}, len(datasetQuotaFields))
		for _, field := range datasetQuotaFields {
			if value, ok := quota[field]; ok {
				entry[field] = value
			}
		}
		summary = append(summary, entry)
	}
	return summary
}

func handleGetDatasetQuota(client *truenas.Client, args map[string]interface{}) (string, error) {
	params, err := datasetQuotaParams(args)
	if err != nil {
		return "", err
	}
	raw, err := client.Call("pool.dataset.get_quota", params...)
	if err != nil {
		return "", err
	}
	var quotas []map[string]interface{}
	if err := json.Unmarshal(raw, &quotas); err != nil {
		return "", fmt.Errorf("failed to parse pool.dataset.get_quota: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeDatasetQuota(quotas), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format dataset quotas: %w", err)
	}
	return string(formatted), nil
}
