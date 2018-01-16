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

package grafana

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"

	"github.com/gosimple/slug"
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

// DeleteDashboard deletes dashboard
// dashboard name is obtained from dashboard data field 'title'
func (c *Client) DeleteDashboard(data string) error {
	// dashboard data should be a valid JSON
	var dashboardJSON map[string]interface{}
	if err := json.Unmarshal([]byte(data), &dashboardJSON); err != nil {
		return trace.Wrap(err)
	}

	response, err := c.Delete(c.Endpoint("api", "dashboards", "db", slug.Make(strings.ToLower(dashboardJSON["title"].(string)))))
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v", response)
	return nil
}
