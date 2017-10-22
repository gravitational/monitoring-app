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

package kapacitor

import (
	"os"
	"strings"
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/gravitational/trace"
	client "github.com/influxdata/kapacitor/client/v1"
	"github.com/influxdata/kapacitor/pipeline"
	"github.com/influxdata/kapacitor/tick"
	"github.com/influxdata/kapacitor/tick/stateful"
)

// Client communicates to Kapacitor
type Client struct {
	clientInterface
}

// NewClient creates a client that interfaces with Kapacitor tasks
func NewClient() (*Client, error) {
	username := os.Getenv(constants.KapacitorUsernameEnv)
	password := os.Getenv(constants.KapacitorPasswordEnv)

	client, err := newClientInterface(constants.KapacitorAPIAddress, username, password)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Client{
		client,
	}, nil
}

// Health checks the status of Kapacitor HTTP API
func (k *Client) Health() error {
	_, _, err := k.Ping()
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// CreateAlert creates an alert task in Kapacitor
func (k *Client) CreateAlert(name string, script string) error {
	taskType := client.StreamTask
	taskTypePipeline := pipeline.StreamEdge
	if strings.Contains(name, "batch") {
		taskType = client.BatchTask
		taskTypePipeline = pipeline.BatchEdge
	}

	if err := validateTick(script, taskTypePipeline); err != nil {
		return trace.Wrap(err)
	}

	policies := []client.DBRP{client.DBRP{
		Database:        constants.Database,
		RetentionPolicy: constants.RetentionPolicy,
	}}

	opts := client.CreateTaskOptions{
		ID:         name,
		Type:       taskType,
		DBRPs:      policies,
		TICKscript: script,
		Status:     client.Enabled,
	}

	if _, err := k.CreateTask(opts); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpdateSMTPConfig updates Kapacitor SMTP configuration
func (k *Client) UpdateSMTPConfig(host string, port int, username, password string) error {
	link := k.ConfigElementLink("smtp", "")
	updateAction := client.ConfigUpdateAction{
		Set: map[string]interface{}{
			"host":     host,
			"port":     port,
			"username": username,
			"password": password,
		},
	}
	return trace.Wrap(k.ConfigUpdate(link, updateAction))
}

// UpdateAlertTarget updates Kapacitor alert target configuration
func (k *Client) UpdateAlertTarget(email string) error {
	link := k.ConfigElementLink("smtp", "")
	updateAction := client.ConfigUpdateAction{
		Set: map[string]interface{}{
			"to": []string{email},
		},
	}
	return trace.Wrap(k.ConfigUpdate(link, updateAction))
}

func validateTick(script string, tickType pipeline.EdgeType) error {
	scope := stateful.NewScope()
	predefinedVars := map[string]tick.Var{}
	_, err := pipeline.CreatePipeline(script, tickType, scope, &deadman{}, predefinedVars)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// deadman is an empty implementation of a kapacitor DeadmanService to allow CreatePipeline
var _ pipeline.DeadmanService = &deadman{}

type deadman struct{}

func (d deadman) Interval() time.Duration { return 0 }
func (d deadman) Threshold() float64      { return 0 }
func (d deadman) Id() string              { return "" }
func (d deadman) Message() string         { return "" }
func (d deadman) Global() bool            { return false }

// clientInterface represents a connection to a Kapacitor instance
type clientInterface interface {
	CreateTask(opt client.CreateTaskOptions) (client.Task, error)
	ConfigUpdate(link client.Link, action client.ConfigUpdateAction) error
	ConfigElementLink(section, element string) client.Link
	Ping() (time.Duration, string, error)
}

// newClientInterface creates a Kapacitor client connection
func newClientInterface(url, username, password string) (clientInterface, error) {
	var creds *client.Credentials
	if username != "" && password != "" {
		creds = &client.Credentials{
			Method:   client.UserAuthentication,
			Username: username,
			Password: password,
		}
	}
	clt, err := client.New(client.Config{
		URL:         url,
		Credentials: creds,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &paginatingClient{clt, constants.KapacitorFetchRate}, nil
}

// ensure paginatingClient is a clientInterface
var _ clientInterface = &paginatingClient{}

// paginatingClient is a Kapacitor client that automatically prefetches
// data from kapacitor FetchRate elements at a time
type paginatingClient struct {
	*client.Client
	FetchRate int // specifies the number of elements to fetch from Kapacitor at a time
}
