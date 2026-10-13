package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeServicesNormalisesNameAndCountsPids(t *testing.T) {
	raw := []map[string]interface{}{
		{"service": "nfs", "state": "RUNNING", "enable": true, "pids": []interface{}{float64(10), float64(11)}},
		{"name": "cifs", "state": "STOPPED", "enable": false},
		{"state": "RUNNING"},
	}

	got := summarizeServices(raw, false)

	want := []map[string]interface{}{
		{"service": "cifs", "state": "STOPPED", "enable": false, "pid_count": 0},
		{"service": "nfs", "state": "RUNNING", "enable": true, "pid_count": 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeServices =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeServicesRunningOnly(t *testing.T) {
	raw := []map[string]interface{}{
		{"service": "nfs", "state": "RUNNING", "enable": true},
		{"service": "cifs", "state": "STOPPED", "enable": false},
	}

	got := summarizeServices(raw, true)

	if len(got) != 1 || got[0]["service"] != "nfs" {
		t.Errorf("running-only result = %#v, want only nfs", got)
	}
}
