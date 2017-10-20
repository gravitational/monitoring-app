package lib

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/gravitational/trace"
)

// OneOf returns true if the value is present in the list of values
func OneOf(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// APIClient defines generic interface for an API client
type APIClient interface {
	// Health checks the API readiness
	Health() error
}

// WaitForAPI spins until the API can be reached successfully or the provided context is cancelled
func WaitForAPI(ctx context.Context, client APIClient) (err error) {
	for {
		select {
		case <-time.After(PollInterval):
			err = client.Health()
			if err != nil {
				log.Infof("API is not ready: %v", trace.DebugReport(err))
			} else {
				return nil
			}
		case <-ctx.Done():
			return trace.ConnectionProblem(err, "API is not ready")
		}
	}
}
