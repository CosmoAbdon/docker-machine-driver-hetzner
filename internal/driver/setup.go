package driver

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"go.yaml.in/yaml/v2"
)

func (d *Driver) waitForRunningServer() error {
	start_time := time.Now()
	for {
		srvstate, err := d.GetState()
		if err != nil {
			return fmt.Errorf("could not get state: %w", err)
		}

		if srvstate == state.Running {
			break
		}

		elapsed_time := time.Since(start_time).Seconds()
		if d.WaitForRunningTimeout > 0 && int(elapsed_time) > d.WaitForRunningTimeout {
			return fmt.Errorf("server exceeded wait-for-running-timeout")
		}

		time.Sleep(time.Duration(d.WaitOnPolling) * time.Second)
	}
	return nil
}

func (d *Driver) waitForInitialStartup(srv hcloud.ServerCreateResult) error {
	if len(srv.NextActions) != 0 {
		if err := d.waitForMultipleActions("server.NextActions", srv.NextActions); err != nil {
			return fmt.Errorf("could not wait for NextActions: %w", err)
		}
	}

	return d.waitForRunningServer()
}

func (d *Driver) makeCreateServerOptions() (*hcloud.ServerCreateOpts, error) {
	pgrp, err := d.getPlacementGroup()
	if err != nil {
		return nil, err
	}

	userData, err := d.getUserData()
	if err != nil {
		return nil, err
	}

	srvopts := hcloud.ServerCreateOpts{
		Name:           d.GetMachineName(),
		UserData:       userData,
		Labels:         d.ServerLabels,
		PlacementGroup: pgrp,
	}

	err = d.setPublicNetIfRequired(&srvopts)
	if err != nil {
		return nil, err
	}

	networks, err := d.createNetworks()
	if err != nil {
		return nil, err
	}
	srvopts.Networks = networks

	firewalls, err := d.createFirewalls()
	if err != nil {
		return nil, err
	}
	srvopts.Firewalls = firewalls

	volumes, err := d.createVolumes()
	if err != nil {
		return nil, err
	}
	srvopts.Volumes = volumes

	if srvopts.Location, err = d.getLocationNullable(); err != nil {
		return nil, fmt.Errorf("could not get location: %w", err)
	}
	if srvopts.ServerType, err = d.getType(); err != nil {
		return nil, fmt.Errorf("could not get type: %w", err)
	}
	if srvopts.Image, err = d.getImage(); err != nil {
		return nil, fmt.Errorf("could not get image: %w", err)
	}
	key, err := d.getKey()
	if err != nil {
		return nil, fmt.Errorf("could not get ssh key: %w", err)
	}
	srvopts.SSHKeys = append(d.cachedAdditionalKeys, key)
	return &srvopts, nil
}

func (d *Driver) getUserData() (string, error) {
	var baseUserData string

	if d.userDataFile != "" {
		readUserData, err := os.ReadFile(d.userDataFile)
		if err != nil {
			return "", fmt.Errorf("could not read user data file: %w", err)
		}
		baseUserData = string(readUserData)
	} else {
		baseUserData = d.userData
	}

	if d.additionalUserData == "" {
		return baseUserData, nil
	}

	if baseUserData == "" {
		return d.additionalUserData, nil
	}

	merged, err := mergeUserData(baseUserData, d.additionalUserData)
	if err != nil {
		return "", fmt.Errorf("could not merge user data: %w", err)
	}
	return merged, nil
}

func mergeUserData(base, additional string) (string, error) {
	var baseMap, additionalMap map[string]interface{}

	baseContent := strings.TrimPrefix(base, "#cloud-config\n")
	additionalContent := strings.TrimPrefix(additional, "#cloud-config\n")

	if err := yaml.Unmarshal([]byte(baseContent), &baseMap); err != nil {
		return "", fmt.Errorf("could not parse base user data as YAML: %w", err)
	}

	if err := yaml.Unmarshal([]byte(additionalContent), &additionalMap); err != nil {
		return "", fmt.Errorf("could not parse additional user data as YAML: %w", err)
	}

	if baseMap == nil {
		baseMap = make(map[string]interface{})
	}

	mergeYAMLMaps(baseMap, additionalMap)

	merged, err := yaml.Marshal(baseMap)
	if err != nil {
		return "", fmt.Errorf("could not serialize merged user data: %w", err)
	}

	return "#cloud-config\n" + string(merged), nil
}

func mergeYAMLMaps(base, additional map[string]interface{}) {
	for key, additionalValue := range additional {
		baseValue, exists := base[key]
		if !exists {
			base[key] = additionalValue
			continue
		}

		baseSlice, baseIsSlice := baseValue.([]interface{})
		additionalSlice, additionalIsSlice := additionalValue.([]interface{})
		if baseIsSlice && additionalIsSlice {
			base[key] = append(additionalSlice, baseSlice...)
			continue
		}

		baseMap, baseIsMap := baseValue.(map[string]interface{})
		additionalMap, additionalIsMap := additionalValue.(map[string]interface{})
		if baseIsMap && additionalIsMap {
			mergeYAMLMaps(baseMap, additionalMap)
			continue
		}

		base[key] = additionalValue
	}
}

func (d *Driver) createNetworks() ([]*hcloud.Network, error) {
	networks := []*hcloud.Network{}
	for _, networkIDorName := range d.Networks {
		network, err := d.getClient().GetNetwork(context.Background(), networkIDorName)
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return instrumented(networks), nil
}

func (d *Driver) createFirewalls() ([]*hcloud.ServerCreateFirewall, error) {
	firewalls := []*hcloud.ServerCreateFirewall{}
	for _, firewallIDorName := range d.Firewalls {
		firewall, err := d.getClient().GetFirewall(context.Background(), firewallIDorName)
		if err != nil {
			return nil, err
		}
		firewalls = append(firewalls, &hcloud.ServerCreateFirewall{Firewall: *firewall})
	}
	return instrumented(firewalls), nil
}

func (d *Driver) createVolumes() ([]*hcloud.Volume, error) {
	volumes := []*hcloud.Volume{}
	for _, volumeIDorName := range d.Volumes {
		volume, err := d.getClient().GetVolume(context.Background(), volumeIDorName)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, volume)
	}
	return instrumented(volumes), nil
}
