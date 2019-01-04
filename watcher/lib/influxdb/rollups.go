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
	"strconv"
	"strings"
	"text/template"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/utils"

	"github.com/gravitational/trace"
)

// Rollup is the rollup configuration
type Rollup struct {
	// Retention is the retention policy for this rollup
	Retention string `json:"retention"`
	// Measurement is the name of the measurement to run rollup on
	Measurement string `json:"measurement,omitempty"`
	// Name is both the name of the rollup query and the name of the
	// new measurement rollup data will be inserted into
	Name string `json:"name"`
	// Functions is a list of functions for rollup calculation
	Functions []Function `json:"functions"`
	// CustomFrom is a custom 'from' clause. Either CustomFrom or Measurement must be provided.
	CustomFrom string `json:"custom_from,omitempty"`
	// CustomGroupBy is a custom 'group by' clause
	CustomGroupBy string `json:"custom_group_by,omitempty"`
}

// Check verifies that rollup configuration is correct
func (r Rollup) Check() error {
	var errors []error
	if !utils.OneOf(r.Retention, constants.AllRetentions) {
		errors = append(errors, trace.BadParameter(
			"invalid Retention, must be one of: %v", constants.AllRetentions))
	}
	if r.Measurement == "" && r.CustomFrom == "" {
		errors = append(errors, trace.BadParameter("either Measurement or CustomFrom must be provided"))
	}
	if r.Name == "" {
		errors = append(errors, trace.BadParameter("parameter Name is missing"))
	}
	if len(r.Functions) == 0 {
		errors = append(errors, trace.BadParameter("parameter Functions is empty"))
	}
	for _, function := range r.Functions {
		if err := function.Check(); err != nil {
			errors = append(errors, trace.Wrap(err))
		}
	}
	return trace.NewAggregate(errors...)
}

// buildCreateQuery returns a string with InfluxDB query to create rollup
// based on the rollup configuration
func (r *Rollup) buildCreateQuery() (string, error) {
	var functions []string
	for _, fn := range r.Functions {
		function, err := fn.buildFunction()
		if err != nil {
			return "", trace.Wrap(err)
		}
		functions = append(functions, function)
	}

	var b bytes.Buffer
	err := createQueryTemplate.Execute(&b, map[string]string{
		"name":             r.Name,
		"database":         constants.InfluxDBDatabase,
		"functions":        strings.Join(functions, ", "),
		"retention_into":   r.Retention,
		"measurement_into": r.Name,
		"retention_from":   constants.InfluxDBRetentionPolicy,
		"measurement_from": r.Measurement,
		"custom_from":      r.CustomFrom,
		"custom_group_by":  r.CustomGroupBy,
		"interval":         constants.RetentionToInterval[r.Retention],
	})
	if err != nil {
		return "", trace.Wrap(err)
	}

	return b.String(), nil
}

// buildDeleteQuery returns a string with InfluxDB query to delete rollup
// based on the rollup configuration
func (r *Rollup) buildDeleteQuery() (string, error) {
	var b bytes.Buffer
	err := deleteQueryTemplate.Execute(&b, map[string]string{
		"name":     r.Name,
		"database": constants.InfluxDBDatabase,
	})
	if err != nil {
		return "", trace.Wrap(err)
	}

	return b.String(), nil
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
	var errors []error
	if !utils.OneOf(f.Function, constants.SimpleFunctions) && !f.isComposite() {
		errors = append(errors, trace.BadParameter(
			"invalid Function, must be one of %v, or a composite function starting with one of %v prefixes",
			constants.SimpleFunctions, constants.CompositeFunctions))
	}
	if f.isComposite() {
		funcAndValue := strings.Split(f.Function, "_")
		if len(funcAndValue) != 2 {
			errors = append(errors, trace.BadParameter(
				"percentile function must have format like 'percentile_90', 'top_10', 'bottom_10' or 'sample_1000' "))
		}
	}
	if f.Field == "" {
		errors = append(errors, trace.BadParameter("parameter Field is missing"))
	}
	return trace.NewAggregate(errors...)
}

// buildFunction returns a function string based on the provided function configuration
func (f *Function) buildFunction() (string, error) {
	alias := f.Alias
	if alias == "" {
		alias = f.Field
	}

	// split function name, based on the "_" separator (eg: percentile_99, top_10, ecc)
	err := f.Check()
	if err != nil {
		return "", trace.Wrap(err)
	}

	if !f.isComposite() {
		return fmt.Sprintf(`%v("%v") as %v`, f.Function, f.Field, alias), nil
	}

	funcAndValue := strings.Split(f.Function, "_")
	funcName := funcAndValue[0]
	param := funcAndValue[1]

	err = validateParam(funcName, param)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return fmt.Sprintf(`%v("%v", %v) as %v`, funcName, f.Field, param, alias), nil
}

// isComposite checks if the specified function is composite
func (f *Function) isComposite() bool {
	for _, name := range constants.CompositeFunctions {
		if strings.HasPrefix(f.Function, name) {
			return true
		}
	}
	return false
}

// validateParam checks the function parameter for validity.
func validateParam(funcName, param string) error {
	// convert parameter value as it's always going to be an Integer
	value, err := strconv.Atoi(param)
	if err != nil {
		return trace.Wrap(err)
	}

	switch funcName {
	case constants.FunctionPercentile:
		if value < 0 || value > 100 {
			return trace.BadParameter(
				"percentile value must be between 0 and 100 (inclusive)")
		}
	case constants.FunctionTop, constants.FunctionBottom, constants.FunctionSample:
		if value < 0 {
			return trace.BadParameter(
				"top, bottom and sample value must be greater than or equal to 0")
		}
	}

	return nil
}

var (
	// createQueryTemplate is the template for creating InfluxDB continuous query
	createQueryTemplate = template.Must(template.New("query").Parse(
		`create continuous query "{{.name}}" on {{.database}} begin select {{.functions}} into {{.database}}."{{.retention_into}}"."{{.measurement_into}}" from {{if .custom_from}}{{.custom_from}}{{else}}{{.database}}."{{.retention_from}}"."{{.measurement_from}}"{{end}} group by {{if .custom_group_by}}{{.custom_group_by}}{{else}}*, time({{.interval}}){{end}} end`))
	// deleteQueryTemplate is the template for deleting Influx continuous query
	deleteQueryTemplate = template.Must(template.New("query").Parse(
		`drop continuous query "{{.name}}" on {{.database}}`))
)
