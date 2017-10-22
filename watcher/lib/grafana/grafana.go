package grafana

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

// Client is the Grafana HTTP API client
type Client struct {
	*roundtrip.Client
}

// NewClient returns a Grafana HTTP API client
func NewClient() (*Client, error) {
	username := os.Getenv(constants.GrafanaUsernameEnv)
	if username == "" {
		return nil, trace.BadParameter("%s environment variable is not set", constants.GrafanaUsernameEnv)
	}

	password := os.Getenv(constants.GrafanaPasswordEnv)
	if password == "" {
		return nil, trace.BadParameter("%s environment variable is not set", constants.GrafanaPasswordEnv)
	}

	client, err := roundtrip.NewClient(constants.GrafanaAPIAddress, "", roundtrip.BasicAuth(username, password))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &Client{Client: client}, nil
}

// Health checks the status of Grafana HTTP API
func (c *Client) Health() error {
	// use "home dashboard" API as a health check
	_, err := c.Get(c.Endpoint("api", "dashboards", "home"), url.Values{})
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// CreateDashboard creates a new dashboard from the provided dashboard data
func (c *Client) CreateDashboard(data string) error {
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
