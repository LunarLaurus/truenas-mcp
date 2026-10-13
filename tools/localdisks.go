package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerLocalDiskTools registers query_disks: a read-only disk inventory via disk.query (model, size,
// type, bus, pool membership). Serial numbers, LUN identifiers and credential-like fields are omitted,
// matching get_disk_usage.
func registerLocalDiskTools(r *Registry) {
	r.tools["query_disks"] = Tool{
		Definition: mcp.Tool{
			Name:        "query_disks",
			Description: "Read-only disk inventory from disk.query: name, model, size, type, bus and pool membership. Serial numbers and identifiers are not returned",
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
	out := sanitizeDisks(disks, want)
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// diskIdentityFields are disk.query fields that identify a physical device or its keys. They are never returned.
var diskIdentityFields = map[string]bool{
	"serial": true, "serial_lunid": true, "lunid": true, "ident": true, "identifier": true, "kmip_uid": true,
}

// sanitizeDisks drops identity and credential-like fields from every disk and keeps only the disks in pool
// when pool is not empty.
func sanitizeDisks(disks []map[string]interface{}, pool string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(disks))
	for _, d := range disks {
		if pool != "" {
			if p, _ := d["pool"].(string); p != pool {
				continue
			}
		}
		for k := range d {
			lk := strings.ToLower(k)
			if diskIdentityFields[lk] || strings.Contains(lk, "passw") || strings.Contains(lk, "secret") || strings.Contains(lk, "token") {
				delete(d, k)
			}
		}
		out = append(out, d)
	}
	return out
}
