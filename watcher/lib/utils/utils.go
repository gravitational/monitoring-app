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

package utils

import (
	"context"
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
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
	Health(context.Context) error
}

// WaitForAPI spins until the API can be reached successfully or the provided context is cancelled
func WaitForAPI(ctx context.Context, client APIClient) (err error) {
	for {
		select {
		case <-time.After(constants.PollInterval):
			err = client.Health(ctx)
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
