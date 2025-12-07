package driver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/config"
	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/hetzner"
	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/logging"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Driver contains hetzner-specific data to implement [drivers.Driver]
type Driver struct {
	*drivers.BaseDriver

	AccessToken  string
	cachedClient *hetzner.Client
	Image             string
	ImageID           int64
	ImageArch         hcloud.Architecture
	cachedImage       *hcloud.Image
	Type              string
	cachedType        *hcloud.ServerType
	Location          string
	cachedLocation    *hcloud.Location
	KeyID             int64
	cachedKey         *hcloud.SSHKey
	IsExistingKey     bool
	originalKey       string
	dangling          []func()
	ServerID          int64
	cachedServer      *hcloud.Server
	userData           string
	userDataFile       string
	additionalUserData string
	Volumes           []string
	Networks          []string
	UsePrivateNetwork bool
	DisablePublic4    bool
	DisablePublic6    bool
	PrimaryIPv4       string
	cachedPrimaryIPv4 *hcloud.PrimaryIP
	PrimaryIPv6       string
	cachedPrimaryIPv6 *hcloud.PrimaryIP
	Firewalls         []string
	ServerLabels      map[string]string
	keyLabels         map[string]string
	placementGroup    string
	cachedPGrp        *hcloud.PlacementGroup

	AdditionalKeys       []string
	AdditionalKeyIDs     []int64
	cachedAdditionalKeys []*hcloud.SSHKey

	WaitOnError           int
	WaitOnPolling         int
	WaitForRunningTimeout int

	// internal housekeeping
	version string
	usesDfr bool
}

const (
	defaultImage = config.DefaultImage
	defaultType  = config.DefaultType

	flagAPIToken           = config.FlagAPIToken
	flagImage              = config.FlagImage
	flagImageID            = config.FlagImageID
	flagImageArch          = config.FlagImageArch
	flagType               = config.FlagType
	flagLocation           = config.FlagLocation
	flagExKeyID            = config.FlagExKeyID
	flagExKeyPath          = config.FlagExKeyPath
	flagUserData           = config.FlagUserData
	flagUserDataFile       = config.FlagUserDataFile
	flagAdditionalUserData = config.FlagAdditionalUserData
	flagVolumes            = config.FlagVolumes
	flagNetworks           = config.FlagNetworks
	flagUsePrivateNetwork  = config.FlagUsePrivateNetwork
	flagDisablePublic4     = config.FlagDisablePublic4
	flagDisablePublic6     = config.FlagDisablePublic6
	flagPrimary4           = config.FlagPrimary4
	flagPrimary6           = config.FlagPrimary6
	flagDisablePublic      = config.FlagDisablePublic
	flagFirewalls          = config.FlagFirewalls
	flagAdditionalKeys     = config.FlagAdditionalKeys
	flagServerLabel        = config.FlagServerLabel
	flagKeyLabel           = config.FlagKeyLabel
	flagPlacementGroup     = config.FlagPlacementGroup
	flagAutoSpread         = config.FlagAutoSpread

	flagSshUser = config.FlagSSHUser
	flagSshPort = config.FlagSSHPort

	defaultSSHPort = config.DefaultSSHPort
	defaultSSHUser = config.DefaultSSHUser

	flagWaitOnError              = config.FlagWaitOnError
	defaultWaitOnError           = config.DefaultWaitOnError
	flagWaitOnPolling            = config.FlagWaitOnPolling
	defaultWaitOnPolling         = config.DefaultWaitOnPolling
	flagWaitForRunningTimeout    = config.FlagWaitForRunning
	defaultWaitForRunningTimeout = config.DefaultWaitForRunningTimeout

	legacyFlagUserDataFromFile = config.LegacyFlagUserDataFromFile
	legacyFlagDisablePublic4   = config.LegacyFlagDisablePublic4
	legacyFlagDisablePublic6   = config.LegacyFlagDisablePublic6

	emptyImageArchitecture = config.EmptyImageArchitecture
)

// initializes a new driver instance;  [drivers.Driver.NewDriver]
func NewDriver(version string) *Driver {
	if runningInstrumented {
		instrumented("running instrument mode") // will be a no-op when not built with instrumentation
	}
	return &Driver{
		Type:          defaultType,
		IsExistingKey: false,
		BaseDriver:    &drivers.BaseDriver{},
		version:       version,
	}
}

