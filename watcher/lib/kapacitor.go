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

package lib

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

// KapaClient represents a connection to a kapacitor instance
type KapaClient interface {
	CreateTask(opt client.CreateTaskOptions) (client.Task, error)
	Ping() (time.Duration, string, error)
}

// NewKapaClient creates a Kapacitor client connection
func NewKapaClient(url, username, password string) (KapaClient, error) {
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

	return &PaginatingKapaClient{clt, KapacitorFetchRate}, nil
}

// ensure PaginatingKapaClient is a KapaClient
var _ KapaClient = &PaginatingKapaClient{}

// PaginatingKapaClient is a Kapacitor client that automatically navigates
// through Kapacitor's pagination to fetch all results
type PaginatingKapaClient struct {
	KapaClient
	FetchRate int // specifies the number of elements to fetch from Kapacitor at a time
}

// KapacitorClient communicates to Kapacitor
type KapacitorClient struct {
	kapaClient KapaClient
}

// NewKapacitorClient creates a client that interfaces with Kapacitor tasks
func NewKapacitorClient() (*KapacitorClient, error) {
	username := os.Getenv(KapacitorUsernameEnv)
	password := os.Getenv(KapacitorPasswordEnv)

	kapaClient, err := NewKapaClient(KapacitorAPIAddress, username, password)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &KapacitorClient{
		kapaClient: kapaClient,
	}, nil
}

func (k *KapacitorClient) Health() error {
	_, _, err := k.kapaClient.Ping()
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (k *KapacitorClient) CreateAlert(name string, script string) error {
	taskType := client.StreamTask
	taskTypePipeline := pipeline.StreamEdge
	if strings.Contains(name, "batch") {
		taskType = client.BatchTask
		taskTypePipeline = pipeline.BatchEdge
	}

	if err := validateTick(script, taskTypePipeline); err != nil {
		return trace.Wrap(err)
	}

	dbrps := make([]client.DBRP, 1)
	dbrps[1] = client.DBRP{
		Database:        KapacitorDatabase,
		RetentionPolicy: KapacitorRetentionPolicy,
	}

	opts := client.CreateTaskOptions{
		ID:         name,
		Type:       taskType,
		DBRPs:      dbrps,
		TICKscript: script,
		Status:     client.Enabled,
	}

	if _, err := k.kapaClient.CreateTask(opts); err != nil {
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
