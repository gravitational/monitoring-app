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

package influxdb

import (
	"fmt"
	"net/url"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

// Client is the InfluxDB API client
type Client struct {
	*roundtrip.Client
}

// NewClient creates a new InfluxDB client
func NewClient() (*Client, error) {
	client, err := roundtrip.NewClient(constants.InfluxDBAPIAddress, "",
		roundtrip.BasicAuth(constants.InfluxDBAdminUser, constants.InfluxDBAdminPassword))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &Client{Client: client}, nil
}

// Health checks the API readiness
func (c *Client) Health() error {
	_, err := c.Get(c.Endpoint("ping"), url.Values{})
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// Setup sets up InfluxDB database
func (c *Client) Setup() error {
	queries := []string{
		fmt.Sprintf(createAdminQuery, constants.InfluxDBAdminUser, constants.InfluxDBAdminPassword),
		fmt.Sprintf(createUserQuery, constants.InfluxDBGrafanaUser, constants.InfluxDBGrafanaPassword),
		fmt.Sprintf(createDatabaseQuery, constants.InfluxDBDatabase, constants.DurationDefault),
		fmt.Sprintf(grantReadQuery, constants.InfluxDBDatabase, constants.InfluxDBGrafanaUser),
		fmt.Sprintf(createRetentionPolicyQuery, constants.InfluxDBRetentionPolicy,
			constants.InfluxDBDatabase, constants.DurationDefault) + " default",
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionMedium, constants.InfluxDBDatabase,
			constants.DurationMedium),
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionLong, constants.InfluxDBDatabase,
			constants.DurationLong),
	}
	for _, query := range queries {
		log.Infof("%v", query)

		response, err := c.PostForm(c.Endpoint("query"), url.Values{"q": []string{query}})
		if err != nil {
			return trace.Wrap(err)
		}

		log.Infof("%v %v %v", response.Code(), response.Headers(), string(response.Bytes()))
	}
	return nil
}

// CreateRollup creates a new rollup query in the database
func (c *Client) CreateRollup(r Rollup) error {
	err := r.Check()
	if err != nil {
		return trace.Wrap(err)
	}

	query, err := buildQuery(r)
	if err != nil {
		return trace.Wrap(err)
	}
	log.Infof("%v", query)

	response, err := c.PostForm(c.Endpoint("query"), url.Values{"q": []string{query}})
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v %v %v", response.Code(), response.Headers(), string(response.Bytes()))
	return nil
}

const (
	// createAdminQuery is the InfluxDB query to create admin user
	createAdminQuery = "create user %v with password '%v' with all privileges"
	// createUserQuery is the InfluxDB query to create a non-privileged user
	createUserQuery = "create user %v with password '%v'"
	// grantReadQuery is the InfluxDB query to grant read privileges on a database to a user
	grantReadQuery = "grant read on %q to %v"
	// createDatabaseQuery is the InfluxDB query to create a database
	createDatabaseQuery = "create database %q with duration %v"
	// createRetentionPolicyQuery is the InfluxDB query to create a retention policy
	createRetentionPolicyQuery = "create retention policy %q on %q duration %v replication 1"
)
