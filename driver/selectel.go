package driver

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/selectel/docker-machine-driver/openstack"
)

const (
	// ssh
	defaultSSHUser        = "root"
	defaultSSHPort        = 22
	defaultKeyPairName    = "docker-machine-key"

	// volume
	defaultVolumeName = "volume for %s"
	defaultVolumeType = "fast.%s"
	defaultVolumeSize = 5

	// flavor
	defaultCPUValue = 1
	defaultRAMValue = 512

	// other
	defaultImage = "Ubuntu 16.04 LTS 64-bit"
)

type Driver struct {
	*drivers.BaseDriver
	client           openstack.Client
	AuthUrl          string
	DomainName       string
	Username         string
	Password         string
	ProjectID        string
	Region           string
	AvailabilityZone string
	ServerID         string
	VolumeID         string
	Proxy            string
	RAM              int
	CPU              int
	SSHKeyName       string
	SSHPublicKeyPath string
	ServerName       string
	VolumeName       string
	VolumeSize       int
	VolumeType       string
	FlavorName       string
	FlavorID         string
	ImageName        string
	ImageID          string
	NetworkID        string
}

func NewDriver(hostName string, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		// openstack variables
		mcnflag.StringFlag{
			EnvVar: "OS_AUTH_URL",
			Name:   "os-auth-url",
			Usage:  "OpenStack authentication URL",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_REGION_NAME",
			Name:   "os-region",
			Usage:  "OpenStack region name",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_AVAILABILITY_ZONE",
			Name:   "os-availability-zone",
			Usage:  "OpenStack availability zone",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PROJECT_DOMAIN_NAME",
			Name:   "os-domain-name",
			Usage:  "OpenStack domain name (identity v3 only)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_USERNAME",
			Name:   "os-username",
			Usage:  "OpenStack username",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PASSWORD",
			Name:   "os-password",
			Usage:  "OpenStack user password",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PROJECT_ID",
			Name:   "os-project-id",
			Usage:  "OpenStack project id",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLAVOR_ID",
			Name:   "os-flavor-id",
			Usage:  "OpenStack flavor id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLAVOR_NAME",
			Name:   "os-flavor-name",
			Usage:  "OpenStack flavor name to use for the instance",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_IMAGE_NAME",
			Name:   "os-image-name",
			Usage:  "OpenStack flavor name to use for the instance",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_IMAGE_ID",
			Name:   "os-image-id",
			Usage:  "OpenStack image id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_NETWORK_NAME",
			Name:   "os-net-name",
			Usage:  "OpenStack network name the machine will be connected on",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_NETWORK_ID",
			Name:   "os-net-id",
			Usage:  "OpenStack network id the machine will be connected on",
			Value:  "",
		},

		// ssh-key variables
		mcnflag.StringFlag{
			EnvVar: "SEL_SSH_USER",
			Name:   "sel-ssh-user",
			Usage:  "SSH user for connecting to the server",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "SEL_SSH_PORT",
			Name:   "sel-ssh-port",
			Usage:  "SSH port for connecting to the server",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "SEL_SSH_PAIR_NAME",
			Name:   "sel-ssh-pair-name",
			Usage:  "Existing keypair name",
			Value:  defaultKeyPairName,
		},
		mcnflag.StringFlag{
			EnvVar: "SEL_SSH_PRIVATE_KEY_PATH",
			Name:   "sel-ssh-private-key-path",
			Usage:  "Private keyfile to use for SSH (absolute path)",
		},

		// volume variables
		mcnflag.StringFlag{
			EnvVar: "SEL_VOLUME_NAME",
			Name:   "sel-volume-name",
			Usage:  "Name of the server volume",
		},
		mcnflag.StringFlag{
			EnvVar: "SEL_VOLUME_TYPE",
			Name:   "sel-volume-type",
			Usage:  "Base volume type for server",
		},
		mcnflag.IntFlag{
			EnvVar: "SEL_VOLUME_SIZE",
			Name:   "sel-volume-size",
			Usage:  "Volume size",
			Value:  defaultVolumeSize,
		},

		// other variables
		mcnflag.StringFlag{
			EnvVar: "SEL_SERVER_NAME",
			Name:   "sel-server-name",
			Usage:  "Name of future server",
		},
		mcnflag.StringFlag{
			EnvVar: "SEL_PROXY",
			Name:   "sel-proxy",
			Usage:  "Proxy for the OS services",
		},
		mcnflag.IntFlag{
			EnvVar: "SEL_CPU_VALUE",
			Name:   "sel-cpu",
			Usage:  "Count of vCPU for server",
			Value:  defaultCPUValue,
		},
		mcnflag.IntFlag{
			EnvVar: "SEL_RAM_VALUE",
			Name:   "sel-ram",
			Usage:  "Count of RAM for server",
			Value:  defaultRAMValue,
		},
	}
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	// openstack
	d.ServerName = opts.String("sel-server-name")
	d.FlavorID = opts.String("os-flavor-id")
	d.FlavorName = opts.String("os-flavor-name")
	d.ImageID = opts.String("os-image-id")
	d.ImageName = opts.String("os-image-name")
	d.NetworkID = opts.String("os-net-id")
	d.ServerName = opts.String("sel-server-name")
	d.AuthUrl = opts.String("os-auth-url")
	d.Username = opts.String("os-username")
	d.Password = opts.String("os-password")
	d.DomainName = opts.String("os-domain-name")
	d.Region = opts.String("os-region")
	d.ProjectID = opts.String("os-project-id")
	d.AvailabilityZone = opts.String("os-availability-zone")

	// ssh
	d.SSHKeyName = opts.String("sel-ssh-pair-name")
	d.SSHKeyPath = opts.String("sel-ssh-private-key-path")
	d.SSHPublicKeyPath = fmt.Sprintf("%s.pub", d.SSHKeyPath)

	// selectel
	d.RAM = opts.Int("sel-ram")
	d.CPU = opts.Int("sel-cpu")

	// volumes
	d.VolumeSize = opts.Int("sel-volume-size")
	d.VolumeName = opts.String("sel-volume-name")
	d.VolumeType = opts.String("sel-volume-type")

	// other
	d.Proxy = opts.String("sel-proxy")

	// replace variables if needed
	if len(d.ServerName) == 0 {
		d.ServerName = d.GetMachineName()
	}
	if len(d.VolumeName) == 0 {
		d.VolumeName = fmt.Sprintf(defaultVolumeName, d.ServerName)
	}
	if len(d.VolumeType) == 0 {
		d.VolumeType = fmt.Sprintf(defaultVolumeType, d.AvailabilityZone)
	}
	return d.checkConfig()
}

