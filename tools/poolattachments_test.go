package tools

import (
	"reflect"
	"testing"
)

func TestSummarizePoolAttachmentsKeepsTypeServiceAndSortedAttachments(t *testing.T) {
	raw := []map[string]interface{}{
		{"type": "NFS Share", "service": "nfs", "attachments": []interface{}{"/mnt/tank/b", "/mnt/tank/a"}},
		{"type": "Kubernetes", "service": nil, "attachments": []interface{}{}},
		{"type": "Virtual Machine"},
	}

	got := summarizePoolAttachments(raw)

	want := []map[string]interface{}{
		{"type": "Kubernetes", "service": nil, "attachments": []string{}},
		{"type": "NFS Share", "service": "nfs", "attachments": []string{"/mnt/tank/a", "/mnt/tank/b"}},
		{"type": "Virtual Machine", "service": nil, "attachments": []string{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizePoolAttachments =\n%#v\nwant\n%#v", got, want)
	}
}
