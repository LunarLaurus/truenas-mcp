package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeISCSISessionsKeepsEndpointsAndLimits(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"initiator": "iqn.1998-01.com.vmware:esx-b", "initiator_addr": "198.51.100.12", "target": "iqn.2005-10.org.freenas.ctl:esx",
			"header_digest": "None", "data_digest": "CRC32C", "max_burst_length": float64(262144),
			"first_burst_length": float64(65536), "max_data_segment_length": float64(8192), "immediate_data": true,
			"iser": false, "offload": false, "chap_secret": "SECRET", "auth": "hunter2",
		},
		{"initiator": "iqn.1998-01.com.vmware:esx-a", "initiator_addr": "198.51.100.11", "target": "iqn.2005-10.org.freenas.ctl:esx"},
	}

	got := summarizeISCSISessions(raw)

	want := []map[string]interface{}{
		{
			"initiator": "iqn.1998-01.com.vmware:esx-a", "initiator_addr": "198.51.100.11",
			"target": "iqn.2005-10.org.freenas.ctl:esx",
		},
		{
			"initiator": "iqn.1998-01.com.vmware:esx-b", "initiator_addr": "198.51.100.12",
			"target": "iqn.2005-10.org.freenas.ctl:esx", "header_digest": "None", "data_digest": "CRC32C",
			"max_burst_length": float64(262144), "first_burst_length": float64(65536),
			"max_data_segment_length": float64(8192), "immediate_data": true, "iser": false, "offload": false,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeISCSISessions =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeISCSISessionsOmitsAuthFields(t *testing.T) {
	raw := []map[string]interface{}{{"initiator": "i", "chap_secret": "s", "auth": "a", "peer_secret": "p"}}

	got := summarizeISCSISessions(raw)

	for _, forbidden := range []string{"chap_secret", "auth", "peer_secret"} {
		if _, present := got[0][forbidden]; present {
			t.Errorf("%q must not be echoed", forbidden)
		}
	}
}
