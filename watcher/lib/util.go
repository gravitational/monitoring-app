package lib

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/trace"
)

func OneOf(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

type APIClient interface {
	Health() error
}

// WaitForAPI spins until the API can be reached successfully or the provided context is cancelled
func WaitForAPI(ctx context.Context, client APIClient) error {
	for {
		select {
		case <-time.After(PollInterval):
			err := client.Health()
			if err != nil {
				log.Infof("API is not ready: %v", trace.DebugReport(err))
			} else {
				return nil
			}
		case <-ctx.Done():
			return trace.Errorf("API is not ready")
		}
	}
}
