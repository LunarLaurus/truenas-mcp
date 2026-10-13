package tools

import (
	"sort"
	"testing"
)

// TrueNAS 25.10 (Goldeye) validates reporting.get_data graph names against a fixed set.
// arcfreememory and arcavailablememory are rejected by it.
func TestArcGraphNamesMatchGoldeyeReportingSet(t *testing.T) {
	want := []string{"arcactualrate", "arcrate", "arcresult", "arcsize"}

	got := append([]string(nil), arcGraphNames...)
	sort.Strings(got)

	if len(got) != len(want) {
		t.Fatalf("arcGraphNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arcGraphNames = %v, want %v", got, want)
		}
	}
}

func TestDefaultArcGraphsAreAllValidNames(t *testing.T) {
	valid := make(map[string]bool, len(arcGraphNames))
	for _, name := range arcGraphNames {
		valid[name] = true
	}
	if len(defaultArcGraphs) == 0 {
		t.Fatal("defaultArcGraphs must not be empty")
	}
	for _, name := range defaultArcGraphs {
		if !valid[name] {
			t.Errorf("default graph %q is not in arcGraphNames", name)
		}
	}
}
