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

// Config is the configuration for InfluxDB
type Config struct {
	// InfluxDBAdminUser is the InfluxDB admin username
	InfluxDBAdminUser string
	// InfluxDBAdminPassword is the InfluxDB admin password
	InfluxDBAdminPassword string
	// InfluxDBGrafanaUser is the InfluxDB grafana username
	InfluxDBGrafanaUser string
	// InfluxDBGrafanaPassword is the InfluxDB grafana password
	InfluxDBGrafanaPassword string
	// InfluxDBTelegrafUser is the InfluxDB telegraf username
	InfluxDBTelegrafUser string
	// InfluxDBTelegrafPassword is the InfluxDB telegraf password
	InfluxDBTelegrafPassword string
}

// Client is the InfluxDB API client
type Client struct {
	client client_v2.Client
}

// NewClient creates a new InfluxDB client
func NewClient(config Config) (*Client, error) {
	log.Info("Initializing client.")
	client, err := client_v2.NewHTTPClient(client_v2.HTTPConfig{
		Addr:     constants.InfluxDBAPIAddress,
		Username: config.InfluxDBAdminUser,
		Password: config.InfluxDBAdminPassword,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// check authentication
	err = authenticateClient(client)
	if err == nil {
		return &Client{client: client}, nil
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

// Setup populates empty InfluxDB database with default users and retention policies.
// If retention policy exists then creation will be skipped.
func (c *Client) Setup(config Config) error {
	// create database for storing metrics from cluster
	log.Infof("Creating database %s", constants.InfluxDBDatabase)
	if err := c.execQuery(fmt.Sprintf(createDatabaseQuery, constants.InfluxDBDatabase)); err != nil {
		return trace.Wrap(err, "failed to create database %s", constants.InfluxDBDatabase)
	}

	var users = map[string]string{
		config.InfluxDBGrafanaUser:  config.InfluxDBGrafanaPassword,
		config.InfluxDBTelegrafUser: config.InfluxDBTelegrafPassword,
	}

	for user, password := range users {
		if err := c.UpsertUser(user, password); err != nil {
			return trace.Wrap(err)
		}
	}

	var privileges = map[string]string{
		config.InfluxDBGrafanaUser:  "read",
		config.InfluxDBTelegrafUser: "write",
	}
	for user, grants := range privileges {
		if err := c.GrantUserPrivileges(user, grants); err != nil {
			return trace.Wrap(err)
		}
	}

	if err := c.setupRetentionPolicies(); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpsertUser creates user if does not exist in the database and updates its password if different
func (c *Client) UpsertUser(user, password string) error {
	query := fmt.Sprintf(createUserQuery, user, password)
	log.Infof("Creating user %s.", user)

	err := c.execQuery(query)
	if err != nil {
		if trace.IsAlreadyExists(ConvertError(err)) {
			log.Infof("User %s already exists with different password. Updating password...", user)
			if err = c.execQuery(fmt.Sprintf(updatePasswordQuery, user, password)); err != nil {
				return trace.Wrap(err, "failed to update password for user %s", user)
			}

		}
		return trace.Wrap(err, "failed to create user %s", user)
	}
	return nil
}

// GrantUserPrivileges sets user privileges in the database
func (c *Client) GrantUserPrivileges(user, privileges string) error {
	query := fmt.Sprintf(grantPrivilegesQuery, privileges, constants.InfluxDBDatabase, user)
	if err := c.execQuery(query); err != nil {
		return trace.Wrap(err, "failed to grant privileges for user %s on database %s", user, constants.InfluxDBDatabase)
	}
	return nil
}

// CreateRollup creates a new rollup query in the database
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

// UpdateRollup updates an existing rollup query or creates a new one in the database.
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

func (c *Client) setupRetentionPolicies() error {
	queries := []string{
		fmt.Sprintf(createRetentionPolicyQuery, constants.InfluxDBRetentionPolicy,
			constants.InfluxDBDatabase, constants.DurationDefault) + " default",
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionMedium, constants.InfluxDBDatabase,
			constants.DurationMedium),
		fmt.Sprintf(createRetentionPolicyQuery, constants.RetentionLong, constants.InfluxDBDatabase,
			constants.DurationLong),
	}
	for _, query := range queries {
		log.WithField("query", query).Info("Setup retention policies query.")

		err := c.execQuery(query)
		if err != nil {
			if trace.IsAlreadyExists(ConvertError(err)) {
				log.Info("Retention policy already exists with different attributes. Skipping it.")
				continue
			}
			return trace.Wrap(err)
		}
	}
	return nil
}

// ConvertError converts error from InfluxDB query results
func ConvertError(err error) error {
	if strings.Contains(err.Error(), "already exists") {
		return trace.AlreadyExists(err.Error())
	}
	if strings.Contains(err.Error(), "authorization failed") {
		return trace.AccessDenied(err.Error())
	}
	if strings.Contains(err.Error(), "create admin user first or disable authentication") {
		return trace.NotFound(err.Error())
	}
	return err
}

func authenticateClient(client client_v2.Client) error {
	response, err := client.Query(client_v2.NewQuery(checkAuthenticationQuery, "", ""))
	if err != nil {
		return trace.Wrap(err)
	}
	if response.Error() != nil {
		return trace.Wrap(ConvertError(response.Error()))
	}

	return nil
}

func backwardsCompatibleClient(config Config) (*client_v2.Client, error) {
	client, err := client_v2.NewHTTPClient(client_v2.HTTPConfig{
		Addr:     constants.InfluxDBAPIAddress,
		Username: constants.InfluxDBAdminUser,
		Password: constants.InfluxDBAdminPassword,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	response, err := client.Query(client_v2.NewQuery(checkAuthenticationQuery, "", ""))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if response.Error() != nil {
		log.Errorf("Failed to authorize with user %v and default user. Probably password was changed in influxdb secret.", config.InfluxDBAdminUser)
		return nil, trace.Wrap(response.Error())
	}

	return &client, nil
}

func createAdminUser(client client_v2.Client, config Config) error {
	log.Infof("Creating admin user %v", config.InfluxDBAdminUser)
	query := fmt.Sprintf(createAdminQuery, config.InfluxDBAdminUser, config.InfluxDBAdminPassword)
	response, err := client.Query(client_v2.NewQuery(query, "", ""))
	if err != nil {
		return trace.Wrap(err, "failed to create admin user")
	}
	if response.Error() != nil {
		return trace.Wrap(response.Error(), "failed to create admin user")
	}
	return nil
}

const (
	// createAdminQuery is the InfluxDB query to create admin user
	createAdminQuery = "create user %v with password '%v' with all privileges"
	// updatePasswordQuery is the InfluxDB query to update password for the user
	updatePasswordQuery = "set password for %v = '%v'"
	// createUserQuery is the InfluxDB query to create a non-privileged user
	createUserQuery = "create user %v with password '%v'"
	// grantPrivilegesQuery is the InfluxDB query to grant privileges on a database to a user
	grantPrivilegesQuery = "grant %v on %q to %v"
	// createDatabaseQuery is the InfluxDB query to create a database
	createDatabaseQuery = "create database %q"
	// createRetentionPolicyQuery is the InfluxDB query to create a retention policy
	createRetentionPolicyQuery = "create retention policy %q on %q duration %v replication 1"
	// checkAuthenticationQuery is the simple query to check that authentication was successful
	checkAuthenticationQuery = "show databases"
)
