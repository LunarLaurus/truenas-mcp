package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeRsyncTasksOmitsRemoteAuthAndFormatsSchedule(t *testing.T) {
	raw := []map[string]interface{}{{
		"desc": "photos", "path": "/mnt/tank/photos", "remotehost": "backup.lan", "direction": "PUSH", "enabled": true,
		"user": "root", "remotepass": "hunter2", "ssh_credentials": map[string]interface{}{"x": "y"},
		"schedule": map[string]interface{}{"minute": "30", "hour": "2", "dom": "*", "month": "*", "dow": "*"},
	}}

	got := summarizeRsyncTasks(raw)

	want := []map[string]interface{}{{
		"description": "photos", "path": "/mnt/tank/photos", "remote_host": "backup.lan", "direction": "PUSH",
		"enabled": true, "schedule": "30 2 * * *",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeRsyncTasks =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeCloudSyncTasksKeepsAllowListedFieldsOnly(t *testing.T) {
	raw := []map[string]interface{}{{
		"description": "offsite", "direction": "PUSH", "transfer_mode": "SYNC", "path": "/mnt/tank", "enabled": true,
		"credentials": map[string]interface{}{"provider": map[string]interface{}{"secret": "s3cr3t"}},
		"schedule":    map[string]interface{}{"minute": "0", "hour": "1", "dom": "*", "month": "*", "dow": "*"},
		"job":         map[string]interface{}{"state": "SUCCESS"},
	}}

	got := summarizeCloudSyncTasks(raw)

	want := []map[string]interface{}{{
		"description": "offsite", "direction": "PUSH", "transfer_mode": "SYNC", "path": "/mnt/tank", "enabled": true,
		"schedule": "0 1 * * *", "last_state": "SUCCESS",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeCloudSyncTasks =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeReportingExportersNeverEchoesAttributes(t *testing.T) {
	raw := []map[string]interface{}{{
		"name": "graphite", "enabled": true,
		"attributes": map[string]interface{}{"exporter_type": "GRAPHITE", "destination_ip": "198.51.100.5", "secret": "x"},
	}}

	got := summarizeReportingExporters(raw)

	want := []map[string]interface{}{{"name": "graphite", "enabled": true, "type": "GRAPHITE", "destination": "198.51.100.5"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeReportingExporters =\n%#v\nwant\n%#v", got, want)
	}
}

func TestMarkUICertificateFlagsTheCertificateTheWebUIUses(t *testing.T) {
	certs := []map[string]interface{}{
		{"id": float64(1), "name": "old"},
		{"id": float64(2), "name": "truenas_default"},
	}

	got := markUICertificate(certs, 2)

	if got[0]["is_ui_certificate"] != false || got[1]["is_ui_certificate"] != true {
		t.Errorf("markUICertificate = %#v, want only id 2 flagged", got)
	}
}
