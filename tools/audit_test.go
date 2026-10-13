package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeAuditKeepsSafeFieldsAndOmitsPayloads(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"timestamp": "2026-10-13T10:00:00Z", "username": "root", "address": "192.168.0.5",
			"event": "AUTHENTICATION", "success": true, "service": "MIDDLEWARE",
			"event_data":   map[string]interface{}{"credentials": map[string]interface{}{"password": "hunter2"}},
			"service_data": map[string]interface{}{"vers": "1"},
			"session":      "abc", "audit_id": "def",
		},
		{"username": "admin"},
	}

	got := summarizeAudit(raw)

	want := []map[string]interface{}{
		{
			"timestamp": "2026-10-13T10:00:00Z", "username": "root", "address": "192.168.0.5",
			"event": "AUTHENTICATION", "success": true,
		},
		{"timestamp": nil, "username": "admin", "address": "", "event": "", "success": false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeAudit =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSummarizeAuditNeverEchoesEventData(t *testing.T) {
	raw := []map[string]interface{}{{"event_data": "x", "service_data": "y", "session": "s", "audit_id": "a"}}

	got := summarizeAudit(raw)

	for _, forbidden := range []string{"event_data", "service_data", "session", "audit_id"} {
		if _, present := got[0][forbidden]; present {
			t.Errorf("%q must not be echoed", forbidden)
		}
	}
}

func TestAuditQueryParamsDefaultsAndNewestFirst(t *testing.T) {
	got, err := auditQueryParams(map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []interface{}{map[string]interface{}{
		"services":      []interface{}{"MIDDLEWARE"},
		"query-filters": []interface{}{},
		"query-options": map[string]interface{}{"order_by": []interface{}{"-message_timestamp"}, "limit": 25},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("auditQueryParams =\n%#v\nwant\n%#v", got, want)
	}
}

func TestAuditQueryParamsClampsLimitAndNormalisesService(t *testing.T) {
	cases := []struct {
		limit float64
		want  int
	}{{500, 200}, {200, 200}, {0, 25}, {-3, 25}, {10, 10}}
	for _, tc := range cases {
		got, err := auditQueryParams(map[string]interface{}{"service": "smb", "limit": tc.limit})
		if err != nil {
			t.Fatalf("limit %v: unexpected error: %v", tc.limit, err)
		}
		data := got[0].(map[string]interface{})
		options := data["query-options"].(map[string]interface{})
		if options["limit"] != tc.want {
			t.Errorf("limit %v -> %v, want %d", tc.limit, options["limit"], tc.want)
		}
		if !reflect.DeepEqual(data["services"], []interface{}{"SMB"}) {
			t.Errorf("services = %#v, want [SMB]", data["services"])
		}
	}
}

func TestAuditQueryParamsRejectsUnknownService(t *testing.T) {
	if _, err := auditQueryParams(map[string]interface{}{"service": "KERNEL"}); err == nil {
		t.Error("expected an error for an unknown audit service")
	}
}
