package hetzner

import (
	"context"
	"errors"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (c *Client) GetServerByID(ctx context.Context, id int64) (*hcloud.Server, error) {
	if id == 0 {
		return nil, errors.New("server ID was 0")
	}

	srv, _, err := c.hcloud.Server.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not get server by ID: %w", err)
	}
	return srv, nil
}

func (c *Client) CreateServer(ctx context.Context, opts hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, error) {
	result, _, err := c.hcloud.Server.Create(ctx, opts)
	if err != nil {
		return hcloud.ServerCreateResult{}, fmt.Errorf("could not create server: %w", err)
	}
	return result, nil
}

func (c *Client) DeleteServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, error) {
	result, _, err := c.hcloud.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("could not delete server: %w", err)
	}
	return result.Action, nil
}

func (c *Client) RebootServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, error) {
	action, _, err := c.hcloud.Server.Reboot(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("could not reboot server: %w", err)
	}
	return action, nil
}

func (c *Client) PowerOnServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, error) {
	action, _, err := c.hcloud.Server.Poweron(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("could not power on server: %w", err)
	}
	return action, nil
}

func (c *Client) ShutdownServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, error) {
	action, _, err := c.hcloud.Server.Shutdown(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("could not shutdown server: %w", err)
	}
	return action, nil
}

func (c *Client) PowerOffServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, error) {
	action, _, err := c.hcloud.Server.Poweroff(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("could not power off server: %w", err)
	}
	return action, nil
}
