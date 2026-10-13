package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeInterfacesPicksNameTypeMTUAndAddresses(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name": "eno1",
			"type": "PHYSICAL",
			"state": map[string]interface{}{
				"mtu":        float64(9000),
				"link_state": "LINK_STATE_UP",
				"aliases": []interface{}{
					map[string]interface{}{"type": "INET", "address": "192.0.2.12", "netmask": float64(24)},
					map[string]interface{}{"type": "LINK", "address": "aa:bb:cc:dd:ee:ff"},
				},
			},
		},
		{"name": "br0", "type": "BRIDGE", "state": map[string]interface{}{"mtu": float64(1500)}},
		{"type": "VLAN"},
	}

	got := summarizeInterfaces(raw)

	want := []map[string]interface{}{
		{"name": "br0", "type": "BRIDGE", "mtu": float64(1500), "addresses": []string{}},
		{"name": "eno1", "type": "PHYSICAL", "mtu": float64(9000), "link_state": "LINK_STATE_UP", "addresses": []string{"192.0.2.12/24"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeInterfaces =\n%#v\nwant\n%#v", got, want)
	}
}
