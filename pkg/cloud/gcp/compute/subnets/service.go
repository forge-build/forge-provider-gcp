/*
Copyright 2024 The Forge Authors.

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

package subnets

import (
	"context"

	k8scloud "github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/go-logr/logr"
	"google.golang.org/api/compute/v1"

	"github.com/forge-build/forge-provider-gcp/pkg/cloud"
)

const ServiceName = "subnets-reconciler"

type subnetsInterface interface {
	Get(ctx context.Context, key *meta.Key, options ...k8scloud.Option) (*compute.Subnetwork, error)
	Insert(ctx context.Context, key *meta.Key, obj *compute.Subnetwork, options ...k8scloud.Option) error
	Delete(ctx context.Context, key *meta.Key, options ...k8scloud.Option) error
}

// Scope is an interfaces that hold used methods.
type Scope interface {
	cloud.Build
	SubnetSpecs() []*compute.Subnetwork
}

// Service implements subnets reconciler.
type Service struct {
	scope   Scope
	subnets subnetsInterface
	Log     logr.Logger
}

var _ cloud.Reconciler = &Service{}

// New returns Service from given scope.
func New(scope Scope) *Service {
	cloudScope := scope.Cloud()
	if scope.IsSharedVpc() {
		cloudScope = scope.NetworkCloud()
	}

	return &Service{
		scope:   scope,
		subnets: cloudScope.Subnetworks(),
		Log:     scope.Log(ServiceName),
	}
}
