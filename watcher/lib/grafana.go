package lib

import (
	"encoding/json"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
)

// GrafanaClient is Grafana HTTP API client
type GrafanaClient struct {
	*roundtrip.Client
}

// NewGrafanaClient returns a Grafana HTTP API client
func NewGrafanaClient() (*GrafanaClient, error) {
	username := os.Getenv(GrafanaUsernameEnv)
	if username == "" {
		return nil, trace.BadParameter("%s environment variable is not set", GrafanaUsernameEnv)
	}

	password := os.Getenv(GrafanaPasswordEnv)
	if password == "" {
		return nil, trace.BadParameter("%s environment variable is not set", GrafanaPasswordEnv)
	}

	client, err := roundtrip.NewClient(GrafanaAPIAddress, "", roundtrip.BasicAuth(username, password))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &GrafanaClient{Client: client}, nil
}

// Health checks the status of Grafana HTTP API
func (c *GrafanaClient) Health() error {
	// use "home dashboard" API as a health check
	_, err := c.Get(c.Endpoint("api", "dashboards", "home"), url.Values{})
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// CreateDashboard creates a new dashboard from the provided dashboard data
func (c *GrafanaClient) CreateDashboard(data string) error {
	// dashboard data should be a valid JSON
	var dashboardJSON map[string]interface{}
	if err := json.Unmarshal([]byte(data), &dashboardJSON); err != nil {
		return trace.Wrap(err)
	}

	response, err := c.PostJSON(c.Endpoint("api", "dashboards", "db"), CreateDashboardRequest{
		Dashboard: dashboardJSON,
		Overwrite: true,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v", response)
	return nil
}

// CreateDashboardRequest is request to create a new dashboard
type CreateDashboardRequest struct {
	// Dashboard is the dashboard data
	Dashboard map[string]interface{} `json:"dashboard"`
	// Overwrite is whether to overwrite existing dashboard with newer version or with same dashboard title
	Overwrite bool `json:"overwrite"`
}
