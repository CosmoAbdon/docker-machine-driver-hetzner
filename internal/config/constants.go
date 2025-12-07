package config

import "github.com/hetznercloud/hcloud-go/v2/hcloud"
import "slices"

const (
	DefaultImage = "ubuntu-24.04"
	DefaultType  = "cpx22"

	DefaultSSHPort = 22
	DefaultSSHUser = "root"

	DefaultWaitOnError           = 0
	DefaultWaitOnPolling         = 1
	DefaultWaitForRunningTimeout = 0
)

const (
	FlagAPIToken           = "hetzner-api-token"
	FlagImage              = "hetzner-image"
	FlagImageID            = "hetzner-image-id"
	FlagImageArch          = "hetzner-image-arch"
	FlagType               = "hetzner-server-type"
	FlagLocation           = "hetzner-server-location"
	FlagExKeyID            = "hetzner-existing-key-id"
	FlagExKeyPath          = "hetzner-existing-key-path"
	FlagUserData           = "hetzner-user-data"
	FlagUserDataFile       = "hetzner-user-data-file"
	FlagAdditionalUserData = "hetzner-additional-user-data"
	FlagVolumes            = "hetzner-volumes"
	FlagNetworks           = "hetzner-networks"
	FlagUsePrivateNetwork  = "hetzner-use-private-network"
	FlagDisablePublic4     = "hetzner-disable-public-ipv4"
	FlagDisablePublic6     = "hetzner-disable-public-ipv6"
	FlagPrimary4           = "hetzner-primary-ipv4"
	FlagPrimary6           = "hetzner-primary-ipv6"
	FlagDisablePublic      = "hetzner-disable-public"
	FlagFirewalls          = "hetzner-firewalls"
	FlagAdditionalKeys     = "hetzner-additional-key"
	FlagServerLabel        = "hetzner-server-label"
	FlagKeyLabel           = "hetzner-key-label"
	FlagPlacementGroup     = "hetzner-placement-group"
	FlagAutoSpread         = "hetzner-auto-spread"
	FlagSSHUser            = "hetzner-ssh-user"
	FlagSSHPort            = "hetzner-ssh-port"
	FlagWaitOnError        = "hetzner-wait-on-error"
	FlagWaitOnPolling      = "hetzner-wait-on-polling"
	FlagWaitForRunning     = "hetzner-wait-for-running-timeout"

	LegacyFlagUserDataFromFile = "hetzner-user-data-from-file"
	LegacyFlagDisablePublic4   = "hetzner-disable-public-4"
	LegacyFlagDisablePublic6   = "hetzner-disable-public-6"
)

const (
	LabelAutoSpreadPG = "auto-spread"
	LabelAutoCreated  = "auto-created"
	AutoSpreadPGName  = "__auto_spread"
)

const EmptyImageArchitecture = hcloud.Architecture("")

var LegacyDefaultImages = []string{
	DefaultImage,
	"ubuntu-18.04",
	"ubuntu-16.04",
	"debian-9",
}

func IsDefaultImageName(imageName string) bool {
	return slices.Contains(LegacyDefaultImages, imageName)
}
