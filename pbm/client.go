/*
Copyright (c) 2017 VMware, Inc. All Rights Reserved.

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

package pbm

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/pbm/methods"
	"github.com/vmware/govmomi/pbm/types"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	vim "github.com/vmware/govmomi/vim25/types"
)

type Client struct {
	*soap.Client

	ServiceContent types.PbmServiceInstanceContent
}

func NewClient(ctx context.Context, c *vim25.Client) (*Client, error) {
	sc := c.Client.NewServiceClient("/pbm/sdk", "urn:pbm")

	req := types.PbmRetrieveServiceContent{
		This: vim.ManagedObjectReference{
			Type:  "PbmServiceInstance",
			Value: "ServiceInstance",
		},
	}

	res, err := methods.PbmRetrieveServiceContent(ctx, sc, &req)
	if err != nil {
		return nil, err
	}

	return &Client{sc, res.Returnval}, nil
}

func (c *Client) QueryProfile(ctx context.Context, rtype types.PbmProfileResourceType, category string) ([]types.PbmProfileId, error) {
	req := types.PbmQueryProfile{
		This:            c.ServiceContent.ProfileManager,
		ResourceType:    rtype,
		ProfileCategory: category,
	}

	res, err := methods.PbmQueryProfile(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) RetrieveContent(ctx context.Context, ids []types.PbmProfileId) ([]types.BasePbmProfile, error) {
	req := types.PbmRetrieveContent{
		This:       c.ServiceContent.ProfileManager,
		ProfileIds: ids,
	}

	res, err := methods.PbmRetrieveContent(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) CheckRequirements(ctx context.Context, hubs []types.PbmPlacementHub, ref *types.PbmServerObjectRef, preq []types.BasePbmPlacementRequirement) ([]types.PbmPlacementCompatibilityResult, error) {
	req := types.PbmCheckRequirements{
		This:                        c.ServiceContent.PlacementSolver,
		HubsToSearch:                hubs,
		PlacementSubjectRef:         ref,
		PlacementSubjectRequirement: preq,
	}

	res, err := methods.PbmCheckRequirements(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) GetCompatibleDatastores(ctx context.Context, datastores []vim.ManagedObjectReference, ref *types.PbmServerObjectRef, ids []types.PbmProfileId) ([]vim.ManagedObjectReference, error) {
	var hubs []types.PbmPlacementHub
	var req []types.BasePbmPlacementRequirement
	var compatibleDatastores []vim.ManagedObjectReference

	for _, ds := range datastores {
		hubs = append(hubs, types.PbmPlacementHub{
			HubType: ds.Type,
			HubId:   ds.Value,
		})
	}

	for _, id := range ids {
		req = append(req, &types.PbmPlacementCapabilityProfileRequirement{
			ProfileId: id,
		})
	}

	compatibilityResult, err := c.CheckRequirements(ctx, hubs, ref, req)
	if err != nil {
		return nil, err
	}

	for _, res := range compatibilityResult {
		if len(res.Error) == 0 {
			compatibleDatastores = append(compatibleDatastores, vim.ManagedObjectReference{
				Type:  res.Hub.HubType,
				Value: res.Hub.HubId,
			})
		}
	}

	return compatibleDatastores, nil
}

func (c *Client) GetNonCompatibleDatastores(ctx context.Context, datastores []vim.ManagedObjectReference, ref *types.PbmServerObjectRef, ids []types.PbmProfileId) ([]vim.ManagedObjectReference, error) {
	var hubs []types.PbmPlacementHub
	var req []types.BasePbmPlacementRequirement
	var nonCompatibleDatastores []vim.ManagedObjectReference

	for _, ds := range datastores {
		hubs = append(hubs, types.PbmPlacementHub{
			HubType: ds.Type,
			HubId:   ds.Value,
		})
	}

	for _, id := range ids {
		req = append(req, &types.PbmPlacementCapabilityProfileRequirement{
			ProfileId: id,
		})
	}

	compatibilityResult, err := c.CheckRequirements(ctx, hubs, nil, req)
	if err != nil {
		return nil, err
	}

	for _, res := range compatibilityResult {
		if len(res.Error) > 0 {
			nonCompatibleDatastores = append(nonCompatibleDatastores, vim.ManagedObjectReference{
				Type:  res.Hub.HubType,
				Value: res.Hub.HubId,
			})
		}
	}

	return nonCompatibleDatastores, nil
}

func (c *Client) CreatePbmProfile(ctx context.Context, capabilityProfileCreateSpec types.PbmCapabilityProfileCreateSpec) (types.PbmProfileId, error) {
	req := types.PbmCreate{
		This:       c.ServiceContent.ProfileManager,
		CreateSpec: capabilityProfileCreateSpec,
	}

	res, err := methods.PbmCreate(ctx, c, &req)
	if err != nil {
		var profileID types.PbmProfileId
		return profileID, err
	}

	return res.Returnval, nil
}

func (c *Client) UpdatePbmProfile(ctx context.Context, id types.PbmProfileId, capabilityProfileUpdateSpec types.PbmCapabilityProfileUpdateSpec) error {
	req := types.PbmUpdate{
		This:       c.ServiceContent.ProfileManager,
		ProfileId:  id,
		UpdateSpec: capabilityProfileUpdateSpec,
	}

	_, err := methods.PbmUpdate(ctx, c, &req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeletePbmProfile(ctx context.Context, ids []types.PbmProfileId) ([]types.PbmProfileOperationOutcome, error) {
	req := types.PbmDelete{
		This:      c.ServiceContent.ProfileManager,
		ProfileId: ids,
	}

	res, err := methods.PbmDelete(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) QueryAssociatedEntity(ctx context.Context, id types.PbmProfileId, entityType string) ([]types.PbmServerObjectRef, error) {
	req := types.PbmQueryAssociatedEntity{
		This:       c.ServiceContent.ProfileManager,
		Profile:    id,
		EntityType: entityType,
	}

	res, err := methods.PbmQueryAssociatedEntity(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) QueryAssociatedEntities(ctx context.Context, ids []types.PbmProfileId) ([]types.PbmQueryProfileResult, error) {
	req := types.PbmQueryAssociatedEntities{
		This:     c.ServiceContent.ProfileManager,
		Profiles: ids,
	}

	res, err := methods.PbmQueryAssociatedEntities(ctx, c, &req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}

func (c *Client) GetPbmProfileIDByName(ctx context.Context, profileName string) (string, error) {
	resourceType := types.PbmProfileResourceType{
		ResourceType: string(types.PbmProfileResourceTypeEnumSTORAGE),
	}
	category := types.PbmProfileCategoryEnumREQUIREMENT
	ids, err := c.QueryProfile(ctx, resourceType, string(category))
	if err != nil {
		return "", err
	}

	profiles, err := c.RetrieveContent(ctx, ids)
	if err != nil {
		return "", err
	}

	for i := range profiles {
		profile := profiles[i].GetPbmProfile()
		if profile.Name == profileName {
			return profile.ProfileId.UniqueId, nil
		}
	}
	return "", fmt.Errorf("No PBM profile found with name: %q", profileName)
}
