package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeSnapshotTasksFormatsScheduleAndRetention(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"dataset": "tank/vm-datastore", "recursive": true, "enabled": true, "allow_empty": false,
			"lifetime_value": float64(2), "lifetime_unit": "WEEK",
			"exclude":  []interface{}{"tank/vm-datastore/tmp"},
			"schedule": map[string]interface{}{"minute": "0", "hour": "*", "dom": "*", "month": "*", "dow": "*"},
		},
		{"dataset": "tank/shared", "enabled": false},
		{"recursive": true},
	}

	got := summarizeSnapshotTasks(raw)

	want := []map[string]interface{}{
		{
			"dataset": "tank/shared", "recursive": false, "enabled": false, "allow_empty": nil,
			"retention": "", "schedule": "", "exclude_count": 0,
		},
		{
			"dataset": "tank/vm-datastore", "recursive": true, "enabled": true, "allow_empty": false,
			"retention": "2 WEEK", "schedule": "0 * * * *", "exclude_count": 1,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeSnapshotTasks =\n%#v\nwant\n%#v", got, want)
	}
}
