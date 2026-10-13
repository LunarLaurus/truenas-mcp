package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// ipmiChassisFields are the documented ipmi.chassis.info result fields.
var ipmiChassisFields = []string{
	"system_power", "power_overload", "interlock", "power_fault", "power_control_fault", "power_restore_policy",
	"last_power_event", "chassis_intrusion", "front_panel_lockout", "drive_fault", "cooling/fan_fault",
	"chassis_identify_state",
}

// registerIPMITools registers the read-only IPMI chassis tool.
func registerIPMITools(r *Registry) {
	r.tools["get_ipmi_chassis"] = Tool{
		Definition: mcp.Tool{
			Name: "get_ipmi_chassis",
			Description: "Read-only IPMI chassis status (ipmi.chassis.info, role IPMI_READ): power state, faults, " +
				"intrusion and identify state. Systems without a BMC return a short 'IPMI not available' message",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleGetIPMIChassis,
	}
}

// ipmiUnavailableMessage turns an ipmi.chassis.info failure into a one-line message using the first line
// of the error, so a system without a BMC reports plainly instead of failing the tool call.
func ipmiUnavailableMessage(err error) string {
	firstLine := "unknown error"
	if err != nil {
		if line := strings.TrimSpace(strings.SplitN(err.Error(), "\n", 2)[0]); line != "" {
			firstLine = line
		}
	}
	return "IPMI not available on this system: " + firstLine
}

// summarizeIPMIChassis keeps the documented chassis fields that are present.
func summarizeIPMIChassis(raw map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{}, len(ipmiChassisFields))
	for _, field := range ipmiChassisFields {
		if value, ok := raw[field]; ok {
			summary[field] = value
		}
	}
	return summary
}

func handleGetIPMIChassis(client *truenas.Client, args map[string]interface{}) (string, error) {
	raw, err := client.Call("ipmi.chassis.info", map[string]interface{}{"query-remote": false})
	if err != nil {
		return ipmiUnavailableMessage(err), nil
	}
	var chassis map[string]interface{}
	if err := json.Unmarshal(raw, &chassis); err != nil {
		return "", fmt.Errorf("failed to parse ipmi.chassis.info: %w", err)
	}
	formatted, err := json.MarshalIndent(summarizeIPMIChassis(chassis), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format IPMI chassis status: %w", err)
	}
	return string(formatted), nil
}
