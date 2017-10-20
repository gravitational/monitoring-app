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

	"github.com/gravitational/trace"
	client "github.com/influxdata/kapacitor/client/v1"
	"github.com/influxdata/kapacitor/pipeline"
	"github.com/influxdata/kapacitor/tick"
	"github.com/influxdata/kapacitor/tick/stateful"
)

const (
	// APIAddress is the API adrress of Kapacitor running on the same pod
	APIAddress = "http://localhost:9092"
	// FetchRate is the rate Kapacitor Client will consume responses
	FetchRate = 100
	// Database is the InfluxDB database from where data is streamed
	Database = "k8s"
	// RetentionPolicy is the InfluxDB retention policy
	RetentionPolicy = "default"
	// UsernameEnv is the name of environment variable with Kapacitor username
	UsernameEnv = "KAPACITOR_USERNAME"
	// PasswordEnv is the name of environment variable with Kapacitor password
	PasswordEnv = "KAPACITOR_PASSWORD"
)

// ClientInterface represents a connection to a kapacitor instance
type ClientInterface interface {
	CreateTask(opt client.CreateTaskOptions) (client.Task, error)
	Ping() (time.Duration, string, error)
}

// NewClientInterface creates a Kapacitor client connection
func NewClientInterface(url, username, password string) (ClientInterface, error) {
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

	return &PaginatingClientInterface{clt, FetchRate}, nil
}

// ensure PaginatingClientInterface is a ClientInterface
var _ ClientInterface = &PaginatingClientInterface{}

// PaginatingClientInterface is a Kapacitor client that automatically prefetches
// data from kapacitor FetchRate elements at a time
type PaginatingClientInterface struct {
	ClientInterface
	FetchRate int // specifies the number of elements to fetch from Kapacitor at a time
}

// Client communicates to Kapacitor
type Client struct {
	ClientInterface
}

// NewClient creates a client that interfaces with Kapacitor tasks
func NewClient() (*Client, error) {
	username := os.Getenv(UsernameEnv)
	password := os.Getenv(PasswordEnv)

	kapaClient, err := NewClientInterface(APIAddress, username, password)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Client{
		kapaClient,
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

// CreateAlert creates alert task in Kapacitor
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

	dbrps := []client.DBRP{client.DBRP{
		Database:        Database,
		RetentionPolicy: RetentionPolicy,
	}}

	opts := client.CreateTaskOptions{
		ID:         name,
		Type:       taskType,
		DBRPs:      dbrps,
		TICKscript: script,
		Status:     client.Enabled,
	}

	if _, err := k.CreateTask(opts); err != nil {
		return trace.Wrap(err)
	}
	return nil
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

type deadman struct {
	interval  time.Duration
	threshold float64
	id        string
	message   string
	global    bool
}

func (d deadman) Interval() time.Duration { return d.interval }
func (d deadman) Threshold() float64      { return d.threshold }
func (d deadman) Id() string              { return d.id }
func (d deadman) Message() string         { return d.message }
func (d deadman) Global() bool            { return d.global }
