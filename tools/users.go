package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// registerUserTools registers the read-only account audit view.
func registerUserTools(r *Registry) {
	r.tools["query_users"] = Tool{
		Definition: mcp.Tool{
			Name: "query_users",
			Description: "Read-only account audit from user.query: username, uid, built-in, locked, SMB, SSH password login " +
				"and whether the account has sudo rights. Password hashes, SSH keys and other credential fields are not returned",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryUsers,
	}
}

// nonEmptyList reports whether value is a JSON array with at least one entry.
func nonEmptyList(value interface{}) bool {
	list, ok := value.([]interface{})
	return ok && len(list) > 0
}

// summarizeUsers reduces user.query entries to an allowlist, sorted by username. Hashes, keys and passwords
// are never copied.
func summarizeUsers(raw []map[string]interface{}) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, user := range raw {
		username, _ := user["username"].(string)
		builtin, _ := user["builtin"].(bool)
		locked, _ := user["locked"].(bool)
		smb, _ := user["smb"].(bool)
		sshPassword, _ := user["ssh_password_enabled"].(bool)
		summary = append(summary, map[string]interface{}{
			"username":             username,
			"uid":                  user["uid"],
			"builtin":              builtin,
			"locked":               locked,
			"smb":                  smb,
			"ssh_password_enabled": sshPassword,
			"has_sudo":             nonEmptyList(user["sudo_commands"]) || nonEmptyList(user["sudo_commands_nopasswd"]),
		})
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i]["username"].(string) < summary[j]["username"].(string)
	})
	return summary
}

func handleQueryUsers(client *truenas.Client, args map[string]interface{}) (string, error) {
	users, err := callClientList(client, "user.query")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeUsers(users), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format users: %w", err)
	}
	return string(formatted), nil
}
