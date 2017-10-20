package lib

import "time"

const (
	// ModeDashboards is the mode in which watcher polls for new dashboards
	ModeDashboards = "dashboards"

	// ModeRollups is the mode in which watcher polls for new rollups
	ModeRollups = "rollups"

	// ModeAlerts is the mode in which watcher polls for new alerts and
	// monitoring configuration updates
	ModeAlerts = "alerts"

	// GrafanaAPIAddress is the API address of Grafana running in the same pod
	GrafanaAPIAddress = "http://localhost:3000"

	// GrafanaUsernameEnv is the name of environment variable with Grafana username
	GrafanaUsernameEnv = "GRAFANA_USERNAME"

	// GrafanaPasswordEnv is the name of environment variable with Grafana password
	GrafanaPasswordEnv = "GRAFANA_PASSWORD"

	// DashboardPrefix is the prefix of configmaps with dashboards data
	DashboardPrefix = "dashboard-"

	// PollInterval is interval between attempts to reach API
	PollInterval = 2 * time.Second

	// InfluxDBAPIAddress if the API address of InfluxDB running in the same pod
	InfluxDBAPIAddress = "http://localhost:8086"

	// InfluxDBDatabase is the name of the database where all metrics go
	InfluxDBDatabase = "k8s"
	// InfluxDBAdminUser is the InfluxDB admin username
	InfluxDBAdminUser = "root" // "root" for backward compatibility
	// InfluxDBAdminPassword is the InfluxDB admin password
	InfluxDBAdminPassword = "root" // "root" for backward compatibility
	// InfluxDBGrafanaUser is the InfluxDB user for Grafana
	InfluxDBGrafanaUser = "grafana"
	// InfluxDBGrafanaPassword is the InfluxDB password for Grafana
	InfluxDBGrafanaPassword = "grafana"

	// InfluxDBRetentionPolicy is the name of the default retention policy
	InfluxDBRetentionPolicy = "default"

	// RollupsPrefix is the prefix of configmaps with rollups
	RollupsPrefix = "rollups-"

	// RetentionLong is the name of the "long" retention policy
	RetentionLong = "long"
	// RetentionMedium is the name of the "medium" retention policy
	RetentionMedium = "medium"

	// DurationDefault is the duration of "default" retention policy in format InfluxDB expects
	DurationDefault = "24h"
	// DurationMedium is the duration of "medium" retention policy in format InfluxDB expects
	DurationMedium = "4w"
	// DurationLong is the duration of "long" retention policy in format InfluxDB expects
	DurationLong = "52w"

	// IntervalMedium is the aggregation interval for "medium" retention policy
	IntervalMedium = "5m"
	// IntervalLong is the aggregation interval for "long" retention policy
	IntervalLong = "1h"

	// FunctionMean is the average function
	FunctionMean = "mean"
	// FunctionMedian is the median function
	FunctionMedian = "median"
	// FunctionSum is the sum function
	FunctionSum = "sum"
	// FunctionMax is the max function
	FunctionMax = "max"
	// FunctionMin is the min function
	FunctionMin = "min"
	// FunctionPercentile is the percentile function
	FunctionPercentile = "percentile"

	// MonitoringLabel is the label for resources with configuration updates
	MonitoringLabel = "monitoring"
	// MonitoringUpdateAlert defines the update for an alert
	MonitoringUpdateAlert = "alert"
	// MonitoringUpdateAlertTarget defines the update for an alert target
	MonitoringUpdateAlertTarget = "alerttarget"

	// ResourceSpecKey specifies the name of the key with raw resource specification
	ResourceSpecKey = "spec"

	// SmtpSecret specifies the name of the secret with SMTP configuration update
	SmtpSecret = "smtp-configuration-update"
	// AlertTargetConfigMap specifies the name of the configmap with alert target update
	AlertTargetConfigMap = "alert-target-update"

	// KapacitorAlertFrom specifies default sender's email for alert email notifications
	KapacitorAlertFrom = "noreply@gravitational.com"

	// KapacitorSMTPSecret specifies the name of the kapacitor's SMTP configuration secret
	KapacitorSMTPSecret = "smtp-configuration"

	// KapacitorAlertTargetConfigMap specifies the name of the kapacitor's alert target configmap
	KapacitorAlertTargetConfigMap = "alerting-addresses"

	// AppLabel specifies the name of the label to define the type of an application
	AppLabel = "app"
	// ComponentLabel specifies the name of the label to define a sub-component
	ComponentLabel = "component"

	// MonitoringApp defines the monitoring application label
	MonitoringApp = "monitoring"

	// ComponentKapacitor defines the Kapacitor monitoring application component
	ComponentKapacitor = "kapacitor"
)

var (
	// AllModes contains names of all modes the watcher can run in
	AllModes = []string{
		ModeAlerts,
		ModeDashboards,
		ModeRollups,
	}

	// AllRetentions contains names of all supported retention policies
	AllRetentions = []string{
		RetentionLong,
		RetentionMedium,
	}

	// RetentionToInterval maps the name of retention policy name to aggregation interval
	RetentionToInterval = map[string]string{
		RetentionLong:   IntervalLong,
		RetentionMedium: IntervalMedium,
	}

	// AllFunctions contains names of functions, excluding percentile (because percentile is
	// formatted like 'percentile_X')
	AllFunctions = []string{
		FunctionMean,
		FunctionMedian,
		FunctionSum,
		FunctionMax,
		FunctionMin,
	}
)
