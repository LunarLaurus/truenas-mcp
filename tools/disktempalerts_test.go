package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeTemperatureAlertsKeepsReadableFields(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"level": "WARNING", "klass": "SMART", "source": "", "text": "Disk sdb is at 61 C", "dismissed": false,
			"last_occurrence": "2026-10-13T10:00:00Z", "uuid": "u", "key": "k", "mail": map[string]interface{}{},
			"formatted": "<b>x</b>", "args": map[string]interface{}{"device": "sdb"},
		},
		{"level": "CRITICAL", "text": "Disk sda is at 70 C"},
	}

	got := summarizeTemperatureAlerts(raw)

	want := []map[string]interface{}{
		{"level": "CRITICAL", "klass": "", "text": "Disk sda is at 70 C", "dismissed": false, "last_occurrence": nil},
		{
			"level": "WARNING", "klass": "SMART", "text": "Disk sdb is at 61 C", "dismissed": false,
			"last_occurrence": "2026-10-13T10:00:00Z",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeTemperatureAlerts =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeTemperatureAlertsEmptyIsEmptyList(t *testing.T) {
	got := summarizeTemperatureAlerts(nil)

	if got == nil || len(got) != 0 {
		t.Errorf("summarizeTemperatureAlerts(nil) = %#v, want an empty non-nil list", got)
	}
}
