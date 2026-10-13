package tools

import (
	"reflect"
	"testing"
	"time"
)

var certificateTestNow = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

func TestSummarizeCertificatesComputesDaysLeftAndOmitsKeyMaterial(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name": "web", "cert_type": "CERTIFICATE", "common": "nas.example", "san": []interface{}{"nas.example", "nas"},
			"from": "2026-01-01T00:00:00Z", "until": "2026-10-23T12:00:00Z", "expired": false,
			"key_type": "RSA", "key_length": float64(2048), "digest_algorithm": "SHA256",
			"certificate": "-----BEGIN CERTIFICATE-----", "privatekey": "-----BEGIN PRIVATE KEY-----",
			"CSR": "-----BEGIN CERTIFICATE REQUEST-----", "chain_list": []interface{}{"pem"},
		},
	}

	got := summarizeCertificates(raw, certificateTestNow)

	want := []map[string]interface{}{
		{
			"name": "web", "cert_type": "CERTIFICATE", "common": "nas.example", "san": []interface{}{"nas.example", "nas"},
			"from": "2026-01-01T00:00:00Z", "until": "2026-10-23T12:00:00Z", "days_left": 10, "expired": false,
			"key_type": "RSA", "key_length": float64(2048), "digest_algorithm": "SHA256",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeCertificates =\n%#v\nwant\n%#v", got, want)
	}
	for _, forbidden := range []string{"certificate", "privatekey", "CSR", "chain_list"} {
		if _, present := got[0][forbidden]; present {
			t.Errorf("%q must not be echoed", forbidden)
		}
	}
}

func TestSummarizeCertificatesFlagsExpiredWithNegativeDaysLeft(t *testing.T) {
	raw := []map[string]interface{}{{"name": "old", "until": "2026-10-11T12:00:00Z", "expired": false}}

	got := summarizeCertificates(raw, certificateTestNow)

	if got[0]["days_left"] != -2 || got[0]["expired"] != true {
		t.Errorf("days_left=%v expired=%v, want -2 and true", got[0]["days_left"], got[0]["expired"])
	}
}

func TestSummarizeCertificatesToleratesNullUntil(t *testing.T) {
	raw := []map[string]interface{}{
		{"name": "broken", "until": nil, "expired": nil},
		{"name": "garbled", "until": "not a date", "expired": true},
		{"name": "bare"},
	}

	got := summarizeCertificates(raw, certificateTestNow)

	for _, entry := range got {
		if entry["days_left"] != nil {
			t.Errorf("%v: days_left = %v, want nil when until cannot be parsed", entry["name"], entry["days_left"])
		}
	}
	if got[0]["expired"] != false || got[1]["expired"] != false {
		t.Errorf("a missing until with no API flag must not be reported expired: %#v", got)
	}
	if got[2]["expired"] != true {
		t.Errorf("the API expired flag should be kept when until cannot be parsed")
	}
}

func TestSummarizeCertificatesAcceptsDateObjectAndSortsByName(t *testing.T) {
	untilMillis := float64(certificateTestNow.Add(48 * time.Hour).UnixMilli())
	raw := []map[string]interface{}{
		{"name": "zeta", "until": map[string]interface{}{"$date": untilMillis}},
		{"name": "alpha"},
	}

	got := summarizeCertificates(raw, certificateTestNow)

	if got[0]["name"] != "alpha" || got[1]["name"] != "zeta" {
		t.Fatalf("not sorted by name: %#v", got)
	}
	if got[1]["days_left"] != 2 {
		t.Errorf("days_left from a $date object = %v, want 2", got[1]["days_left"])
	}
}

func TestSummarizeCertificatesAcceptsCtimeDatesAsReportedLive(t *testing.T) {
	raw := []map[string]interface{}{{"name": "default", "until": "Thu Oct 22 12:00:00 2026"}}

	got := summarizeCertificates(raw, certificateTestNow)

	if got[0]["days_left"] != 9 || got[0]["expired"] != false {
		t.Errorf("days_left=%v expired=%v, want 9 and false", got[0]["days_left"], got[0]["expired"])
	}
}

func TestSummarizeCertificatesKeepsTheCertificateID(t *testing.T) {
	raw := []map[string]interface{}{{"id": float64(7), "name": "truenas_default", "until": "Thu Jan  1 00:00:00 2026"}}

	got := summarizeCertificates(raw, time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC))

	if got[0]["id"] != float64(7) {
		t.Errorf("id = %v, want 7 so the UI certificate can be identified", got[0]["id"])
	}
}

func TestParseUICertificateIDHandlesNumbersObjectsAndTheMaskedValue(t *testing.T) {
	cases := []struct {
		name   string
		value  interface{}
		wantID int
		wantOK bool
	}{
		{"number", float64(3), 3, true},
		{"object with id", map[string]interface{}{"id": float64(5)}, 5, true},
		{"masked for the read-only role", "********", 0, false},
		{"null", nil, 0, false},
		{"object without id", map[string]interface{}{"name": "x"}, 0, false},
	}
	for _, tc := range cases {
		id, ok := parseUICertificateID(tc.value)
		if id != tc.wantID || ok != tc.wantOK {
			t.Errorf("%s: parseUICertificateID(%v) = (%d, %v), want (%d, %v)", tc.name, tc.value, id, ok, tc.wantID, tc.wantOK)
		}
	}
}
