package tools

import (
	"reflect"
	"testing"
)

func TestJoinISCSIMapsEachExtentToItsTargets(t *testing.T) {
	targets := []map[string]interface{}{
		{"id": float64(1), "name": "esx-a", "alias": "ESX A", "groups": []interface{}{map[string]interface{}{"auth": float64(3)}}},
		{"id": float64(2), "name": "esx-b"},
	}
	extents := []map[string]interface{}{
		{"id": float64(10), "name": "vmfs", "type": "DISK", "disk": "zvol/Internal/Vmware-Block", "blocksize": float64(512), "enabled": true, "serial": "abc"},
		{"id": float64(11), "name": "spare", "type": "FILE", "path": "/mnt/p/f", "enabled": false},
	}
	links := []map[string]interface{}{
		{"target": float64(1), "extent": float64(10), "lunid": float64(0)},
		{"target": float64(2), "extent": float64(10), "lunid": float64(1)},
	}

	got := joinISCSI(targets, extents, links)

	want := []map[string]interface{}{
		{
			"extent": "vmfs", "type": "DISK", "backing": "zvol/Internal/Vmware-Block",
			"blocksize": float64(512), "enabled": true,
			"targets": []map[string]interface{}{
				{"target": "esx-a", "lunid": float64(0)},
				{"target": "esx-b", "lunid": float64(1)},
			},
		},
		{
			"extent": "spare", "type": "FILE", "backing": "/mnt/p/f",
			"enabled": false, "targets": []map[string]interface{}{},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("joinISCSI =\n%#v\nwant\n%#v", got, want)
	}
}

func TestJoinISCSIOmitsAuthAndSerialFields(t *testing.T) {
	extents := []map[string]interface{}{{"id": float64(1), "name": "e", "serial": "s", "naa": "0x1", "enabled": true}}

	got := joinISCSI(nil, extents, nil)

	for _, forbidden := range []string{"serial", "naa", "id"} {
		if _, present := got[0][forbidden]; present {
			t.Errorf("%q must not be echoed in the joined view", forbidden)
		}
	}
}