func (d *Driver) DriverName() string {
	return "hetzner"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "HETZNER_API_TOKEN",
			Name:   flagAPIToken,
			Usage:  "Project-specific Hetzner API token",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_IMAGE",
			Name:   flagImage,
			Usage:  "Image to use for server creation",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_IMAGE_ID",
			Name:   flagImageID,
			Usage:  "Image to use for server creation",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_IMAGE_ARCH",
			Name:   flagImageArch,
			Usage:  "Image architecture for lookup to use for server creation",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_TYPE",
			Name:   flagType,
			Usage:  "Server type to create",
			Value:  defaultType,
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_LOCATION",
			Name:   flagLocation,
			Usage:  "Location to create machine at",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_EXISTING_KEY_ID",
			Name:   flagExKeyID,
			Usage:  "Existing key ID to use for server; requires --hetzner-existing-key-path",
			Value:  "0",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_EXISTING_KEY_PATH",
			Name:   flagExKeyPath,
			Usage:  "Path to existing key (new public key will be created unless --hetzner-existing-key-id is specified)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_USER_DATA",
			Name:   flagUserData,
			Usage:  "Cloud-init based user data (inline).",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_ADDITIONAL_USER_DATA",
			Name:   flagAdditionalUserData,
			Usage:  "Additional Cloud-init based user data (inline).",
			Value:  "",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_USER_DATA_FROM_FILE",
			Name:   legacyFlagUserDataFromFile,
			Usage:  "DEPRECATED, legacy.",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_USER_DATA_FILE",
			Name:   flagUserDataFile,
			Usage:  "Cloud-init based user data (read from file)",
			Value:  "",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_VOLUMES",
			Name:   flagVolumes,
			Usage:  "Volume IDs or names which should be attached to the server",
			Value:  []string{},
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_NETWORKS",
			Name:   flagNetworks,
			Usage:  "Network IDs or names which should be attached to the server private network interface",
			Value:  []string{},
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_USE_PRIVATE_NETWORK",
			Name:   flagUsePrivateNetwork,
			Usage:  "Use private network",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_DISABLE_PUBLIC_IPV4",
			Name:   flagDisablePublic4,
			Usage:  "Disable public ipv4",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_DISABLE_PUBLIC_4",
			Name:   legacyFlagDisablePublic4,
			Usage:  "DEPRECATED, legacy",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_DISABLE_PUBLIC_IPV6",
			Name:   flagDisablePublic6,
			Usage:  "Disable public ipv6",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_DISABLE_PUBLIC_6",
			Name:   legacyFlagDisablePublic6,
			Usage:  "DEPRECATED, legacy",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_DISABLE_PUBLIC",
			Name:   flagDisablePublic,
			Usage:  "Disable public ip (v4 & v6)",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_PRIMARY_IPV4",
			Name:   flagPrimary4,
			Usage:  "Existing primary IPv4 address",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_PRIMARY_IPV6",
			Name:   flagPrimary6,
			Usage:  "Existing primary IPv6 address",
			Value:  "",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_FIREWALLS",
			Name:   flagFirewalls,
			Usage:  "Firewall IDs or names which should be applied on the server",
			Value:  []string{},
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_ADDITIONAL_KEYS",
			Name:   flagAdditionalKeys,
			Usage:  "Additional public keys to be attached to the server",
			Value:  []string{},
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_SERVER_LABELS",
			Name:   flagServerLabel,
			Usage:  "Key value pairs of additional labels to assign to the server",
			Value:  []string{},
		},
		mcnflag.StringSliceFlag{
			EnvVar: "HETZNER_KEY_LABELS",
			Name:   flagKeyLabel,
			Usage:  "Key value pairs of additional labels to assign to the SSH key",
			Value:  []string{},
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_PLACEMENT_GROUP",
			Name:   flagPlacementGroup,
			Usage:  "Placement group ID or name to add the server to; will be created if it does not exist",
			Value:  "",
		},
		mcnflag.BoolFlag{
			EnvVar: "HETZNER_AUTO_SPREAD",
			Name:   flagAutoSpread,
			Usage:  "Auto-spread on a docker-machine-specific default placement group",
		},
		mcnflag.StringFlag{
			EnvVar: "HETZNER_SSH_USER",
			Name:   flagSshUser,
			Usage:  "SSH username",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "HETZNER_SSH_PORT",
			Name:   flagSshPort,
			Usage:  "SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.IntFlag{
			EnvVar: "HETZNER_WAIT_ON_ERROR",
			Name:   flagWaitOnError,
			Usage:  "Wait if an error happens while creating the server",
			Value:  defaultWaitOnError,
		},
		mcnflag.IntFlag{
			EnvVar: "HETZNER_WAIT_ON_POLLING",
			Name:   flagWaitOnPolling,
			Usage:  "Period for waiting between requests when waiting for some state to change",
			Value:  defaultWaitOnPolling,
		},
		mcnflag.IntFlag{
			EnvVar: "HETZNER_WAIT_FOR_RUNNING_TIMEOUT",
			Name:   flagWaitForRunningTimeout,
			Usage:  "Period for waiting for a machine to be running before failing",
			Value:  defaultWaitForRunningTimeout,
		},
	}
}