func (d *Driver) Start() error {
	d.MustAuthenticateIfNeeded()
	return d.client.StartServer(d.ServerID)
}

func (d *Driver) Stop() error {
	d.MustAuthenticateIfNeeded()
	return d.client.StopServer(d.ServerID)
}

func (d *Driver) Save() error {
	return errors.New("selectel driver does not support save")
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() (err error) {
	d.MustAuthenticateIfNeeded()
	log.Infof("Removing server with id '%s'...", d.ServerID)
	if err := d.client.RemoveServer(d.ServerID); err != nil {
		log.Error(err)
	}

	// wait when we may remove volume
	d.client.WaitForVolumeStatus(d.VolumeID, "available")

	log.Infof("Removing volume with id '%s'...", d.VolumeID)
	if err := d.client.DeleteVolume(d.VolumeID); err != nil {
		log.Error(err)
	}

	log.Info("Removing flavor...")
	if err := d.client.DeleteVolume(d.FlavorID); err != nil {
		log.Errorf("Can't remove flavor with id '%s'. Is it public?", d.FlavorID)
	}

	log.Info("Removing ssh-key...")
	if err := d.client.DeleteKeyPair(d.SSHKeyName); err != nil {
		log.Errorf("Can't remove ssh-key with name '%s'", d.SSHKeyName)
	}
	return
}

func (d *Driver) Restart() error {
	d.MustAuthenticateIfNeeded()
	return d.client.RestartServer(d.ServerID)
}

func (d *Driver) PreCreateCheck() (err error) {
	if err := d.Authenticate(); err != nil {
		return err
	}

	if err := requireFreeFloatingIP(d.client); err != nil {
		return err
	}

	if err := d.resolveNamesAndIds(); err != nil {
		return err
	}

	if err := createPublicKeyIfNeeded(d.client, d.SSHKeyName, d.SSHPublicKeyPath); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Create() error {
	volumeOpts := volumes.CreateOpts{
		Name:             d.VolumeName,
		VolumeType:       d.VolumeType,
		Size:             d.VolumeSize,
		ImageID:          d.ImageID,
		AvailabilityZone: d.AvailabilityZone,
	}
	volume, err := d.client.CreateVolume(volumeOpts)
	if err != nil {
		return err
	}

	d.VolumeID = volume.ID

	log.Info("Volume created", d.VolumeID)
	log.Info("Waiting volume AVAILABLE status...")
	d.client.WaitForVolumeStatus(volume.ID, "available")

	serverOpts := servers.CreateOpts{
		AvailabilityZone: d.AvailabilityZone,
		Name:             d.ServerName,
		FlavorRef:        d.FlavorID,
		Metadata: map[string]string{
			"x_sel_server_password_hash": fmt.Sprintf("$6$%s", "server_password_hash"),
		},
		Networks: []servers.Network{
			{
				UUID: d.NetworkID,
			},
		},
	}
	serverVolumeOpts := bootfromvolume.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		BlockDevice: []bootfromvolume.BlockDevice{
			{
				BootIndex:       0,
				UUID:            volume.ID,
				SourceType:      "volume",
				DestinationType: "volume",
			},
		},
	}
	serverKeyPairOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverVolumeOpts,
		KeyName:           d.SSHKeyName,
	}

	server, err := d.client.BootInstanceFromVolume(serverKeyPairOpts)
	if err != nil {
		return err
	}
	log.Info("Booted server from volume. ID:", server.ID)
	d.ServerID = server.ID
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}

	var err error
	log.Debug("Trying to attach floating ip...")
	if d.IPAddress, err = d.client.AttachFirstFreeFloatingIP(d.ServerID); err != nil {
		return "", err
	}
	log.Info("Successfully attached first available IP", d.IPAddress)
	return d.IPAddress, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) DriverName() string {
	return "selectel"
}

