package openstack

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type Client interface {
	CreateVolume(opts volumes.CreateOpts) (*volumes.Volume, error)
	DeleteVolume(volumeID string) error
	WaitForVolumeStatus(volumeID, status string) error

	BootInstanceFromVolume(opts servers.CreateOptsBuilder) (*servers.Server, error)
	DeleteServer(serverID string) error
	WaitForServerStatus(serverID, status string) error

	AttachFloatingIP(serverID, floatingIP string) error
}

type GenericClient struct {
	Provider     *gophercloud.ServiceClient
	Compute      *gophercloud.ServiceClient
	BlockStorage *gophercloud.ServiceClient
}

const (
	ComputeEndpointType = "compute"
	VolumeEndpointType  = "volumev2"
)

func NewClient(opts gophercloud.AuthOptions, region string) (Client, error) {
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	blockStorageClient, err := openstack.NewBlockStorageV2(provider, gophercloud.EndpointOpts{
		Region: region,
		Type:   VolumeEndpointType,
	})
	if err != nil {
		return nil, err
	}

	computeClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
		Type:   ComputeEndpointType,
	})
	if err != nil {
		return nil, err
	}

	return &GenericClient{
		Compute:      computeClient,
		BlockStorage: blockStorageClient,
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

func (client *GenericClient) WaitForServerStatus(serverID, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		server, err := servers.Get(client.Compute, serverID).Extract()
		if err != nil {
			return true, err
		}

		if strings.ToLower(server.Status) == "error" {
			return true, fmt.Errorf("Instance creation failed. Instance is in ERROR state")
		}

		if strings.ToLower(server.Status) == status {
			return true, nil
		}

		return false, nil
	}, 10, 4*time.Second)
}

func (client *GenericClient) AttachFloatingIP(serverID, floatingIP string) error {
	opts := floatingips.AssociateOpts{
		FloatingIP: floatingIP,
	}
	return floatingips.AssociateInstance(client.Compute, serverID, opts).Err
}
