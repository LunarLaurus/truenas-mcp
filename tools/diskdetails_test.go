package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeDiskDetailsAcceptsAFlatListAndMarksPoolMembership(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{
			"devname": "sdb", "model": "ST3000NM0023", "size": float64(3000592982016), "type": "HDD", "bus": "SAS",
			"rotationrate": float64(7200), "imported_zpool": "IBM", "exported_zpool": nil, "passwd": "masked", "serial": "ZZZ",
			"enclosure": map[string]interface{}{"number": float64(0), "slot": float64(4)},
		},
		map[string]interface{}{"devname": "sda", "model": "SSD", "size": float64(120034123776), "type": "SSD", "bus": "SATA", "imported_zpool": nil, "exported_zpool": "oldpool"},
	}

	got := summarizeDiskDetails(raw)

	want := []map[string]interface{}{
		{"name": "sda", "model": "SSD", "size_gb": 111.8, "type": "SSD", "bus": "SATA", "rpm": nil, "pool": nil, "exported_pool": "oldpool", "in_use": false, "slot": nil},
		{"name": "sdb", "model": "ST3000NM0023", "size_gb": 2794.5, "type": "HDD", "bus": "SAS", "rpm": float64(7200), "pool": "IBM", "exported_pool": nil, "in_use": true, "slot": "0:4"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeDiskDetails =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeDiskDetailsOmitsPasswordsAndSerials(t *testing.T) {
	raw := []interface{}{map[string]interface{}{"devname": "sdc", "passwd": "x", "serial": "y", "kmip_uid": "z"}}

	got := summarizeDiskDetails(raw)

	for _, forbidden := range []string{"passwd", "serial", "kmip_uid"} {
		if _, present := got[0][forbidden]; present {
			t.Errorf("%q must not be echoed", forbidden)
		}
	}
}

func TestSummarizeDiskDetailsFlattensUsedAndUnusedGroups(t *testing.T) {
	raw := map[string]interface{}{
		"used":   []interface{}{map[string]interface{}{"devname": "sdb", "imported_zpool": "IBM"}},
		"unused": []interface{}{map[string]interface{}{"devname": "sdz", "imported_zpool": nil}},
	}

	got := summarizeDiskDetails(raw)

	if len(got) != 2 || got[0]["name"] != "sdb" || got[1]["name"] != "sdz" {
		t.Fatalf("summarizeDiskDetails = %#v, want sdb and sdz flattened and sorted", got)
	}
	if got[0]["in_use"] != true || got[1]["in_use"] != false {
		t.Errorf("in_use = %v, %v, want true, false", got[0]["in_use"], got[1]["in_use"])
	}
}

func TestSummarizeDiskDetailsTrimsPaddedModelAndKeepsEmptyEnclosureAsNoSlot(t *testing.T) {
	// Sun OEM drives report an empty enclosure object and space-padded identifiers.
	raw := []interface{}{map[string]interface{}{"devname": "sdq", "model": "H7230AS60SUN3.0T  ", "enclosure": map[string]interface{}{}}}

	got := summarizeDiskDetails(raw)

	if got[0]["slot"] != nil {
		t.Errorf("slot = %v, want nil for an empty enclosure object", got[0]["slot"])
	}
	if got[0]["model"] != "H7230AS60SUN3.0T" {
		t.Errorf("model = %q, want the padding trimmed", got[0]["model"])
	}
}
