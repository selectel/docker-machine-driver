package openstack

import (
	"strings"
	"time"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
)

type Client interface {
	CreateVolume(opts volumes.CreateOpts) (*volumes.Volume, error)
	DeleteVolume(volumeID string) error
	WaitForVolumeStatus(volumeID, status string) error
}

type GenericClient struct {
	Provider     *gophercloud.ServiceClient
	BlockStorage *gophercloud.ServiceClient
}

const (
	VolumeEndpointType = "volumev2"
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

	return &GenericClient{
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
