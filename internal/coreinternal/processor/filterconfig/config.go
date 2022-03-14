// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filterconfig // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterconfig"

import (
	"errors"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset"
)

// MatchConfig has two optional MatchProperties one to define what is processed
// by the processor, captured under the 'include' and the second, exclude, to
// define what is excluded from the processor.
type MatchConfig struct {
	// Include specifies the set of input data properties that must be present in order
	// for this processor to apply to it.
	// Note: If `exclude` is specified, the input data is compared against those
	// properties after the `include` properties.
	// This is an optional field. If neither `include` and `exclude` are set, all input data
	// are processed. If `include` is set and `exclude` isn't set, then all
	// input data matching the properties in this structure are processed.
	Include *MatchProperties `mapstructure:"include"`

	// Exclude specifies when this processor will not be applied to the input data
	// which match the specified properties.
	// Note: The `exclude` properties are checked after the `include` properties,
	// if they exist, are checked.
	// If `include` isn't specified, the `exclude` properties are checked against
	// all input data.
	// This is an optional field. If neither `include` and `exclude` are set, all input data
	// is processed. If `exclude` is set and `include` isn't set, then all the
	// input data that does not match the properties in this structure are processed.
	Exclude *MatchProperties `mapstructure:"exclude"`
}

// MatchProperties specifies the set of properties in a spans/log/metric to match
// against and if the input data should be included or excluded from the
// processor. At least one of services (spans only), names or
// attributes must be specified. It is supported to have all specified, but
// this requires all the properties to match for the inclusion/exclusion to
// occur.
// The following are examples of invalid configurations:
//  attributes/bad1:
//    # This is invalid because include is specified with neither services or
//    # attributes.
//    include:
//    actions: ...
//
//  span/bad2:
//    exclude:
//    	# This is invalid because services, span_names and attributes have empty values.
//      services:
//      span_names:
//      attributes:
//    actions: ...
// Please refer to processor/attributesprocessor/testdata/config.yaml and
// processor/spanprocessor/testdata/config.yaml for valid configurations.
type MatchProperties struct {
	// Config configures the matching patterns used when matching span properties.
	filterset.Config `mapstructure:",squash"`

	// Note: For spans, one of Services, SpanNames, Attributes, Resources or Libraries must be specified with a
	// non-empty value for a valid configuration.

	// For logs, one of LogNames, Attributes, Resources or Libraries must be specified with a
	// non-empty value for a valid configuration.

	// For metrics, one of MetricNames, Expressions, or ResourceAttributes must be specified with a
	// non-empty value for a valid configuration.

	// Services specify the list of items to match service name against.
	// A match occurs if the span's service name matches at least one item in this list.
	// This is an optional field.
	Services []string `mapstructure:"services"`

	// SpanNames specify the list of items to match span name against.
	// A match occurs if the span name matches at least one item in this list.
	// This is an optional field.
	SpanNames []string `mapstructure:"span_names"`

	// LogNames is a list of strings that the LogRecord's name field must match
	// against.
	// Deprecated: the Name field is removed from the log data model.
	LogNames []string `mapstructure:"log_names"`

	// MetricNames is a list of strings to match metric name against.
	// A match occurs if metric name matches at least one item in the list.
	// This field is optional.
	MetricNames []string `mapstructure:"metric_names"`

	// Attributes specifies the list of attributes to match against.
	// All of these attributes must match exactly for a match to occur.
	// Only match_type=strict is allowed if "attributes" are specified.
	// This is an optional field.
	Attributes []Attribute `mapstructure:"attributes"`

	// Resources specify the list of items to match the resources against.
	// A match occurs if the data's resources match at least one item in this list.
	// This is an optional field.
	Resources []Attribute `mapstructure:"resources"`

	// Libraries specify the list of items to match the implementation library against.
	// A match occurs if the span's implementation library matches at least one item in this list.
	// This is an optional field.
	Libraries []InstrumentationLibrary `mapstructure:"libraries"`

	// Expressions specifies the list of expr expressions to match metrics against,
	// and is not supported for other data types.
	// A match occurs if any datapoint in a metric matches at least one expression in this list.
	Expressions []string `mapstructure:"expressions"`

	// ResourceAttributes specify the list of items to match metrics resources against.
	// A match occurs if the data's resources match at least one item in this list.
	// Deprecated: use Resources field instead
	ResourceAttributes []Attribute `mapstructure:"resource_attributes"`
}

