/*
Copyright 2017 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constants

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

	// Aggregate Functions

	// FunctionCount is the count function
	FunctionCount = "count"
	// FunctionDistinct is the distinct function
	FunctionDistinct = "distinct"
	// FunctionIntegral is the integral function
	FunctionIntegral = "integral"
	// FunctionMean is the mean function
	FunctionMean = "mean"
	// FunctionMedian is the median function
	FunctionMedian = "median"
	// FunctionMode is the mode function
	FunctionMode = "mode"
	// FunctionSpread is the spread function
	FunctionSpread = "spread"
	// FunctionStdDev is the stddev function
	FunctionStdDev = "stddev"
	// FunctionSum is the sum function
	FunctionSum = "sum"

	// Selector Functions

	// FunctionBottom is the bottom function
	FunctionBottom = "bottom"
	// FunctionFirst is the first function
	FunctionFirst = "first"
	// FunctionLast is the last function
	FunctionLast = "last"
	// FunctionMax is the max function
	FunctionMax = "max"
	// FunctionMin is the min function
	FunctionMin = "min"
	// FunctionPercentile is the percentile function
	FunctionPercentile = "percentile"
	// FunctionSample is the sample function
	FunctionSample = "sample"
	// FunctionTop is the top function
	FunctionTop = "top"

	// MonitoringLabel is the label for resources with configuration updates
	MonitoringLabel = "monitoring"
	// MonitoringUpdateAlert defines the update for an alert
	MonitoringUpdateAlert = "alert"
	// MonitoringUpdateAlertTarget defines the update for an alert target
	MonitoringUpdateAlertTarget = "alert-target"
	// MonitoringUpdateDashboard defines the update for a dashboard
	MonitoringUpdateDashboard = "dashboard"
	// MonitoringUpdateRollup defines the update for a rollup
	MonitoringUpdateRollup = "rollup"
	// MonitoringUpdateSMTP defines the update for kapacitor SMTP configuration
	MonitoringUpdateSMTP = "smtp"

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

	// KapacitorAPIAddress is the API adrress of Kapacitor running on the same pod
	KapacitorAPIAddress = "http://localhost:9092"
	// KapacitorFetchRate is the rate Kapacitor Client will consume responses
	KapacitorFetchRate = 100
	// Database is the InfluxDB database from where data is streamed
	Database = "k8s"
	// RetentionPolicy is the InfluxDB retention policy
	RetentionPolicy = "default"
	// KapacitorUsernameEnv is the name of environment variable with Kapacitor username
	KapacitorUsernameEnv = "KAPACITOR_USERNAME"
	// KapacitorPasswordEnv is the name of environment variable with Kapacitor password
	KapacitorPasswordEnv = "KAPACITOR_PASSWORD"
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

	// CompositeFunctions lists those functions that need an additional
	// parameter, specified in the name and identified via the "_" separator
	// and thus formatted like 'percentile_99', 'top_10', 'bottom_3'
	CompositeFunctions = []string{
		FunctionPercentile,
		FunctionBottom,
		FunctionTop,
		FunctionSample,
	}
	// SimpleFunctions that don't need the composite structure explained above
	SimpleFunctions = []string{
		FunctionCount,
		FunctionDistinct,
		FunctionIntegral,
		FunctionMean,
		FunctionMedian,
		FunctionMode,
		FunctionSpread,
		FunctionStdDev,
		FunctionSum,
		FunctionFirst,
		FunctionLast,
		FunctionMax,
		FunctionMin,
	}
)
