package tools

import "fmt"

// mutatingTools names every tool that changes state on the TrueNAS system.
// SetReadOnly(true) hides these from ListTools and refuses them in CallTool.
// A tool that is not listed here is treated as read-only, so a new mutating
// tool must be added (TestMutatingVerbsAreAllClassified catches the usual verbs).
var mutatingTools = map[string]bool{
	"download_update":             true,
	"apply_update":                true,
	"system_reboot":               true,
	"delete_boot_environment":     true,
	"create_scrub_schedule":       true,
	"run_scrub":                   true,
	"delete_scrub_schedule":       true,
	"refresh_directory_cache":     true,
	"configure_directory_service": true,
	"leave_directory_service":     true,
	"create_dataset":              true,
	"create_smb_share":            true,
	"create_nfs_share":            true,
	"dismiss_alert":               true,
	"restore_alert":               true,
	"upgrade_app":                 true,
	"start_app":                   true,
	"stop_app":                    true,
	"update_app":                  true,
	"install_app":                 true,
	"delete_app":                  true,
}

// SetReadOnly toggles read-only mode: mutating tools are hidden and refused.
func (r *Registry) SetReadOnly(readOnly bool) {
	r.readOnly = readOnly
}

// refuseIfReadOnly returns an error when name is a mutating tool and read-only mode is on.
func (r *Registry) refuseIfReadOnly(name string) error {
	if r.readOnly && mutatingTools[name] {
		return fmt.Errorf("tool %q is disabled: the server is running in read-only mode", name)
	}
	return nil
}
