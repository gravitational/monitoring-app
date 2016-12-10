package lib

const (
	// GrafanaAPIAddress is the address of Grafana running in the same pod
	GrafanaAPIAddress = "http://localhost:3000"

	// GrafanaUsernameEnv is the name of environment variable with Grafana username
	GrafanaUsernameEnv = "GRAFANA_USERNAME"

	// GrafanaPasswordEnv is the name of environment variable with Grafana password
	GrafanaPasswordEnv = "GRAFANA_PASSWORD"

	// DashboardPrefix is the prefix of configmaps with dashboards data
	DashboardPrefix = "dashboard-"
)
