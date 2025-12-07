package hetzner

import (
	"context"
	"fmt"
	"net"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (c *Client) GetLocation(ctx context.Context, name string) (*hcloud.Location, error) {
	if name == "" {
		return nil, nil
	}

	location, _, err := c.hcloud.Location.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not get location by name: %w", err)
	}
	if location == nil {
		return nil, fmt.Errorf("unknown location: %v", name)
	}
	return location, nil
}

func (c *Client) GetServerType(ctx context.Context, name string) (*hcloud.ServerType, error) {
	stype, _, err := c.hcloud.ServerType.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not get server type by name: %w", err)
	}
	if stype == nil {
		return nil, fmt.Errorf("unknown server type: %v", name)
	}
	return stype, nil
}

func (c *Client) GetImageByID(ctx context.Context, id int64) (*hcloud.Image, error) {
	image, _, err := c.hcloud.Image.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not get image by ID %v: %w", id, err)
	}
	if image == nil {
		return nil, fmt.Errorf("image ID not found: %v", id)
	}
	return image, nil
}

func (c *Client) GetImageByNameAndArch(ctx context.Context, name string, arch hcloud.Architecture) (*hcloud.Image, error) {
	image, _, err := c.hcloud.Image.GetByNameAndArchitecture(ctx, name, arch)
	if err != nil {
		return nil, fmt.Errorf("could not get image by name %v: %w", name, err)
	}
	if image == nil {
		return nil, fmt.Errorf("image not found: %v[%v]", name, arch)
	}
	return image, nil
}

func (c *Client) GetPrimaryIP(ctx context.Context, nameOrIP string) (*hcloud.PrimaryIP, error) {
	if nameOrIP == "" {
		return nil, nil
	}

	client := c.hcloud.PrimaryIP

	var getter func(context.Context, string) (*hcloud.PrimaryIP, *hcloud.Response, error)
	if net.ParseIP(nameOrIP) != nil {
		getter = client.GetByIP
	} else {
		getter = client.Get
	}

	ip, _, err := getter(ctx, nameOrIP)
	if err != nil {
		return nil, fmt.Errorf("could not get primary IP: %w", err)
	}
	if ip == nil {
		return nil, fmt.Errorf("primary IP not found: %v", nameOrIP)
	}
	return ip, nil
}

func (c *Client) GetNetwork(ctx context.Context, nameOrID string) (*hcloud.Network, error) {
	network, _, err := c.hcloud.Network.Get(ctx, nameOrID)
	if err != nil {
		return nil, fmt.Errorf("could not get network by ID or name: %w", err)
	}
	if network == nil {
		return nil, fmt.Errorf("network '%s' not found", nameOrID)
	}
	return network, nil
}

func (c *Client) GetFirewall(ctx context.Context, nameOrID string) (*hcloud.Firewall, error) {
	firewall, _, err := c.hcloud.Firewall.Get(ctx, nameOrID)
	if err != nil {
		return nil, fmt.Errorf("could not get firewall by ID or name: %w", err)
	}
	if firewall == nil {
		return nil, fmt.Errorf("firewall '%s' not found", nameOrID)
	}
	return firewall, nil
}

func (c *Client) GetVolume(ctx context.Context, nameOrID string) (*hcloud.Volume, error) {
	volume, _, err := c.hcloud.Volume.Get(ctx, nameOrID)
	if err != nil {
		return nil, fmt.Errorf("could not get volume by ID or name: %w", err)
	}
	if volume == nil {
		return nil, fmt.Errorf("volume '%s' not found", nameOrID)
	}
	return volume, nil
}
