package tools

import "fmt"

// mutatingTools names every tool that changes state on the TrueNAS system.
// Read-only mode never exposes these. Every registered tool must be in exactly one of
// mutatingTools and readOnlyTools (TestEveryRegisteredToolIsClassifiedExactlyOnce), and every
// write API method must have an owning tool listed here (TestEveryWriteMethodHasAMutatingOwner).
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

// readOnlyTools is the allowlist exposed in read-only mode. It is default-deny: a tool that is not
// listed here is hidden and refused when the server runs with --read-only, even if nobody remembered
// to classify it as mutating.
var readOnlyTools = map[string]bool{
	"analyze_capacity":             true,
	"check_updates":                true,
	"get_app_catalog_details":      true,
	"get_app_config":               true,
	"get_arc_metrics":              true,
	"get_current_boot_environment": true,
	"get_dataset_quota":            true,
	"get_directory_service_status": true,
	"get_disk_metrics":             true,
	"get_disk_temperature_alerts":  true,
	"get_disk_temperatures":        true,
	"get_disk_usage":               true,
	"get_ipmi_chassis":             true,
	"get_iscsi_sessions":           true,
	"get_network_metrics":          true,
	"get_nfs_status":               true,
	"get_pool_attachments":         true,
	"get_pool_capacity_details":    true,
	"get_pool_health":              true,
	"get_pool_processes":           true,
	"get_scrub_status":             true,
	"get_system_metrics":           true,
	"get_ups_metrics":              true,
	"list_alerts":                  true,
	"list_directory_certificates":  true,
	"query_apps":                   true,
	"query_audit":                  true,
	"query_boot_environments":      true,
	"query_certificates":           true,
	"query_datasets":               true,
	"query_directory_services":     true,
	"query_disks":                  true,
	"query_interfaces":             true,
	"query_iscsi":                  true,
	"query_jobs":                   true,
	"query_pools":                  true,
	"query_reporting_exporters":    true,
	"query_scheduled_jobs":         true,
	"query_scrub_schedules":        true,
	"query_services":               true,
	"query_shares":                 true,
	"query_snapshot_tasks":         true,
	"query_snapshots":              true,
	"query_users":                  true,
	"query_vms":                    true,
	"query_zfs_resources":          true,
	"search_app_catalog":           true,
	"system_health":                true,
	"system_info":                  true,
	"tasks_get":                    true,
	"tasks_list":                   true,
	"update_status":                true,
}

// SetReadOnly toggles read-only mode: only tools in readOnlyTools are exposed; everything else is hidden and refused.
func (r *Registry) SetReadOnly(readOnly bool) {
	r.readOnly = readOnly
}

// refuseIfReadOnly returns an error when read-only mode is on and name is not an allowlisted read-only tool.
func (r *Registry) refuseIfReadOnly(name string) error {
	if r.readOnly && !readOnlyTools[name] {
		return fmt.Errorf("tool %q is disabled: the server is running in read-only mode", name)
	}
	return nil
}
