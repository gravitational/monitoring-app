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
	// MonitoringNamespace is the name of k8s namespace where all our monitoring stuff goes
	MonitoringNamespace = "monitoring"

	// ModeDashboards is the mode in which watcher polls for new dashboards
	ModeDashboards = "dashboards"

	// ModeAlerts is the mode in which watcher polls for new alerts and
	// monitoring configuration updates
	ModeAlerts = "alerts"

	// ModeAutoscale is the mode in which watcher updates the number of
	// Prometheus/Alertmanager replicas based on the number of nodes.
	ModeAutoscale = "autoscale"

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

	// MonitoringLabel is the label for resources with configuration updates
	MonitoringLabel = "monitoring"
	// MonitoringUpdateAlert defines the update for an alert
	MonitoringUpdateAlert = "alert"
	// MonitoringUpdateAlertTarget defines the update for an alert target
	MonitoringUpdateAlertTarget = "alert-target"
	// MonitoringUpdateDashboard defines the update for a dashboard
	MonitoringUpdateDashboard = "dashboard"
	// MonitoringUpdateSMTP defines the update for kapacitor SMTP configuration
	MonitoringUpdateSMTP = "smtp"

	// ResourceSpecKey specifies the name of the key with raw resource specification
	ResourceSpecKey = "spec"

	// AlertFrom specifies default sender's email for alert email notifications
	AlertFrom = "noreply@gravitational.com"

	// SMTPSecret specifies the name of the SMTP configuration secret
	SMTPSecret = "smtp-configuration"

	// AlertTargetConfigMap specifies the name of the alert target configmap
	AlertTargetConfigMap = "alerting-addresses"

	// AppLabel specifies the name of the label to define the type of an application
	AppLabel = "app"
	// ComponentLabel specifies the name of the label to define a sub-component
	ComponentLabel = "component"

	// MonitoringApp defines the monitoring application label
	MonitoringApp = "monitoring"

	// NodeRoleLabel is the label with Kubernetes node role.
	NodeRoleLabel = "gravitational.io/k8s-role"
	// MasterLabel is the label that marks Kubernetes master nodes.
	MasterLabel = "master"

	// AlermanagerName is the name of the Alertmanager CRD object.
	AlertmanagerName = "main"
	// PrometheusName is the name of the Prometheus CRD object.
	PrometheusName = "k8s"
)

var (
	// AllModes contains names of all modes the watcher can run in
	AllModes = []string{
		ModeAlerts,
		ModeDashboards,
		ModeAutoscale,
	}
)
