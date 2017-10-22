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
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/utils"

	"github.com/gravitational/trace"
)

// Rollup is the rollup configuration
type Rollup struct {
	// Retention is the retention policy for this rollup
	Retention string `json:"retention"`
	// Measurement is the name of the measurement to run rollup on
	Measurement string `json:"measurement"`
	// Name is both the name of the rollup query and the name of the
	// new measurement rollup data will be inserted into
	Name string `json:"name"`
	// Functions is a list of functions for rollup calculation
	Functions []Function `json:"functions"`
}

// Check verifies that rollup configuration is correct
func (r Rollup) Check() error {
	if !utils.OneOf(r.Retention, constants.AllRetentions) {
		return trace.BadParameter(
			"invalid Retention, must be one of: %v", constants.AllRetentions)
	}
	if r.Measurement == "" {
		return trace.BadParameter("parameter Measurement is missing")
	}
	if r.Name == "" {
		return trace.BadParameter("parameter Name is missing")
	}
	if len(r.Functions) == 0 {
		return trace.BadParameter("parameter Functions is empty")
	}
	for _, rollup := range r.Functions {
		err := rollup.Check()
		if err != nil {
			return trace.Wrap(err)
		}
	}
	return nil
}

// Function defines a single rollup function
type Function struct {
	// Function is the function name (mean, max, etc.)
	Function string `json:"function"`
	// Field is the name of the field to apply the function to
	Field string `json:"field"`
	// Alias is the optional alias for the new field in the rollup table
	Alias string `json:"alias,omitempty"`
}

// Check verifies the function configuration is correct
func (f Function) Check() error {
	if !utils.OneOf(f.Function, constants.AllFunctions) &&
		!strings.HasPrefix(f.Function, constants.FunctionPercentile) {
		return trace.BadParameter(
			"invalid Function, must be one of: %v, or start with %q",
			constants.AllFunctions, constants.FunctionPercentile)
	}
	if f.Field == "" {
		return trace.BadParameter("parameter Field is missing")
	}
	return nil
}

// buildQuery returns a string with InfluxDB query based on the rollup configuration
func buildQuery(r Rollup) (string, error) {
	var functions []string
	for _, fn := range r.Functions {
		function, err := buildFunction(fn)
		if err != nil {
			return "", trace.Wrap(err)
		}
		functions = append(functions, function)
	}

	var b bytes.Buffer
	err := queryTemplate.Execute(&b, map[string]string{
		"name":             r.Name,
		"database":         constants.InfluxDBDatabase,
		"functions":        strings.Join(functions, ", "),
		"retention_into":   r.Retention,
		"measurement_into": r.Name,
		"retention_from":   constants.InfluxDBRetentionPolicy,
		"measurement_from": r.Measurement,
		"interval":         constants.RetentionToInterval[r.Retention],
	})
	if err != nil {
		return "", trace.Wrap(err)
	}

	return b.String(), nil
}

// buildFunction returns a function string based on the provided function configuration
func buildFunction(f Function) (string, error) {
	alias := f.Alias
	if alias == "" {
		alias = f.Field
	}
	if strings.HasPrefix(f.Function, constants.FunctionPercentile) {
		value, err := parsePercentileValue(f.Function)
		if err != nil {
			return "", trace.Wrap(err)
		}
		return fmt.Sprintf("%v(%v, %v) as %v", constants.FunctionPercentile, f.Field, value, alias), nil
	}
	return fmt.Sprintf("%v(%v) as %v", f.Function, f.Field, alias), nil
}

// parsePercentileValue parses the percentile value from the strings like "percentile_90"
func parsePercentileValue(data string) (string, error) {
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		return "", trace.BadParameter(
			"percentile function must have format like 'percentile_90'")
	}
	value, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", trace.Wrap(err)
	}
	if value < 0 || value > 100 {
		return "", trace.BadParameter(
			"percentile value must be between 0 and 100 (inclusive)")
	}
	return parts[1], nil
}

var (
	// queryTemplate is the template of the InfluxDB rollup query
	queryTemplate = template.Must(template.New("query").Parse(
		`create continuous query "{{.name}}" on {{.database}} begin select {{.functions}} into {{.database}}."{{.retention_into}}"."{{.measurement_into}}" from {{.database}}."{{.retention_from}}"."{{.measurement_from}}" group by *, time({{.interval}}) end`))
)
