package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// defaultZFSProperties are the properties reported when the caller names none.
var defaultZFSProperties = []string{"sync", "logbias", "recordsize", "compression", "dedup", "used", "available"}

// registerZFSResourceTools registers the read-only ZFS property lookup.
func registerZFSResourceTools(r *Registry) {
	r.tools["query_zfs_resources"] = Tool{
		Definition: mcp.Tool{
			Name: "query_zfs_resources",
			Description: "Read-only ZFS dataset and zvol properties (zfs.resource.query, role ZFS_RESOURCE_READ) for the " +
				"given paths: name, type and the requested property values (parsed). Children are not included. " +
				"The API documents this method as deprecated and scheduled for removal in v28",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Dataset or zvol paths, for example 'Internal/Vmware-Block'",
					},
					"properties": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"type": "string"},
						"description": "Optional: ZFS property names (default: sync, logbias, recordsize, " +
							"compression, dedup, used, available)",
					},
				},
				"required": []string{"paths"},
			},
		},
		Handler: handleQueryZFSResources,
	}
}

// stringList returns the non-empty strings of a decoded JSON array, ignoring other element types.
func stringList(value interface{}) []string {
	items, _ := value.([]interface{})
	names := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok && text != "" {
			names = append(names, text)
		}
	}
	return names
}

// zfsRequestedProperties returns the property names asked for, or the defaults when none are valid.
func zfsRequestedProperties(args map[string]interface{}) []string {
	if names := stringList(args["properties"]); len(names) > 0 {
		return names
	}
	return defaultZFSProperties
}

// toInterfaceList converts a string slice to a []interface{} for use as a call parameter.
func toInterfaceList(names []string) []interface{} {
	list := make([]interface{}, 0, len(names))
	for _, name := range names {
		list = append(list, name)
	}
	return list
}

// zfsResourceQueryParams builds the zfs.resource.query call parameters: the paths, the requested
// properties (default set when none are given) and no children.
func zfsResourceQueryParams(args map[string]interface{}) ([]interface{}, error) {
	paths := stringList(args["paths"])
	if len(paths) == 0 {
		return nil, fmt.Errorf("the 'paths' argument is required and must list at least one dataset or zvol path")
	}
	return []interface{}{map[string]interface{}{
		"paths":        toInterfaceList(paths),
		"properties":   toInterfaceList(zfsRequestedProperties(args)),
		"get_children": false,
	}}, nil
}

// summarizeZFSResources reduces zfs.resource.query entries to name, type and the requested properties.
// Each property is reported as its parsed value, falling back to the raw string, and is left out when
// the entry does not carry it. Entries are sorted by name.
func summarizeZFSResources(raw []map[string]interface{}, requested []string) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, resource := range raw {
		properties := make(map[string]interface{}, len(requested))
		available, _ := resource["properties"].(map[string]interface{})
		for _, name := range requested {
			property, ok := available[name].(map[string]interface{})
			if !ok {
				continue
			}
			if value, ok := property["value"]; ok && value != nil {
				properties[name] = value
			} else if rawValue, ok := property["raw"]; ok {
				properties[name] = rawValue
			}
		}
		summary = append(summary, map[string]interface{}{
			"name":       resource["name"],
			"type":       resource["type"],
			"properties": properties,
		})
	}
	sort.SliceStable(summary, func(i, j int) bool {
		left, _ := summary[i]["name"].(string)
		right, _ := summary[j]["name"].(string)
		return left < right
	})
	return summary
}

func handleQueryZFSResources(client *truenas.Client, args map[string]interface{}) (string, error) {
	params, err := zfsResourceQueryParams(args)
	if err != nil {
		return "", err
	}
	raw, err := client.Call("zfs.resource.query", params...)
	if err != nil {
		return "", err
	}
	var resources []map[string]interface{}
	if err := json.Unmarshal(raw, &resources); err != nil {
		return "", fmt.Errorf("failed to parse zfs.resource.query: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeZFSResources(resources, zfsRequestedProperties(args)), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format ZFS resources: %w", err)
	}
	return string(formatted), nil
}
