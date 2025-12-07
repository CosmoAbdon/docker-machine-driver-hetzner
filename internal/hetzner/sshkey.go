package hetzner

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"golang.org/x/crypto/ssh"
)

func (c *Client) GetSSHKeyByID(ctx context.Context, id int64) (*hcloud.SSHKey, error) {
	key, _, err := c.hcloud.SSHKey.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not get SSH key by ID: %w", err)
	}
	return key, nil
}

func (c *Client) GetSSHKeyByFingerprint(ctx context.Context, fingerprint string) (*hcloud.SSHKey, error) {
	key, _, err := c.hcloud.SSHKey.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("could not get SSH key by fingerprint: %w", err)
	}
	return key, nil
}

func (c *Client) GetSSHKeyByPublicKey(ctx context.Context, publicKeyBytes []byte) (*hcloud.SSHKey, error) {
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse SSH public key: %w", err)
	}

	fingerprint := ssh.FingerprintLegacyMD5(publicKey)
	return c.GetSSHKeyByFingerprint(ctx, fingerprint)
}

func (c *Client) CreateSSHKey(ctx context.Context, opts hcloud.SSHKeyCreateOpts) (*hcloud.SSHKey, error) {
	key, _, err := c.hcloud.SSHKey.Create(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create SSH key: %w", err)
	}
	if key == nil {
		return nil, fmt.Errorf("SSH key creation returned nil without error")
	}
	return key, nil
}

func (c *Client) DeleteSSHKey(ctx context.Context, key *hcloud.SSHKey) error {
	_, err := c.hcloud.SSHKey.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("could not delete SSH key: %w", err)
	}
	return nil
}
