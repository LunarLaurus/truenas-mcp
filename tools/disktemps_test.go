package tools

import (
	"reflect"
	"testing"
)

func TestDiskNamesFromQuerySkipsEntriesWithoutAName(t *testing.T) {
	raw := []map[string]interface{}{
		{"name": "sda"},
		{"name": ""},
		{"serial": "nameless"},
		{"name": "nvme0n1"},
	}

	got := diskNamesFromQuery(raw)

	want := []string{"sda", "nvme0n1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diskNamesFromQuery = %v, want %v", got, want)
	}
}

func TestDiskTemperatureParamsMatchDocumentedSignature(t *testing.T) {
	// disk.temperatures takes (name: []string, include_thresholds: bool).
	got := diskTemperatureParams([]string{"sda", "sdb"}, true)

	want := []interface{}{[]string{"sda", "sdb"}, true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diskTemperatureParams = %#v, want %#v", got, want)
	}
}

func TestDiskTemperatureAggParamsMatchDocumentedSignature(t *testing.T) {
	// disk.temperature_agg takes (names: []string, days: int); days defaults to 7.
	if got, want := diskTemperatureAggParams([]string{"sda"}, 0), []interface{}{[]string{"sda"}, 7}; !reflect.DeepEqual(got, want) {
		t.Errorf("default days: got %#v, want %#v", got, want)
	}
	if got, want := diskTemperatureAggParams([]string{"sda"}, 30), []interface{}{[]string{"sda"}, 30}; !reflect.DeepEqual(got, want) {
		t.Errorf("30 days: got %#v, want %#v", got, want)
	}
}
