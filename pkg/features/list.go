package features

//lint:file-ignore U1000 we want to introduce this feature flag unused.

var (
	// EventTimelineClusteredEventsUI enables the Event Timeline UI for Clustered Events.
	// NB: When removing this feature flag, remove references in ui/src/utils/featureFlags.js
	EventTimelineClusteredEventsUI = registerFeature("Enable Event Timeline Clustered Events UI", "ROX_EVENT_TIMELINE_CLUSTERED_EVENTS_UI", true)

	// AdmissionControlService enables running admission control as a separate microservice.
	AdmissionControlService = registerFeature("Separate admission control microservice", "ROX_ADMISSION_CONTROL_SERVICE", true)

	// AdmissionControlEnforceOnUpdate enables support for having the admission controller enforce on updates.
	AdmissionControlEnforceOnUpdate = registerFeature("Allow admission controller to enforce on update", "ROX_ADMISSION_CONTROL_ENFORCE_ON_UPDATE", true)

	// PolicyImportExport feature flag enables policy import and export
	PolicyImportExport = registerFeature("Enable Import/Export for Analyst Workflow", "ROX_POLICY_IMPORT_EXPORT", true)

	// AuthTestMode feature flag allows test mode flow for new auth provider in UI
	AuthTestMode = registerFeature("Enable Auth Test Mode UI", "ROX_AUTH_TEST_MODE_UI", true)

	// CurrentUserInfo enables showing information about the current user in UI
	CurrentUserInfo = registerFeature("Enable Current User Info UI", "ROX_CURRENT_USER_INFO", true)

	// ComplianceInNodes enables running of node-related Compliance checks in the compliance pods
	ComplianceInNodes = registerFeature("Enable compliance checks in nodes", "ROX_COMPLIANCE_IN_NODES", true)

	// RocksDB enables running of RocksDB
	RocksDB = registerFeature("Runs RocksDB instead of BadgerDB", "ROX_ROCKSDB", true)

	// csvExport enables CSV export of search results.
	csvExport = registerFeature("Enable CSV export of search results", "ROX_CSV_EXPORT", false)

	// ClusterHealthMonitoring enables monitoring of sensor and collector health
	ClusterHealthMonitoring = registerFeature("Enable cluster health monitoring", "ROX_CLUSTER_HEALTH_MONITORING", false)

	// SupportSlimCollectorMode enables support for retrieving slim Collector bundles from central.
	SupportSlimCollectorMode = registerFeature("Support slim Collector mode", "ROX_SUPPORT_SLIM_COLLECTOR_MODE", false)

	// AwsSecurityHubIntegration enables the AWS Security Hub Integration UI.
	AwsSecurityHubIntegration = registerFeature("Show AWS Security Hub Integration in UI", "ROX_AWS_SECURITY_HUB_INTEGRATION", false)
)
