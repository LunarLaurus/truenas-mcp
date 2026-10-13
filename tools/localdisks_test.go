package tools

import (
	"reflect"
	"testing"
)

func TestSanitizeDisksDropsSerialsIdentifiersAndCredentials(t *testing.T) {
	disks := []map[string]interface{}{{
		"name": "sda", "model": "EXAMPLE-3T", "size": float64(3000592982016), "type": "HDD", "bus": "SAS", "pool": "tank",
		"serial": "ABC123", "serial_lunid": "ABC123_5000", "lunid": "5000", "ident": "ABC123", "identifier": "{serial}ABC123",
		"passwd": "x", "kmip_uid": "y", "api_token": "z", "client_secret": "w",
	}}

	got := sanitizeDisks(disks, "")

	want := []map[string]interface{}{{
		"name": "sda", "model": "EXAMPLE-3T", "size": float64(3000592982016), "type": "HDD", "bus": "SAS", "pool": "tank",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sanitizeDisks =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSanitizeDisksFiltersByPool(t *testing.T) {
	disks := []map[string]interface{}{
		{"name": "sda", "pool": "tank"},
		{"name": "sdb", "pool": "backup"},
		{"name": "sdc", "pool": nil},
	}

	got := sanitizeDisks(disks, "backup")

	if len(got) != 1 || got[0]["name"] != "sdb" {
		t.Errorf("sanitizeDisks(pool=backup) = %#v, want only sdb", got)
	}
	if all := sanitizeDisks(disks, ""); len(all) != 3 {
		t.Errorf("no pool filter returned %d disks, want 3", len(all))
	}
}
