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
	"strings"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/gravitational/trace"
	client_v2 "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

// Client is the InfluxDB API client
type Client struct {
	client client_v2.Client
}

// NewClient creates a new InfluxDB client
func NewClient(username, password string) (*Client, error) {
	client, err := client_v2.NewHTTPClient(client_v2.HTTPConfig{
		Addr:     constants.InfluxDBAPIAddress,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &Client{client: client}, nil
}

// Health checks the API readiness
func (c *Client) Health() error {
	const noWait = 0 // do not need to wait for leader of InfluxDB cluster
	_, _, err := c.client.Ping(noWait)
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
		fmt.Sprintf(createDatabaseQuery, constants.InfluxDBDatabase),
		fmt.Sprintf(grantReadQuery, constants.InfluxDBDatabase, constants.InfluxDBGrafanaUser),
		fmt.Sprintf(createRetentionPolicyQuery, constants.InfluxDBRetentionPolicy,
			constants.InfluxDBDatabase, constants.DurationDefault) + " default",
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionMedium, constants.InfluxDBDatabase,
			constants.DurationMedium),
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionLong, constants.InfluxDBDatabase,
			constants.DurationLong),
	}
	for _, query := range queries {
		log.WithField("query", query).Info("Setup query.")

		if err := c.execQuery(query); err != nil {
			return trace.Wrap(err)
		}
	}
	return nil
}

// CreateRollup creates a rollup query in the database
func (c *Client) CreateRollup(r Rollup) error {
	err := r.Check()
	if err != nil {
		return trace.Wrap(err)
	}

	query, err := r.buildCreateQuery()
	if err != nil {
		return trace.Wrap(err)
	}
	log.WithField("query", query).Info("New rollup.")

	if err = c.execQuery(query); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// DeleteRollup deletes a rollup query from the database
func (c *Client) DeleteRollup(r Rollup) error {
	err := r.Check()
	if err != nil {
		return trace.Wrap(err)
	}

	query, err := r.buildDeleteQuery()
	if err != nil {
		return trace.Wrap(err)
	}
	log.WithField("query", query).Info("Remove rollup.")

	if err = c.execQuery(query); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpdateRollup updates a rollup query in the database
func (c *Client) UpdateRollup(r Rollup) error {
	err := r.Check()
	if err != nil {
		return trace.Wrap(err)
	}

	deleteQuery, err := r.buildDeleteQuery()
	if err != nil {
		return trace.Wrap(err)
	}
	createQuery, err := r.buildCreateQuery()
	if err != nil {
		return trace.Wrap(err)
	}
	query := strings.Join([]string{deleteQuery, createQuery}, "; ")
	log.WithField("query", query).Info("Update rollup.")

	if err = c.execQuery(query); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (c *Client) execQuery(query string) error {
	response, err := c.client.Query(client_v2.NewQuery(query, "", ""))
	if err != nil {
		return trace.Wrap(err)
	}
	if response.Error() != nil {
		return trace.Wrap(response.Error())
	}

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
	createDatabaseQuery = "create database %q"
	// createRetentionPolicyQuery is the InfluxDB query to create a retention policy
	createRetentionPolicyQuery = "create retention policy %q on %q duration %v replication 1"
)
