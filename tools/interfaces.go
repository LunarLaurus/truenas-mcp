package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerInterfaceTools registers the read-only network interface view.
func registerInterfaceTools(r *Registry) {
	r.tools["query_interfaces"] = Tool{
		Definition: mcp.Tool{
			Name:        "query_interfaces",
			Description: "Read-only network interfaces from interface.query: type, MTU, link state and IP addresses per interface",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryInterfaces,
	}
}

// summarizeInterfaces reduces interface.query entries to name, type, MTU, link state and addresses, sorted by name.
// Fields the system does not report are omitted.
func summarizeInterfaces(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, entry := range raw {
		name, _ := entry["name"].(string)
		if name == "" {
			continue
		}
		item := map[string]interface{}{"name": name, "type": entry["type"]}
		addresses := []string{}

		if state, ok := entry["state"].(map[string]interface{}); ok {
			if mtu, ok := state["mtu"]; ok {
				item["mtu"] = mtu
			}
			if link, ok := state["link_state"].(string); ok && link != "" {
				item["link_state"] = link
			}
			if aliases, ok := state["aliases"].([]interface{}); ok {
				for _, alias := range aliases {
					if a, ok := alias.(map[string]interface{}); ok {
						addresses = appendInetAddress(addresses, a)
					}
				}
			}
		}
		item["addresses"] = addresses
		summary = append(summary, item)
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["name"].(string) < summary[j]["name"].(string)
	})
	return summary
}

// appendInetAddress adds "address/netmask" for INET and INET6 aliases and skips link-layer ones.
func appendInetAddress(addresses []string, alias map[string]interface{}) []string {
	aliasType, _ := alias["type"].(string)
	address, _ := alias["address"].(string)
	if (aliasType != "INET" && aliasType != "INET6") || address == "" {
		return addresses
	}
	if netmask, ok := alias["netmask"].(float64); ok {
		return append(addresses, fmt.Sprintf("%s/%d", address, int(netmask)))
	}
	return append(addresses, address)
}

func handleQueryInterfaces(client *truenas.Client, args map[string]interface{}) (string, error) {
	interfaces, err := callClientList(client, "interface.query")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeInterfaces(interfaces), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format interfaces: %w", err)
	}
	return string(formatted), nil
}
