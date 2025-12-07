package hetzner

import (
	"context"
	"fmt"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/logging"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (c *Client) WaitForAction(ctx context.Context, action *hcloud.Action) error {
	lastProgress := 0
	
	err := c.hcloud.Action.WaitForFunc(ctx, func(update *hcloud.Action) error {
		if update.Progress != lastProgress {
			logging.DebugStep("Action %s: %d%%", logging.Action(update.Command, update.ID), update.Progress)
			lastProgress = update.Progress
		}
		return nil
	}, action)
	
	if err == nil {
		logging.DebugStep("Action %s completed", logging.Action(action.Command, action.ID))
	}
	
	return err
}

func (c *Client) WaitForActions(ctx context.Context, stepName string, actions []*hcloud.Action) error {
	if len(actions) == 0 {
		return nil
	}
	
	logging.DebugStep("%s: starting", stepName)

	for _, action := range actions {
		lastProgress := 0
		
		err := c.hcloud.Action.WaitForFunc(ctx, func(update *hcloud.Action) error {
			if update.Progress != lastProgress {
				logging.DebugStep("%s: %d%%", stepName, update.Progress)
				lastProgress = update.Progress
			}
			return nil
		}, action)

		if err != nil {
			return fmt.Errorf("%s: %w", stepName, err)
		}
	}

	logging.DebugStep("%s completed", stepName)
	return nil
}
