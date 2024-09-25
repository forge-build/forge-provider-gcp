package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"
)

// Labels defines a map of tags.
type Labels map[string]string

// Equals returns true if the tags are equal.
func (in Labels) Equals(other Labels) bool {
	return reflect.DeepEqual(in, other)
}

// HasOwned returns true if the tags contains a tag that marks the resource as owned by the cluster from the perspective of this management tooling.
func (in Labels) HasOwned(build string) bool {
	value, ok := in[BuildTagKey(build)]

	return ok && ResourceLifecycle(value) == ResourceLifecycleOwned
}

// ToComputeFilter returns the string representation of the labels as a filter
// to be used in google compute sdk calls.
func (in Labels) ToComputeFilter() string {
	var builder strings.Builder
	for k, v := range in {
		builder.WriteString(fmt.Sprintf("(labels.%s = %q) ", k, v))
	}

	return builder.String()
}

// Difference returns the difference between this map of tags and the other map of tags.
// Items are considered equals if key and value are equals.
func (in Labels) Difference(other Labels) Labels {
	res := make(Labels, len(in))

	for key, value := range in {
		if otherValue, ok := other[key]; ok && value == otherValue {
			continue
		}
		res[key] = value
	}

	return res
}

// AddLabels adds (and overwrites) the current labels with the ones passed in.
func (in Labels) AddLabels(other Labels) Labels {
	for key, value := range other {
		if in == nil {
			in = make(map[string]string, len(other))
		}
		in[key] = value
	}

	return in
}

// ResourceLifecycle configures the lifecycle of a resource.
type ResourceLifecycle string

const (
	// ResourceLifecycleOwned is the value we use when tagging resources to indicate
	// that the resource is considered owned and managed by the cluster,
	// and in particular that the lifecycle is tied to the lifecycle of the cluster.
	ResourceLifecycleOwned = ResourceLifecycle("owned")

	// NameGCPProviderPrefix is the tag prefix we use to differentiate
	// forge-provider-gcp owned components from other tooling that
	// uses NameKubernetesClusterPrefix.
	NameGCPProviderPrefix = "forge-gcp"

	// NameGCPProviderOwned is the tag name we use to differentiate
	// forge-provider-gcp owned components from other tooling that
	// uses NameKubernetesClusterPrefix.
	NameGCPProviderOwned = NameGCPProviderPrefix + "build-"
)

// BuildTagKey generates the key for resources associated with a build.
func BuildTagKey(name string) string {
	return fmt.Sprintf("%s%s", NameGCPProviderOwned, name)
}

// BuildParams is used to build tags around an gcp resource.
type BuildParams struct {
	// Lifecycle determines the resource lifecycle.
	Lifecycle ResourceLifecycle

	// ClusterName is the cluster associated with the resource.
	BuildName string

	// ResourceID is the unique identifier of the resource to be tagged.
	ResourceID string

	// Any additional tags to be added to the resource.
	// +optional
	Additional Labels
}

// Build builds tags including the cluster tag and returns them in map form.
func Build(params BuildParams) Labels {
	tags := make(Labels)
	for k, v := range params.Additional {
		tags[strings.ToLower(k)] = strings.ToLower(v)
	}

	tags[BuildTagKey(params.BuildName)] = string(params.Lifecycle)

	return tags
}
