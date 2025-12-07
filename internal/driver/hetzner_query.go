package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/hetzner"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (d *Driver) getClient() *hetzner.Client {
	if d.cachedClient != nil {
		return d.cachedClient
	}

	d.cachedClient = hetzner.NewClient(hetzner.ClientConfig{
		Token:          d.AccessToken,
		AppName:        "docker-machine-driver",
		AppVersion:     d.version,
		PollInterval:   time.Duration(d.WaitOnPolling) * time.Second,
		AdditionalOpts: d.getClientInstrumentationOpts(),
	})

	return d.cachedClient
}

func (d *Driver) getLocationNullable() (*hcloud.Location, error) {
	if d.cachedLocation != nil {
		return d.cachedLocation, nil
	}

	location, err := d.getClient().GetLocation(context.Background(), d.Location)
	if err != nil {
		return nil, err
	}
	d.cachedLocation = location
	return location, nil
}

func (d *Driver) getType() (*hcloud.ServerType, error) {
	if d.cachedType != nil {
		return d.cachedType, nil
	}

	stype, err := d.getClient().GetServerType(context.Background(), d.Type)
	if err != nil {
		return nil, err
	}
	d.cachedType = stype
	return instrumented(stype), nil
}

func (d *Driver) getImage() (*hcloud.Image, error) {
	if d.cachedImage != nil {
		return d.cachedImage, nil
	}

	var image *hcloud.Image
	var err error

	if d.ImageID != 0 {
		image, err = d.getClient().GetImageByID(context.Background(), d.ImageID)
		if err != nil {
			return nil, err
		}
	} else {
		arch, err := d.getImageArchitectureForLookup()
		if err != nil {
			return nil, fmt.Errorf("could not determine image architecture: %w", err)
		}

		image, err = d.getClient().GetImageByNameAndArch(context.Background(), d.Image, arch)
		if err != nil {
			return nil, err
		}
	}

	d.cachedImage = image
	return instrumented(image), nil
}

func (d *Driver) getImageArchitectureForLookup() (hcloud.Architecture, error) {
	if d.ImageArch != emptyImageArchitecture {
		return d.ImageArch, nil
	}

	serverType, err := d.getType()
	if err != nil {
		return "", err
	}

	return serverType.Architecture, nil
}

func (d *Driver) getKey() (*hcloud.SSHKey, error) {
	key, err := d.getKeyNullable()
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("key not found: %v", d.KeyID)
	}
	return key, err
}

func (d *Driver) getKeyNullable() (*hcloud.SSHKey, error) {
	if d.cachedKey != nil {
		return d.cachedKey, nil
	}

	key, err := d.getClient().GetSSHKeyByID(context.Background(), d.KeyID)
	if err != nil {
		return nil, err
	}
	d.cachedKey = key
	return instrumented(key), nil
}

func (d *Driver) getRemoteKeyWithSameFingerprintNullable(publicKeyBytes []byte) (*hcloud.SSHKey, error) {
	remoteKey, err := d.getClient().GetSSHKeyByPublicKey(context.Background(), publicKeyBytes)
	if err != nil {
		return nil, err
	}
	return instrumented(remoteKey), nil
}

func (d *Driver) getServerHandle() (*hcloud.Server, error) {
	srv, err := d.getServerHandleNullable()
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, fmt.Errorf("server does not exist: %v", d.ServerID)
	}
	return srv, nil
}

func (d *Driver) getServerHandleNullable() (*hcloud.Server, error) {
	if d.cachedServer != nil {
		return d.cachedServer, nil
	}

	srv, err := d.getClient().GetServerByID(context.Background(), d.ServerID)
	if err != nil {
		return nil, err
	}

	d.cachedServer = srv
	return srv, nil
}

func (d *Driver) waitForAction(a *hcloud.Action) error {
	return d.getClient().WaitForAction(context.Background(), a)
}

func (d *Driver) waitForMultipleActions(step string, a []*hcloud.Action) error {
	return d.getClient().WaitForActions(context.Background(), step, a)
}
