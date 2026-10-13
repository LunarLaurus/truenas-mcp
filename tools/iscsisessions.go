package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// iscsiSessionFields are the iscsi.global.sessions fields worth reporting. Everything else is dropped
// so that no authentication detail can leak through.
var iscsiSessionFields = []string{
	"initiator", "initiator_addr", "initiator_alias", "target", "target_alias",
	"header_digest", "data_digest", "max_data_segment_length", "max_receive_data_segment_length",
	"max_xmit_data_segment_length", "max_burst_length", "first_burst_length", "immediate_data", "iser", "offload",
}

// registerISCSISessionTools registers the read-only iSCSI session view.
func registerISCSISessionTools(r *Registry) {
	r.tools["get_iscsi_sessions"] = Tool{
		Definition: mcp.Tool{
			Name: "get_iscsi_sessions",
			Description: "Read-only active iSCSI sessions (iscsi.global.sessions, role SHARING_ISCSI_GLOBAL_READ): " +
				"initiator IQN and address, target IQN, digests and the negotiated burst and segment limits. " +
				"Authentication details are not returned",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleGetISCSISessions,
	}
}

// summarizeISCSISessions keeps only the allowlisted session fields that are present, sorted by target,
// then initiator, then initiator address.
func summarizeISCSISessions(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, session := range raw {
		entry := make(map[string]interface{}, len(iscsiSessionFields))
		for _, field := range iscsiSessionFields {
			if value, ok := session[field]; ok {
				entry[field] = value
			}
		}
		summary = append(summary, entry)
	}
	sort.SliceStable(summary, func(i, j int) bool {
		for _, field := range []string{"target", "initiator", "initiator_addr"} {
			left, _ := summary[i][field].(string)
			right, _ := summary[j][field].(string)
			if left != right {
				return left < right
			}
		}
		return false
	})
	return summary
}

func handleGetISCSISessions(client *truenas.Client, args map[string]interface{}) (string, error) {
	sessions, err := callClientList(client, "iscsi.global.sessions")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeISCSISessions(sessions), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format iSCSI sessions: %w", err)
	}
	return string(formatted), nil
}
