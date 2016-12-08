package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
)

type GrafanaClient struct {
	*roundtrip.Client
}

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

func (c *GrafanaClient) CreateDashboard(data string) error {
	var dashboardJSON map[string]interface{}
	if err := json.Unmarshal([]byte(data), &dashboardJSON); err != nil {
		return trace.Wrap(err)
	}

	response, err := c.PostJSON(c.Endpoint("api", "dashboards", "db"), CreateDashboardRequest{
		Dashboard: dashboardJSON,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v", response)
	return nil
}

func (c *GrafanaClient) Endpoint(params ...string) string {
	return fmt.Sprintf("%s/%s", GrafanaAPIAddress, strings.Join(params, "/"))
}

type CreateDashboardRequest struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	Overwrite bool                   `json:"overwrite"`
}
