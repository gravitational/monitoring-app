package lib

import (
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
)

// InfluxdbClient is Influxdb API client
type InfluxdbClient struct {
	*roundtrip.Client
}

// NewInfluxdbClient creates a new client
func NewInfluxdbClient() (*InfluxdbClient, error) {
	client, err := roundtrip.NewClient(InfluxdbAPIAddress, "")
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &InfluxdbClient{Client: client}, nil
}

// Health checks the API readiness by querying for the default database
func (c *InfluxdbClient) Health() error {
	response, err := c.Get(c.Endpoint("query"), url.Values{"q": []string{"show databases"}})
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v %v %v", response.Code(), response.Headers(), string(response.Bytes()))
	if !strings.Contains(string(response.Bytes()), InfluxdbDatabase) {
		return trace.NotFound("database %v not found", InfluxdbDatabase)
	}

	return nil
}

// CreateRollup creates a new rollup query in the database
func (c *InfluxdbClient) CreateRollup(r Rollup) error {
	err := r.Check()
	if err != nil {
		return trace.Wrap(err)
	}

	query, err := buildQuery(r)
	if err != nil {
		return trace.Wrap(err)
	}
	log.Infof("%v", query)

	response, err := c.PostForm(c.Endpoint("query"), url.Values{
		"q": []string{query},
	})
	if err != nil {
		return trace.Wrap(err)
	}

	log.Infof("%v %v %v", response.Code(), response.Headers(), string(response.Bytes()))
	return nil
}
