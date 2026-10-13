package tools

import (
	"reflect"
	"testing"
)

func TestDatasetQuotaParamsDefaultsToDatasetQuota(t *testing.T) {
	got, err := datasetQuotaParams(map[string]interface{}{"dataset": "Internal/Vmware"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []interface{}{"Internal/Vmware", "DATASET", []interface{}{}, map[string]interface{}{}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("datasetQuotaParams = %#v, want %#v", got, want)
	}
}

func TestDatasetQuotaParamsNormalisesQuotaType(t *testing.T) {
	got, err := datasetQuotaParams(map[string]interface{}{"dataset": "p/d", "quota_type": "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[1] != "USER" {
		t.Errorf("quota_type = %v, want USER", got[1])
	}
}

func TestDatasetQuotaParamsRejectsBadInput(t *testing.T) {
	cases := []map[string]interface{}{
		{},
		{"dataset": ""},
		{"dataset": "p/d", "quota_type": "VOLUME"},
	}
	for _, args := range cases {
		if _, err := datasetQuotaParams(args); err == nil {
			t.Errorf("expected an error for %#v", args)
		}
	}
}

func TestSummarizeDatasetQuotaKeepsDocumentedFieldsOnly(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"quota_type": "USER", "id": float64(1000), "name": "alice", "used_bytes": float64(4096),
			"quota": float64(1048576), "obj_used": float64(3), "obj_quota": float64(100), "unexpected": "x",
		},
		{"quota_type": "USER", "id": float64(0), "name": nil, "used_bytes": float64(10)},
	}

	got := summarizeDatasetQuota(raw)

	want := []map[string]interface{}{
		{
			"quota_type": "USER", "id": float64(1000), "name": "alice", "used_bytes": float64(4096),
			"quota": float64(1048576), "obj_used": float64(3), "obj_quota": float64(100),
		},
		{"quota_type": "USER", "id": float64(0), "name": nil, "used_bytes": float64(10)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeDatasetQuota =\n%#v\nwant\n%#v", got, want)
	}
}
