package openstack

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/version"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	cmp_fips "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
)

type Client interface {
	CreateVolume(opts volumes.CreateOpts) (*volumes.Volume, error)
	DeleteVolume(volumeID string) error
	WaitForVolumeStatus(volumeID, status string) error

	BootInstanceFromVolume(opts servers.CreateOptsBuilder) (*servers.Server, error)
	DeleteServer(serverID string) error
	SetServerPassword(serverID string, password string) error
	GetServerState(serverID string) (string, error)

	AttachFloatingIP(serverID, floatingIP string) error
	GetAllFloatingIP() ([]floatingips.FloatingIP, error)
}

type GenericClient struct {
	Provider     *gophercloud.ServiceClient
	Compute      *gophercloud.ServiceClient
	BlockStorage *gophercloud.ServiceClient
	Network      *gophercloud.ServiceClient
}

func NewClient(opts gophercloud.AuthOptions, region string) (Client, error) {
	endpointOpts := gophercloud.EndpointOpts{
		Region: region,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	blockStorageClient, err := openstack.NewBlockStorageV2(provider, endpointOpts)
	if err != nil {
		return nil, err
	}

	computeClient, err := openstack.NewComputeV2(provider, endpointOpts)
	if err != nil {
		return nil, err
	}

	networkClient, err := openstack.NewNetworkV2(provider, endpointOpts)
	if err != nil {
		return nil, err
	}

	provider.UserAgent.Prepend(fmt.Sprintf("docker-machine/v%d", version.APIVersion))
	return &GenericClient{
		Compute:      computeClient,
		BlockStorage: blockStorageClient,
		Network:      networkClient,
	}, nil
}

func (client *GenericClient) CreateVolume(opts volumes.CreateOpts) (*volumes.Volume, error) {
	return volumes.Create(client.BlockStorage, opts).Extract()
}

func (client *GenericClient) DeleteVolume(volumeID string) error {
	return volumes.Delete(client.BlockStorage, volumeID).Err
}

func (client *GenericClient) WaitForVolumeStatus(volumeID, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		current, err := volumes.Get(client.BlockStorage, volumeID).Extract()
		if err != nil {
			return true, err
		}

		if strings.ToLower(current.Status) == status {
			return true, nil
		}

		return false, nil
	}, 10, time.Second)
}

func (client *GenericClient) BootInstanceFromVolume(opts servers.CreateOptsBuilder) (*servers.Server, error) {
	return bootfromvolume.Create(client.Compute, opts).Extract()
}

func (client *GenericClient) DeleteServer(serverID string) error {
	return servers.Delete(client.Compute, serverID).Err
}

func (client *GenericClient) GetServerState(serverID string) (string, error) {
	server, err := servers.Get(client.Compute, serverID).Extract()
	if err != nil {
		return "", err
	}
	return server.Status, err
}

func (client *GenericClient) SetServerPassword(serverID string, password string) error {
	return servers.ChangeAdminPassword(client.Compute, serverID, password).ExtractErr()
}

func (client *GenericClient) AttachFloatingIP(serverID, floatingIP string) error {
	opts := cmp_fips.AssociateOpts{
		FloatingIP: floatingIP,
	}
	return cmp_fips.AssociateInstance(client.Compute, serverID, opts).Err
}

func (client *GenericClient) GetAllFloatingIP() ([]floatingips.FloatingIP, error) {
	opts := floatingips.ListOpts{
		Status: "down",
	}

	allPages, err := floatingips.List(client.Network, opts).AllPages()
	if err != nil {
		return nil, err
	}

	return floatingips.ExtractFloatingIPs(allPages)
}
