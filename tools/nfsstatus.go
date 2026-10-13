package tools

import (
	"encoding/json"
	"fmt"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// nfsConfigFields are the nfs.config properties worth reporting for performance work.
var nfsConfigFields = []string{"servers", "managed_nfsd", "protocols", "bindip", "rdma", "v4_krb_enabled", "allow_nonroot"}

// registerNFSStatusTools registers the read-only NFS server status tool.
func registerNFSStatusTools(r *Registry) {
	r.tools["get_nfs_status"] = Tool{
		Definition: mcp.Tool{
			Name: "get_nfs_status",
			Description: "Read-only NFS server status (role SHARING_NFS_READ): nfsd thread count, protocols, bind addresses, " +
				"and the connected NFSv3 and NFSv4 clients (nfs.config, nfs.get_nfs3_clients, nfs.get_nfs4_clients). "+
				"The client lists come from the server's mount records and can be incomplete: on one 25.10.1 system an export mounted by two hosts listed only one",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleGetNFSStatus,
	}
}

// summarizeNFSStatus keeps the performance-relevant nfs.config fields and adds client lists with counts.
func summarizeNFSStatus(config map[string]interface{}, nfs3 []map[string]interface{}, nfs4 []map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{}, len(nfsConfigFields)+4)
	for _, field := range nfsConfigFields {
		if value, ok := config[field]; ok {
			summary[field] = value
		}
	}
	summary["nfs3_client_count"] = len(nfs3)
	summary["nfs4_client_count"] = len(nfs4)
	summary["nfs3_clients"] = nfs3
	summary["nfs4_clients"] = nfs4
	return summary
}

func handleGetNFSStatus(client *truenas.Client, args map[string]interface{}) (string, error) {
	rawConfig, err := client.Call("nfs.config")
	if err != nil {
		return "", err
	}
	var config map[string]interface{}
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return "", fmt.Errorf("failed to parse nfs.config: %w", err)
	}

	nfs3, err := callClientList(client, "nfs.get_nfs3_clients")
	if err != nil {
		return "", err
	}
	nfs4, err := callClientList(client, "nfs.get_nfs4_clients")
	if err != nil {
		return "", err
	}

	formatted, err := json.MarshalIndent(summarizeNFSStatus(config, nfs3, nfs4), "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// callClientList calls a query-style method with no filters and parses a list of objects.
func callClientList(client *truenas.Client, method string) ([]map[string]interface{}, error) {
	raw, err := client.Call(method, []interface{}{}, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var clients []map[string]interface{}
	if err := json.Unmarshal(raw, &clients); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", method, err)
	}
	return clients, nil
}
