package tools

import (
	"encoding/json"
	"fmt"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerISCSITools registers the read-only iSCSI view.
func registerISCSITools(r *Registry) {
	r.tools["query_iscsi"] = Tool{
		Definition: mcp.Tool{
			Name: "query_iscsi",
			Description: "Read-only iSCSI view: each extent with its backing zvol or file, block size, enabled state and the " +
				"targets and LUN ids that present it (iscsi.extent.query, iscsi.target.query, iscsi.targetextent.query). " +
				"Serials, NAA identifiers and authentication details are not returned",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryISCSI,
	}
}

// joinISCSI builds one entry per extent listing the targets that present it.
func joinISCSI(targets, extents, links []map[string]interface{}) []map[string]interface{} {
	targetNames := make(map[float64]string, len(targets))
	for _, target := range targets {
		id, idOK := target["id"].(float64)
		name, nameOK := target["name"].(string)
		if idOK && nameOK {
			targetNames[id] = name
		}
	}

	joined := make([]map[string]interface{}, 0, len(extents))
	for _, extent := range extents {
		entry := map[string]interface{}{
			"extent":  extent["name"],
			"type":    extent["type"],
			"enabled": extent["enabled"],
		}
		if backing, ok := extent["disk"].(string); ok && backing != "" {
			entry["backing"] = backing
		} else if path, ok := extent["path"].(string); ok && path != "" {
			entry["backing"] = path
		}
		if blocksize, ok := extent["blocksize"]; ok {
			entry["blocksize"] = blocksize
		}

		presenting := []map[string]interface{}{}
		extentID, _ := extent["id"].(float64)
		for _, link := range links {
			if linkExtent, _ := link["extent"].(float64); linkExtent != extentID {
				continue
			}
			targetID, _ := link["target"].(float64)
			presenting = append(presenting, map[string]interface{}{
				"target": targetNames[targetID],
				"lunid":  link["lunid"],
			})
		}
		entry["targets"] = presenting
		joined = append(joined, entry)
	}
	return joined
}

func handleQueryISCSI(client *truenas.Client, args map[string]interface{}) (string, error) {
	targets, err := callClientList(client, "iscsi.target.query")
	if err != nil {
		return "", err
	}
	extents, err := callClientList(client, "iscsi.extent.query")
	if err != nil {
		return "", err
	}
	links, err := callClientList(client, "iscsi.targetextent.query")
	if err != nil {
		return "", err
	}

	formatted, err := json.MarshalIndent(joinISCSI(targets, extents, links), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format iSCSI view: %w", err)
	}
	return string(formatted), nil
}