func flagI64(opts drivers.DriverOptions, key string) (int64, error) {
	raw := opts.String(key)
	if raw == "" {
		return 0, nil
	}

	ret, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse int64 for %v: %w", key, err)
	}

	return ret, nil
}


func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	return d.setConfigFromFlags(opts)
}

func (d *Driver) setConfigFromFlagsImpl(opts drivers.DriverOptions) error {
	var err error

	d.AccessToken = opts.String(flagAPIToken)
	d.Image = opts.String(flagImage)
	d.ImageID, err = flagI64(opts, flagImageID)
	if err != nil {
		return err
	}
	err = d.setImageArch(opts.String(flagImageArch))
	if err != nil {
		return err
	}
	d.Location = opts.String(flagLocation)
	d.Type = opts.String(flagType)
	d.KeyID, err = flagI64(opts, flagExKeyID)
	if err != nil {
		return err
	}
	d.IsExistingKey = d.KeyID != 0
	d.originalKey = opts.String(flagExKeyPath)
	err = d.setUserDataFlags(opts)
	if err != nil {
		return err
	}
	d.Volumes = opts.StringSlice(flagVolumes)
	d.Networks = opts.StringSlice(flagNetworks)
	disablePublic := opts.Bool(flagDisablePublic)
	d.UsePrivateNetwork = opts.Bool(flagUsePrivateNetwork) || disablePublic
	d.DisablePublic4 = d.deprecatedBooleanFlag(opts, flagDisablePublic4, legacyFlagDisablePublic4) || disablePublic
	d.DisablePublic6 = d.deprecatedBooleanFlag(opts, flagDisablePublic6, legacyFlagDisablePublic6) || disablePublic
	d.PrimaryIPv4 = opts.String(flagPrimary4)
	d.PrimaryIPv6 = opts.String(flagPrimary6)
	d.Firewalls = opts.StringSlice(flagFirewalls)
	d.AdditionalKeys = opts.StringSlice(flagAdditionalKeys)

	d.SSHUser = opts.String(flagSshUser)
	d.SSHPort = opts.Int(flagSshPort)

	d.WaitOnError = opts.Int(flagWaitOnError)
	d.WaitOnPolling = opts.Int(flagWaitOnPolling)
	d.WaitForRunningTimeout = opts.Int(flagWaitForRunningTimeout)

	d.placementGroup = opts.String(flagPlacementGroup)
	if opts.Bool(flagAutoSpread) {
		if d.placementGroup != "" {
			return d.flagFailure("%v and %v are mutually exclusive", flagAutoSpread, flagPlacementGroup)
		}
		d.placementGroup = config.AutoSpreadPGName
	}

	err = d.setLabelsFromFlags(opts)
	if err != nil {
		return err
	}

	d.SetSwarmConfigFromFlags(opts)

	if d.AccessToken == "" {
		return d.flagFailure("hetzner requires --%v to be set", flagAPIToken)
	}

	if err = d.verifyImageFlags(); err != nil {
		return err
	}

	if err = d.verifyNetworkFlags(); err != nil {
		return err
	}

	instrumented(d)

	if d.usesDfr {
		log.Warn("========== BREAKING CHANGE WARNING ==========")
		log.Warn("Your configuration uses deprecated flags that will be removed in v6")
		log.Warn("Check preceding output for 'DEPRECATED' warnings")
		log.Warn("==============================================")
	}

	return nil
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.SSHPort, nil
}

func (d *Driver) PreCreateCheck() error {
	if err := d.setupExistingKey(); err != nil {
		return err
	}

	if serverType, err := d.getType(); err != nil {
		return fmt.Errorf("could not get type: %w", err)
	} else if d.ImageArch != "" && serverType.Architecture != d.ImageArch {
		log.Warnf("Supplied architecture %v differs from server architecture %v", d.ImageArch, serverType.Architecture)
	}

	if _, err := d.getImage(); err != nil {
		return fmt.Errorf("could not get image: %w", err)
	}

	if _, err := d.getLocationNullable(); err != nil {
		return fmt.Errorf("could not get location: %w", err)
	}

	if _, err := d.getPlacementGroup(); err != nil {
		return fmt.Errorf("could not create placement group: %w", err)
	}

	if _, err := d.getPrimaryIPv4(); err != nil {
		return fmt.Errorf("could not resolve primary IPv4: %w", err)
	}

	if _, err := d.getPrimaryIPv6(); err != nil {
		return fmt.Errorf("could not resolve primary IPv6: %w", err)
	}

	if d.UsePrivateNetwork && len(d.Networks) == 0 {
		return fmt.Errorf("no private network attached")
	}

	return nil
}

