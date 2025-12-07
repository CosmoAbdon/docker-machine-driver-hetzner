package driver

import (
	"context"
	"fmt"
	"os"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/logging"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	mcnssh "github.com/docker/machine/libmachine/ssh"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"golang.org/x/crypto/ssh"
)

func (d *Driver) setupExistingKey() error {
	if !d.IsExistingKey {
		return nil
	}

	if d.originalKey == "" {
		return d.flagFailure("specifying an existing key ID requires the existing key path to be set as well")
	}

	key, err := d.getKey()
	if err != nil {
		return fmt.Errorf("could not get key: %w", err)
	}

	buf, err := os.ReadFile(d.originalKey + ".pub")
	if err != nil {
		return fmt.Errorf("could not read public key: %w", err)
	}

	// Will also parse `ssh-rsa w309jwf0e39jf asdf` public keys
	pubk, _, _, _, err := ssh.ParseAuthorizedKey(buf)
	if err != nil {
		return fmt.Errorf("could not parse authorized key: %w", err)
	}

	if key.Fingerprint != ssh.FingerprintLegacyMD5(pubk) &&
		key.Fingerprint != ssh.FingerprintSHA256(pubk) {
		return fmt.Errorf("remote key %d does not match local key %s", d.KeyID, d.originalKey)
	}

	return nil
}

func (d *Driver) copySSHKeyPair(src string) error {
	if err := mcnutils.CopyFile(src, d.GetSSHKeyPath()); err != nil {
		return fmt.Errorf("could not copy ssh key: %w", err)
	}

	if err := mcnutils.CopyFile(src+".pub", d.GetSSHKeyPath()+".pub"); err != nil {
		return fmt.Errorf("could not copy ssh public key: %w", err)
	}

	if err := os.Chmod(d.GetSSHKeyPath(), 0600); err != nil {
		return fmt.Errorf("could not set permissions on the ssh key: %w", err)
	}

	return nil
}

func (d *Driver) createRemoteKeys() error {
	if d.KeyID == 0 {
		log.Info("Creating SSH key...")

		buf, err := os.ReadFile(d.GetSSHKeyPath() + ".pub")
		if err != nil {
			return fmt.Errorf("could not read ssh public key: %w", err)
		}

		key, err := d.getRemoteKeyWithSameFingerprintNullable(buf)
		if err != nil {
			return fmt.Errorf("error retrieving potentially existing key: %w", err)
		}
		if key == nil {
			logging.Step("SSH key not found in Hetzner, uploading...")

			key, err = d.makeKey(d.GetMachineName(), string(buf), d.keyLabels)
			if err != nil {
				return err
			}
		} else {
			d.IsExistingKey = true
			logging.DebugStep("SSH key found in Hetzner [ID: %d]", key.ID)
		}

		d.KeyID = key.ID
	}
	for i, pubkey := range d.AdditionalKeys {
		key, err := d.getRemoteKeyWithSameFingerprintNullable([]byte(pubkey))
		if err != nil {
			return fmt.Errorf("error checking for existing key for %v: %w", pubkey, err)
		}
		if key == nil {
			logging.Step("Creating additional key #%d...", i)
			key, err = d.makeKey(fmt.Sprintf("%v-additional-%d", d.GetMachineName(), i), pubkey, d.keyLabels)

			if err != nil {
				return fmt.Errorf("error creating new key for %v: %w", pubkey, err)
			}

			logging.Substep("Created key [ID: %d]", key.ID)
			d.AdditionalKeyIDs = append(d.AdditionalKeyIDs, key.ID)
		} else {
			logging.Step("Using existing key %s", logging.Key(key.Name, key.ID))
		}

		d.cachedAdditionalKeys = append(d.cachedAdditionalKeys, key)
	}
	return nil
}

func (d *Driver) prepareLocalKey() error {
	if d.originalKey != "" {
		log.Debug("Copying SSH key...")
		if err := d.copySSHKeyPair(d.originalKey); err != nil {
			return fmt.Errorf("could not copy ssh key pair: %w", err)
		}
	} else {
		log.Debug("Generating SSH key...")
		if err := mcnssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			return fmt.Errorf("could not generate ssh key: %w", err)
		}
	}
	return nil
}

// Creates a new key for the machine and appends it to the dangling key list
func (d *Driver) makeKey(name string, pubkey string, labels map[string]string) (*hcloud.SSHKey, error) {
	keyopts := hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: pubkey,
		Labels:    labels,
	}

	key, err := d.getClient().CreateSSHKey(context.Background(), instrumented(keyopts))
	if err != nil {
		return nil, err
	}

	d.dangling = append(d.dangling, func() {
		err := d.getClient().DeleteSSHKey(context.Background(), key)
		if err != nil {
			log.Error(fmt.Errorf("could not delete ssh key: %w", err))
		}
	})

	return key, nil
}
