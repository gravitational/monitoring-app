package lib

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
)

// InfluxDBClient is InfluxDB API client
type InfluxDBClient struct {
	*roundtrip.Client
}

// NewInfluxDBClient creates a new client
func NewInfluxDBClient() (*InfluxDBClient, error) {
	client, err := roundtrip.NewClient(InfluxDBAPIAddress, "",
		roundtrip.BasicAuth(InfluxDBAdminUser, InfluxDBAdminPassword))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &InfluxDBClient{Client: client}, nil
}

// Health checks the API readiness
func (c *InfluxDBClient) Health() error {
	_, err := c.Get(c.Endpoint("ping"), url.Values{})
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// Setup sets up InfluxDB database
func (c *InfluxDBClient) Setup() error {
	queries := []string{
		fmt.Sprintf(createAdminQuery, InfluxDBAdminUser, InfluxDBAdminPassword),
		fmt.Sprintf(createUserQuery, InfluxDBGrafanaUser, InfluxDBGrafanaPassword),
		fmt.Sprintf(createDatabaseQuery, InfluxDBDatabase, DurationDefault),
		fmt.Sprintf(grantReadQuery, InfluxDBDatabase, InfluxDBGrafanaUser),
		fmt.Sprintf(createRetentionPolicyQuery, InfluxDBRetentionPolicy, InfluxDBDatabase, DurationDefault) + " default",
		fmt.Sprintf(createRetentionPolicyQuery, RetentionMedium, InfluxDBDatabase, DurationMedium),
		fmt.Sprintf(createRetentionPolicyQuery, RetentionLong, InfluxDBDatabase, DurationLong),
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
func (c *InfluxDBClient) CreateRollup(r Rollup) error {
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

var (
	createAdminQuery = "create user %v with password '%v' with all privileges"
	createUserQuery  = "create user %v with password '%v'"
	grantReadQuery   = "grant read on %q to %v"
	// createDatabaseQuery is the InfluxDB query to create a database
	createDatabaseQuery = "create database %q with duration %v"
	// createRetentionPolicyQuery is the InfluxDB query to create a retention policy
	createRetentionPolicyQuery = "create retention policy %q on %q duration %v replication 1"
)
