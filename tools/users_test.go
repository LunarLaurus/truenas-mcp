package tools

import (
	"reflect"
	"testing"
)

func TestSummarizeUsersKeepsAllowlistedFieldsOnly(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"username": "root", "uid": float64(0), "builtin": true, "locked": false, "smb": false,
			"ssh_password_enabled": true, "sudo_commands": []interface{}{}, "sudo_commands_nopasswd": []interface{}{"ALL"},
			"unixhash": "$6$secret", "smbhash": "abc", "sshpubkey": "ssh-ed25519 AAAA", "password": "x",
		},
		{"username": "alice", "uid": float64(1000), "locked": true, "smb": true},
	}

	got := summarizeUsers(raw)

	want := []map[string]interface{}{
		{"username": "alice", "uid": float64(1000), "builtin": false, "locked": true, "smb": true, "ssh_password_enabled": false, "has_sudo": false},
		{"username": "root", "uid": float64(0), "builtin": true, "locked": false, "smb": false, "ssh_password_enabled": true, "has_sudo": true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summarizeUsers =\n%#v\nwant\n%#v", got, want)
	}
}
