package hetzner

import (
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type Client struct {
	hcloud       *hcloud.Client
	pollInterval time.Duration
}

type ClientConfig struct {
	Token           string
	AppName         string
	AppVersion      string
	PollInterval    time.Duration
	AdditionalOpts  []hcloud.ClientOption
}

func NewClient(cfg ClientConfig) *Client {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 1 * time.Second
	}
	
	opts := []hcloud.ClientOption{
		hcloud.WithToken(cfg.Token),
		hcloud.WithApplication(cfg.AppName, cfg.AppVersion),
		hcloud.WithPollOpts(hcloud.PollOpts{
			BackoffFunc: hcloud.ConstantBackoff(cfg.PollInterval),
		}),
	}
	
	opts = append(opts, cfg.AdditionalOpts...)
	
	return &Client{
		hcloud:       hcloud.NewClient(opts...),
		pollInterval: cfg.PollInterval,
	}
}

func (c *Client) HCloud() *hcloud.Client {
	return c.hcloud
}

func (c *Client) PollInterval() time.Duration {
	return c.pollInterval
}
