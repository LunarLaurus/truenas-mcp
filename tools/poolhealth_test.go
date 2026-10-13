package tools

import (
	"reflect"
	"testing"
)

func sampleTankPool() map[string]interface{} {
	mirror := func(errs float64) map[string]interface{} {
		return map[string]interface{}{
			"type": "MIRROR", "name": "mirror-0",
			"stats":    map[string]interface{}{"read_errors": errs, "write_errors": float64(0), "checksum_errors": float64(0)},
			"children": []interface{}{map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{}},
		}
	}
	return map[string]interface{}{
		"name": "tank", "status": "ONLINE", "healthy": true, "status_code": "OK", "status_detail": nil,
		"size": float64(4000), "allocated": float64(1000), "free": float64(3000),
		"fragmentation": "59", "dedup_table_size": float64(1073741824), "dedup_table_quota": "auto",
		"autotrim": map[string]interface{}{"value": "off"},
		"scan": map[string]interface{}{
			"function": "SCRUB", "state": "FINISHED", "errors": float64(0), "percentage": 100.5,
			"end_time": map[string]interface{}{"$date": float64(1700000000000)},
		},
		"topology": map[string]interface{}{
			"data": []interface{}{mirror(0), mirror(3)},
			"log":  []interface{}{}, "cache": []interface{}{}, "special": []interface{}{}, "dedup": []interface{}{}, "spare": []interface{}{},
		},
	}
}

func TestSummarizePoolHealthReportsScanErrorsDedupAndMissingSpecialVdevs(t *testing.T) {
	got := summarizePoolHealth([]map[string]interface{}{sampleTankPool()})

	want := []map[string]interface{}{{
		"name": "tank", "status": "ONLINE", "healthy": true, "status_code": "OK",
		"used_percent": 25.0, "fragmentation": "59", "autotrim": "off",
		"dedup_table_bytes": float64(1073741824), "dedup_table_quota": "auto",
		"scan": map[string]interface{}{
			"function": "SCRUB", "state": "FINISHED", "errors": float64(0), "percent": 100.0, "ended_ms": float64(1700000000000),
		},
		"data_vdevs": []map[string]interface{}{
			{"type": "MIRROR", "disks": 4, "read_errors": float64(0), "write_errors": float64(0), "checksum_errors": float64(0)},
			{"type": "MIRROR", "disks": 4, "read_errors": float64(3), "write_errors": float64(0), "checksum_errors": float64(0)},
		},
		"has_log": false, "has_cache": false, "has_special": false, "has_dedup_vdev": false,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizePoolHealth =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizePoolHealthDetectsDedicatedLogDevice(t *testing.T) {
	pool := sampleTankPool()
	pool["topology"].(map[string]interface{})["log"] = []interface{}{map[string]interface{}{"type": "DISK"}}

	got := summarizePoolHealth([]map[string]interface{}{pool})

	if got[0]["has_log"] != true {
		t.Errorf("has_log = %v, want true when a log vdev exists", got[0]["has_log"])
	}
}

func TestSummarizePoolHealthToleratesNoScanAndNoTopology(t *testing.T) {
	got := summarizePoolHealth([]map[string]interface{}{{"name": "bare", "size": float64(0)}})

	if got[0]["scan"] != nil {
		t.Errorf("scan = %v, want nil when the pool has no scan record", got[0]["scan"])
	}
	if got[0]["used_percent"] != nil {
		t.Errorf("used_percent = %v, want nil for a zero-size pool", got[0]["used_percent"])
	}
}
