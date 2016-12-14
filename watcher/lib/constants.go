package lib

import "time"

const (
	// ModeDashboards is the mode in which watcher polls for new dashboards
	ModeDashboards = "dashboards"

	// ModeRollups is the mode in which watcher polls for new rollups
	ModeRollups = "rollups"

	// GrafanaAPIAddress is the API address of Grafana running in the same pod
	GrafanaAPIAddress = "http://localhost:3000"

	// GrafanaUsernameEnv is the name of environment variable with Grafana username
	GrafanaUsernameEnv = "GRAFANA_USERNAME"

	// GrafanaPasswordEnv is the name of environment variable with Grafana password
	GrafanaPasswordEnv = "GRAFANA_PASSWORD"

	// DashboardPrefix is the prefix of configmaps with dashboards data
	DashboardPrefix = "dashboard-"

	// pollInterval is interval between attempts to reach API
	PollInterval = 2 * time.Second

	// InfluxdbAPIAddress if the API address of Influxdb running in the same pod
	InfluxdbAPIAddress = "http://localhost:8086"

	// InfluxdbDatabase is the name of the database where all metrics go
	InfluxdbDatabase = "k8s"

	// InfluxdbRetentionPolicy is the name of the default retention policy
	InfluxdbRetentionPolicy = "default"

	// RollupsPrefix is the prefix of configmaps with rollups
	RollupsPrefix = "rollups-"

	// RetentionYear is the name of the "long" retention policy
	RetentionYear = "year"
	// RetentionMonth is the name of the "medium" retention policy
	RetentionMonth = "month"

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
)

var (
	// AllRetentions contains names of all supported retention policies
	AllRetentions = []string{
		RetentionYear,
		RetentionMonth,
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
