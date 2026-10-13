package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeCronJobsOmitsCommandAndFormatsSchedule(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"description": "nightly report", "user": "root", "enabled": true,
			"command":  "curl -u admin:hunter2 https://example/report",
			"schedule": map[string]interface{}{"minute": "0", "hour": "3", "dom": "*", "month": "*", "dow": "*"},
		},
		{"description": "alpha", "user": "backup", "enabled": false},
	}

	got := summarizeCronJobs(raw)

	want := []map[string]interface{}{
		{"description": "alpha", "user": "backup", "enabled": false, "schedule": ""},
		{"description": "nightly report", "user": "root", "enabled": true, "schedule": "0 3 * * *"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeCronJobs =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeReplicationsKeepsRouteScheduleAndState(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name": "offsite", "direction": "PUSH", "transport": "SSH", "auto": true, "enabled": true,
			"source_datasets": []interface{}{"IBM", "Television"}, "target_dataset": "backup/tank",
			"schedule":        map[string]interface{}{"minute": "0", "hour": "*", "dom": "*", "month": "*", "dow": "*"},
			"state":           map[string]interface{}{"state": "FINISHED"},
			"ssh_credentials": map[string]interface{}{"attributes": map[string]interface{}{"private_key": "SECRET"}},
		},
		{"name": "local", "auto": false},
	}

	got := summarizeReplications(raw)

	want := []map[string]interface{}{
		{
			"name": "local", "direction": nil, "transport": nil, "auto": false, "enabled": false,
			"source_datasets": []interface{}{}, "target_dataset": nil, "schedule": "", "state": "",
		},
		{
			"name": "offsite", "direction": "PUSH", "transport": "SSH", "auto": true, "enabled": true,
			"source_datasets": []interface{}{"IBM", "Television"}, "target_dataset": "backup/tank",
			"schedule": "0 * * * *", "state": "FINISHED",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeReplications =\n%#v\nwant\n%#v", got, want)
	}
}
