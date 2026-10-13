package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeNFSStatusKeepsDocumentedConfigFieldsAndCountsClients(t *testing.T) {
	config := map[string]interface{}{
		"id":           float64(1),
		"servers":      float64(16),
		"managed_nfsd": true,
		"protocols":    []interface{}{"NFSV3", "NFSV4"},
		"bindip":       []interface{}{},
		"rdma":         false,
		"mountd_port":  float64(618),
		"v4_domain":    "",
	}
	v3 := []map[string]interface{}{{"ip": "192.0.2.11"}, {"ip": "192.0.2.16"}}
	v4 := []map[string]interface{}{{"info": map[string]interface{}{"clientid": float64(1)}}}

	got := summarizeNFSStatus(config, v3, v4)

	want := map[string]interface{}{
		"servers":           float64(16),
		"managed_nfsd":      true,
		"protocols":         []interface{}{"NFSV3", "NFSV4"},
		"bindip":            []interface{}{},
		"rdma":              false,
		"nfs3_client_count": 2,
		"nfs4_client_count": 1,
		"nfs3_clients":      v3,
		"nfs4_clients":      v4,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeNFSStatus =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeNFSStatusToleratesMissingConfigAndClients(t *testing.T) {
	got := summarizeNFSStatus(nil, nil, nil)

	if got["nfs3_client_count"] != 0 || got["nfs4_client_count"] != 0 {
		t.Errorf("empty input must report zero clients, got %#v", got)
	}
	if _, present := got["servers"]; present {
		t.Error("servers must be omitted when the config does not report it")
	}
}
