package tools

import (
	"errors"
	"reflect"
	"testing"
)

func TestIPMIUnavailableMessageUsesFirstLineOfError(t *testing.T) {
	err := errors.New("[ENOENT] No IPMI device found\nTraceback (most recent call last):\n  File x")

	got := ipmiUnavailableMessage(err)

	if want := "IPMI not available on this system: [ENOENT] No IPMI device found"; got != want {
		t.Errorf("ipmiUnavailableMessage = %q, want %q", got, want)
	}
}

func TestIPMIUnavailableMessageHandlesEmptyError(t *testing.T) {
	got := ipmiUnavailableMessage(errors.New(""))

	if want := "IPMI not available on this system: unknown error"; got != want {
		t.Errorf("ipmiUnavailableMessage = %q, want %q", got, want)
	}
}

func TestSummarizeIPMIChassisKeepsDocumentedFieldsOnly(t *testing.T) {
	raw := map[string]interface{}{
		"system_power": "on", "power_overload": "false", "cooling/fan_fault": "false",
		"chassis_identify_state": "off", "unexpected": "x",
	}

	got := summarizeIPMIChassis(raw)

	want := map[string]interface{}{
		"system_power": "on", "power_overload": "false", "cooling/fan_fault": "false", "chassis_identify_state": "off",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeIPMIChassis =\n%#v\nwant\n%#v", got, want)
	}
}
