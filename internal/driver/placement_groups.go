package driver

import (
	"context"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/config"
	"github.com/docker/machine/libmachine/log"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (d *Driver) getAutoPlacementGroup() (*hcloud.PlacementGroup, error) {
	res, err := d.getClient().GetPlacementGroupsByLabel(context.Background(), config.LabelName(config.LabelAutoSpreadPG))
	if err != nil {
		return nil, err
	}

	if len(res) != 0 {
		return res[0], nil
	}

	grp, err := d.makePlacementGroup("Docker-Machine auto spread", map[string]string{
		config.LabelName(config.LabelAutoSpreadPG): "true",
		config.LabelName(config.LabelAutoCreated):  "true",
	})

	return instrumented(grp), err
}

func (d *Driver) makePlacementGroup(name string, labels map[string]string) (*hcloud.PlacementGroup, error) {
	grp, err := d.getClient().CreatePlacementGroup(context.Background(), instrumented(hcloud.PlacementGroupCreateOpts{
		Name:   name,
		Labels: labels,
		Type:   "spread",
	}))

	if grp != nil {
		d.dangling = append(d.dangling, func() {
			err := d.getClient().DeletePlacementGroup(context.Background(), grp)
			if err != nil {
				log.Errorf("Could not delete placement group: %v", err)
			}
		})
	}

	if err != nil {
		return nil, err
	}

	return instrumented(grp), nil
}

func (d *Driver) getPlacementGroup() (*hcloud.PlacementGroup, error) {
	if d.placementGroup == "" {
		return nil, nil
	} else if d.cachedPGrp != nil {
		return d.cachedPGrp, nil
	}

	name := d.placementGroup
	if name == config.AutoSpreadPGName {
		grp, err := d.getAutoPlacementGroup()
		d.cachedPGrp = grp
		return grp, err
	} else {
		grp, err := d.getClient().GetPlacementGroup(context.Background(), name)
		if err != nil {
			return nil, err
		}

		if grp != nil {
			return grp, nil
		}

		return d.makePlacementGroup(name, map[string]string{config.LabelName(config.LabelAutoCreated): "true"})
	}
}
