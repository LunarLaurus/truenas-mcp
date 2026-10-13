package tools

import (
	"reflect"
	"testing"
)

func TestZFSResourceQueryParamsDefaultsProperties(t *testing.T) {
	got, err := zfsResourceQueryParams(map[string]interface{}{"paths": []interface{}{"tank/vm-block"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []interface{}{map[string]interface{}{
		"paths":        []interface{}{"tank/vm-block"},
		"properties":   []interface{}{"sync", "logbias", "recordsize", "compression", "dedup", "used", "available"},
		"get_children": false,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("zfsResourceQueryParams =\n%#v\nwant\n%#v", got, want)
	}
}

func TestZFSResourceQueryParamsUsesRequestedProperties(t *testing.T) {
	got, err := zfsResourceQueryParams(map[string]interface{}{
		"paths": []interface{}{"a", "b"}, "properties": []interface{}{"volblocksize", "", 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := got[0].(map[string]interface{})
	if !reflect.DeepEqual(data["properties"], []interface{}{"volblocksize"}) {
		t.Errorf("properties = %#v, want only the valid string", data["properties"])
	}
	if !reflect.DeepEqual(data["paths"], []interface{}{"a", "b"}) {
		t.Errorf("paths = %#v", data["paths"])
	}
}

func TestZFSResourceQueryParamsRequiresPaths(t *testing.T) {
	cases := []map[string]interface{}{{}, {"paths": []interface{}{}}, {"paths": []interface{}{"", 3}}}
	for _, args := range cases {
		if _, err := zfsResourceQueryParams(args); err == nil {
			t.Errorf("expected an error for %#v", args)
		}
	}
}

func TestSummarizeZFSResourcesUsesParsedValuesAndFallsBackToRaw(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name": "tank/zv", "type": "VOLUME", "pool": "tank", "guid": float64(1),
			"properties": map[string]interface{}{
				"sync":   map[string]interface{}{"raw": "standard", "value": "standard", "source": map[string]interface{}{"type": "DEFAULT"}},
				"used":   map[string]interface{}{"raw": "1048576", "value": float64(1048576)},
				"dedup":  map[string]interface{}{"raw": "off"},
				"hidden": map[string]interface{}{"raw": "x", "value": "x"},
			},
		},
		{"name": "tank/a", "type": "FILESYSTEM"},
	}

	got := summarizeZFSResources(raw, []string{"sync", "used", "dedup", "recordsize"})

	want := []map[string]interface{}{
		{"name": "tank/a", "type": "FILESYSTEM", "properties": map[string]interface{}{}},
		{
			"name": "tank/zv", "type": "VOLUME",
			"properties": map[string]interface{}{"sync": "standard", "used": float64(1048576), "dedup": "off"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeZFSResources =\n%#v\nwant\n%#v", got, want)
	}
}