// ValidateForSpans validates properties for spans.
func (mp *MatchProperties) ValidateForSpans() error {
	if len(mp.LogNames) > 0 || len(mp.MetricNames) > 0 || len(mp.Expressions) > 0 || len(mp.ResourceAttributes) > 0 {
		return errors.New("none of log_names, metric_names, expressions or resource_attributes should be specified for trace spans")
	}

	if mp.MatchType == filterset.Expr {
		return errors.New("expr match type can only be used with metrics")
	}

	if len(mp.Services) == 0 && len(mp.SpanNames) == 0 && len(mp.Attributes) == 0 &&
		len(mp.Libraries) == 0 && len(mp.Resources) == 0 {
		return errors.New(`at least one of "services", "span_names", "attributes", "libraries" or "resources" field must be specified`)
	}

	return nil
}

// ValidateForLogs validates properties for logs.
func (mp *MatchProperties) ValidateForLogs() error {
	if len(mp.Services) > 0 || len(mp.SpanNames) > 0 || len(mp.MetricNames) > 0 || len(mp.Expressions) > 0 || len(mp.ResourceAttributes) > 0 {
		return errors.New("none of services, span_names, metric_names, expressions or resource_attributes should be specified for log records")
	}

	if mp.MatchType == filterset.Expr {
		return errors.New("expr match type can only be used with metrics")
	}

	if len(mp.Attributes) == 0 && len(mp.Libraries) == 0 && len(mp.Resources) == 0 {
		return errors.New(`at least one of "attributes", "libraries" or "resources" field must be specified`)
	}

	return nil
}

// ValidateForMetrics validates properties for metrics.
func (mp *MatchProperties) ValidateForMetrics() error {
	if len(mp.Services) > 0 || len(mp.SpanNames) > 0 || len(mp.LogNames) > 0 {
		return errors.New("none of services, span_names nor log_names should be specified for metrics records")
	}
	if len(mp.ResourceAttributes) > 0 && len(mp.Resources) > 0 {
		return errors.New("cannot specify both resources and resource_attributes - prefer resources")
	}

	// Existing tests make use of empty metrics filters, so we don't currently validate that this has relevant fields set
	return nil
}

func (mp *MatchProperties) ChecksMetrics() bool {
	if mp == nil {
		return false
	} else if mp.MatchType == filterset.Expr {
		return len(mp.Expressions) > 0
	} else {
		return len(mp.MetricNames) > 0 || len(mp.Attributes) > 0
	}
}

func (mp *MatchProperties) ChecksResourceAtributes() bool {
	if mp == nil {
		return false
	} else {
		return len(mp.Resources) > 0 || len(mp.ResourceAttributes) > 0
	}
}

// Attribute specifies the attribute key and optional value to match against.
type Attribute struct {
	// Key specifies the attribute key.
	Key string `mapstructure:"key"`

	// Values specifies the value to match against.
	// If it is not set, any value will match.
	Value interface{} `mapstructure:"value"`
}

// InstrumentationLibrary specifies the instrumentation library and optional version to match against.
type InstrumentationLibrary struct {
	Name string `mapstructure:"name"`
	// version match
	//  expected actual  match
	//  nil      <blank> yes
	//  nil      1       yes
	//  <blank>  <blank> yes
	//  <blank>  1       no
	//  1        <blank> no
	//  1        1       yes
	Version *string `mapstructure:"version"`
}
