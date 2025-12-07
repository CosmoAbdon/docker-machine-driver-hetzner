package hetzner

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (c *Client) GetPlacementGroup(ctx context.Context, nameOrID string) (*hcloud.PlacementGroup, error) {
	grp, _, err := c.hcloud.PlacementGroup.Get(ctx, nameOrID)
	if err != nil {
		return nil, fmt.Errorf("could not get placement group: %w", err)
	}
	return grp, nil
}

func (c *Client) GetPlacementGroupsByLabel(ctx context.Context, labelSelector string) ([]*hcloud.PlacementGroup, error) {
	groups, err := c.hcloud.PlacementGroup.AllWithOpts(ctx, hcloud.PlacementGroupListOpts{
		ListOpts: hcloud.ListOpts{LabelSelector: labelSelector},
	})
	if err != nil {
		return nil, fmt.Errorf("could not list placement groups: %w", err)
	}
	return groups, nil
}

func (c *Client) CreatePlacementGroup(ctx context.Context, opts hcloud.PlacementGroupCreateOpts) (*hcloud.PlacementGroup, error) {
	result, _, err := c.hcloud.PlacementGroup.Create(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create placement group: %w", err)
	}
	return result.PlacementGroup, nil
}

func (c *Client) DeletePlacementGroup(ctx context.Context, pg *hcloud.PlacementGroup) error {
	_, err := c.hcloud.PlacementGroup.Delete(ctx, pg)
	if err != nil {
		return fmt.Errorf("could not delete placement group: %w", err)
	}
	return nil
}
