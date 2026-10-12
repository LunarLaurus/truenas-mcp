package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// Local addition: read-only disk inventory via disk.query (identity, bus,
// SMART capability, pool membership). Credentials-like fields are dropped.
func registerLocalDiskTools(r *Registry) {
	r.tools["query_disks"] = Tool{
		Definition: mcp.Tool{
			Name:        "query_disks",
			Description: "Read-only disk inventory from disk.query: name, model, serial, size, type, bus, SMART capability and pool membership (local addition)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pool": map[string]interface{}{"type": "string", "description": "Optional: only disks in this pool"},
				},
			},
		},
		Handler: handleQueryDisks,
	}
}

func handleQueryDisks(client *truenas.Client, args map[string]interface{}) (string, error) {
	result, err := client.Call("disk.query", []interface{}{}, map[string]interface{}{
		"extra": map[string]interface{}{"pools": true},
	})
	if err != nil {
		result, err = client.Call("disk.query")
		if err != nil {
			return "", err
		}
	}
	var disks []map[string]interface{}
	if err := json.Unmarshal(result, &disks); err != nil {
		return "", fmt.Errorf("failed to parse disks: %w", err)
	}
	want, _ := args["pool"].(string)
	out := make([]map[string]interface{}, 0, len(disks))
	for _, d := range disks {
		if want != "" {
			if p, _ := d["pool"].(string); p != want {
				continue
			}
		}
		for k := range d {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "passw") || strings.Contains(lk, "secret") || strings.Contains(lk, "token") {
				delete(d, k)
			}
		}
		out = append(out, d)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
