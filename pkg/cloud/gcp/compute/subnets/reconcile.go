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

	infrav1 "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"google.golang.org/api/compute/v1"

	"sigs.k8s.io/cluster-api-provider-gcp/cloud/gcperrors"
)

// Reconcile reconciles cluster network components.
func (s *Service) Reconcile(ctx context.Context) error {
	s.Log.Info("Reconciling subnetwork resources")

	// reconcile subnets
	if _, err := s.createOrGetSubnets(ctx); err != nil {
		return err
	}

	return nil
}

// Delete deletes cluster subnetwork components.
func (s *Service) Delete(ctx context.Context) error {
	if s.scope.IsSharedVpc() {
		s.Log.V(1).Info("Shared VPC enabled. Skip deleting subnet resources")
		return nil
	}
	for _, subnetSpec := range s.scope.SubnetSpecs() {
		subnetKey := meta.RegionalKey(subnetSpec.Name, s.getSubnetRegion(subnetSpec))
		s.Log.V(1).Info("Looking for subnet before deleting it", "name", subnetSpec.Name)
		subnet, err := s.subnets.Get(ctx, subnetKey)
		if err != nil {
			if gcperrors.IsNotFound(err) {
				continue
			}
			s.Log.Error(err, "Error getting subnet", "name", subnetSpec.Name)
			return err
		}

		// Skip delete if subnet was not created by CAPG.
		// If subnet description is not set by the Spec, or by our default value, then assume it was created externally.
		if subnet.Description != infrav1.ClusterTagKey(s.scope.Name()) && (subnetSpec.Description == "" || subnet.Description != subnetSpec.Description) {
			s.Log.V(1).Info("Skipping subnet deletion as it was created outside of Cluster API", "name", subnetSpec.Name)
			return nil
		}

		s.Log.Info("Deleting a subnet", "name", subnetSpec.Name)
		if err := s.subnets.Delete(ctx, subnetKey); err != nil {
			if !gcperrors.IsNotFound(err) {
				s.Log.Error(err, "Error deleting subnet", "name", subnetSpec.Name)
				return err
			}
		}
	}

	return nil
}

// createOrGetSubnets creates the subnetworks if they don't exist otherwise return the existing ones.
func (s *Service) createOrGetSubnets(ctx context.Context) ([]*compute.Subnetwork, error) {
	subnets := []*compute.Subnetwork{}
	for _, subnetSpec := range s.scope.SubnetSpecs() {
		s.Log.V(1).Info("Looking for subnet", "name", subnetSpec.Name)
		subnetKey := meta.RegionalKey(subnetSpec.Name, s.getSubnetRegion(subnetSpec))
		subnet, err := s.subnets.Get(ctx, subnetKey)
		if err != nil {
			if !gcperrors.IsNotFound(err) {
				s.Log.Error(err, "Error looking for subnet", "name", subnetSpec.Name)
				return subnets, err
			}

			if s.scope.IsSharedVpc() {
				s.Log.Error(err, "Shared VPC is enabled, but could not find existing subnetwork", "name", subnetSpec.Name)
				return nil, err
			}

			// Subnet was not found, let's create it
			s.Log.V(1).Info("Creating a subnet", "name", subnetSpec.Name)
			if err := s.subnets.Insert(ctx, subnetKey, subnetSpec); err != nil {
				s.Log.Error(err, "Error creating a subnet", "name", subnetSpec.Name)
				return subnets, err
			}

			subnet, err = s.subnets.Get(ctx, subnetKey)
			if err != nil {
				s.Log.Error(err, "Error getting existing subnet", "name", subnetSpec.Name)
				return subnets, err
			}
		}
		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

// getSubnetRegion returns subnet region if user provided it, otherwise returns default scope region.
func (s *Service) getSubnetRegion(subnetSpec *compute.Subnetwork) string {
	if subnetSpec.Region != "" {
		return subnetSpec.Region
	}
	return s.scope.Region()
}