func (d *Driver) Create() error {
	err := d.prepareLocalKey()
	if err != nil {
		return err
	}

	defer d.destroyDangling()
	err = d.createRemoteKeys()
	if err != nil {
		return err
	}

	log.Info("Creating Hetzner server...")

	srvopts, err := d.makeCreateServerOptions()
	if err != nil {
		return err
	}

	srv, err := d.getClient().CreateServer(context.Background(), instrumented(*srvopts))
	if err != nil {
		time.Sleep(time.Duration(d.WaitOnError) * time.Second)
		return err
	}

	logging.Step("Created %s, action: %s", logging.Server(srv.Server.Name, srv.Server.ID), logging.Action(srv.Action.Command, srv.Action.ID))
	if err = d.waitForAction(srv.Action); err != nil {
		return fmt.Errorf("could not wait for action: %w", err)
	}

	d.ServerID = srv.Server.ID
	logging.Step("Waiting for %s to start...", logging.Server(srv.Server.Name, srv.Server.ID))

	err = d.waitForInitialStartup(srv)
	if err != nil {
		return err
	}

	err = d.configureNetworkAccess(srv)
	if err != nil {
		return err
	}

	logging.Step("Server %s ready at %s", logging.Server(srv.Server.Name, srv.Server.ID), d.IPAddress)
	// Successful creation, so no keys dangle anymore
	d.dangling = nil

	return nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", fmt.Errorf("could not execute drivers.MustBeRunning: %w", err)
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", fmt.Errorf("could not get IP: %w", err)
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	srv, err := d.getClient().GetServerByID(context.Background(), d.ServerID)
	if err != nil {
		return state.None, err
	}
	if srv == nil {
		return state.None, errors.New("server not found")
	}

	switch srv.Status {
	case hcloud.ServerStatusInitializing:
		return state.Starting, nil
	case hcloud.ServerStatusRunning:
		return state.Running, nil
	case hcloud.ServerStatusOff:
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Remove() error {
	if err := d.destroyServer(); err != nil {
		return err
	}

	for i, id := range d.AdditionalKeyIDs {
		logging.Step("Destroying additional SSH key #%d [ID: %d]", i, id)
		key, softErr := d.getClient().GetSSHKeyByID(context.Background(), id)
		if softErr != nil {
			logging.WarnStep("Could not retrieve key: %v", softErr)
			continue
		}
		if key == nil {
			logging.WarnStep("Key [ID: %d] no longer exists", id)
			continue
		}

		softErr = d.getClient().DeleteSSHKey(context.Background(), key)
		if softErr != nil {
			logging.WarnStep("Could not remove key: %v", softErr)
		}
	}

	if !d.IsExistingKey && d.KeyID != 0 {
		key, err := d.getKeyNullable()
		if err != nil {
			return fmt.Errorf("could not get ssh key: %w", err)
		}
		if key == nil {
			logging.Step("SSH key no longer exists")
			return nil
		}

		logging.Step("Destroying SSH key %s", logging.Key(key.Name, key.ID))

		if err := d.getClient().DeleteSSHKey(context.Background(), key); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Restart() error {
	srv, err := d.getServerHandle()
	if err != nil {
		return fmt.Errorf("could not get server handle: %w", err)
	}
	if srv == nil {
		return errors.New("server not found")
	}

	act, err := d.getClient().RebootServer(context.Background(), srv)
	if err != nil {
		return err
	}

	logging.Step("Rebooting %s, action: %s", logging.Server(srv.Name, srv.ID), logging.Action(act.Command, act.ID))

	return d.waitForAction(act)
}

func (d *Driver) Start() error {
	srv, err := d.getServerHandle()
	if err != nil {
		return fmt.Errorf("could not get server handle: %w", err)
	}

	act, err := d.getClient().PowerOnServer(context.Background(), srv)
	if err != nil {
		return err
	}

	logging.Step("Starting %s, action: %s", logging.Server(srv.Name, srv.ID), logging.Action(act.Command, act.ID))

	return d.waitForAction(act)
}

func (d *Driver) Stop() error {
	srv, err := d.getServerHandle()
	if err != nil {
		return fmt.Errorf("could not get server handle: %w", err)
	}

	act, err := d.getClient().ShutdownServer(context.Background(), srv)
	if err != nil {
		return err
	}

	logging.Step("Shutting down %s, action: %s", logging.Server(srv.Name, srv.ID), logging.Action(act.Command, act.ID))

	return d.waitForAction(act)
}

func (d *Driver) Kill() error {
	srv, err := d.getServerHandle()
	if err != nil {
		return fmt.Errorf("could not get server handle: %w", err)
	}

	act, err := d.getClient().PowerOffServer(context.Background(), srv)
	if err != nil {
		return err
	}

	logging.Step("Powering off %s, action: %s", logging.Server(srv.Name, srv.ID), logging.Action(act.Command, act.ID))

	return d.waitForAction(act)
}
