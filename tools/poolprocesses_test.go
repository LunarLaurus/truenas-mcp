package tools

import (
	"reflect"
	"strings"
	"testing"
)

func TestPoolIDFromQueryMatchesByName(t *testing.T) {
	pools := []map[string]interface{}{
		{"id": float64(1), "name": "Internal"},
		{"id": float64(7), "name": "Television"},
	}

	got, err := poolIDFromQuery(pools, "Television")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Errorf("poolIDFromQuery = %d, want 7", got)
	}
}

func TestPoolIDFromQueryReportsUnknownPoolWithKnownNames(t *testing.T) {
	pools := []map[string]interface{}{{"id": float64(1), "name": "Internal"}, {"id": float64(2), "name": "Backup"}}

	_, err := poolIDFromQuery(pools, "Missing")

	if err == nil {
		t.Fatal("expected an error for an unknown pool")
	}
	if !strings.Contains(err.Error(), "Missing") || !strings.Contains(err.Error(), "Backup, Internal") {
		t.Errorf("error %q should name the pool and list the known pools", err)
	}
}

func TestTruncateCommandLineLimitsToOneHundredTwentyRunes(t *testing.T) {
	long := strings.Repeat("a", 300)

	got := truncateCommandLine(long)

	if want := strings.Repeat("a", 120) + "..."; got != want {
		t.Errorf("truncateCommandLine length %d, want 123 ending in ...", len(got))
	}
	if short := truncateCommandLine("smbd -D"); short != "smbd -D" {
		t.Errorf("short command line changed to %q", short)
	}
	multibyte := strings.Repeat("é", 130)
	if got := truncateCommandLine(multibyte); got != strings.Repeat("é", 120)+"..." {
		t.Errorf("multibyte command line was cut mid-rune: %q", got)
	}
}

func TestSummarizePoolProcessesTruncatesAndSorts(t *testing.T) {
	raw := []map[string]interface{}{
		{"pid": float64(900), "name": "smbd", "service": "cifs", "cmdline": strings.Repeat("x", 200)},
		{"pid": float64(12), "name": "nfsd", "service": nil, "cmdline": nil},
		{"pid": float64(5), "name": "nfsd"},
	}

	got := summarizePoolProcesses(raw)

	want := []map[string]interface{}{
		{"pid": float64(5), "name": "nfsd", "service": nil, "cmdline": ""},
		{"pid": float64(12), "name": "nfsd", "service": nil, "cmdline": ""},
		{"pid": float64(900), "name": "smbd", "service": "cifs", "cmdline": strings.Repeat("x", 120) + "..."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizePoolProcesses =\n%#v\nwant\n%#v", got, want)
	}
}
