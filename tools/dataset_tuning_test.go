package tools

import "testing"

func TestSimplifyDatasetIncludesTuningProperties(t *testing.T) {
	ds := map[string]interface{}{
		"name": "tank/vm-datastore",
		"type": "FILESYSTEM",
		"pool": "tank",
		"recordsize": map[string]interface{}{
			"parsed": float64(131072), "value": "128 KiB",
		},
		"sync": map[string]interface{}{
			"parsed": "STANDARD", "value": "STANDARD",
		},
		"special_small_block_size": map[string]interface{}{
			"parsed": float64(0), "value": "0",
		},
	}

	got := simplifyDataset(ds)

	want := map[string]interface{}{
		"recordsize":           "128 KiB",
		"sync":                 "STANDARD",
		"special_small_blocks": "0",
	}
	for key, expected := range want {
		if got[key] != expected {
			t.Errorf("summary[%q] = %v, want %v", key, got[key], expected)
		}
	}
	if _, present := got["volblocksize"]; present {
		t.Error("volblocksize must be omitted when the dataset does not report it")
	}
}