func (d *Driver) Authenticate() (err error) {
	opts := openstack.ClientOpts{
		Credentials: gophercloud.AuthOptions{
			IdentityEndpoint: d.AuthUrl,
			Username:         d.Username,
			Password:         d.Password,
			DomainName:       d.DomainName,
			TenantID:         d.ProjectID,
		},
		EndpointOpts: gophercloud.EndpointOpts{
			Region: d.Region,
		},
	}
	if len(d.Proxy) > 0 {
		proxy, err := url.Parse(d.Proxy)
		if err != nil {
			return err
		}

		opts.Proxy = proxy
	}
	d.client, err = openstack.NewClient(opts)
	return err
}

func (d *Driver) MustAuthenticateIfNeeded() {
	if d.client != nil {
		return
	}

	if err := d.Authenticate(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func (d *Driver) GetState() (state.State, error) {
	d.MustAuthenticateIfNeeded()
	status, err := d.client.GetServerState(d.ServerID)
	if err != nil {
		return state.None, err
	}

	switch status {
	case "ACTIVE":
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case "SUSPENDED":
		return state.Saved, nil
	case "SHUTOFF":
		return state.Stopped, nil
	case "BUILD":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	}
	log.Warnf("Found new server status '%s'", status)
	return state.None, nil
}
