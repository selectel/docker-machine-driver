package openstack

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/version"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	cmp_fips "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
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
	AttachFirstFreeFloatingIP(serverID string) (string, error)
	GetAllFloatingIP() ([]floatingips.FloatingIP, error)

	GetPublicKey(keyPairName string) ([]byte, error)
	CreateKeyPair(name string, publicKey string) error
	DeleteKeyPair(name string) error

	GetFlavorBy(name, id *string) (*flavors.Flavor, error)
	CreateFlavor(name string, cpu, ram int) (*flavors.Flavor, error)

	GetImageBy(name, id *string) (*images.Image, error)
}

type GenericClient struct {
	Provider     *gophercloud.ServiceClient
	Compute      *gophercloud.ServiceClient
	BlockStorage *gophercloud.ServiceClient
	Network      *gophercloud.ServiceClient
	Image        *gophercloud.ServiceClient
}

type ClientOpts struct {
	Credentials  gophercloud.AuthOptions
	EndpointOpts gophercloud.EndpointOpts
	Proxy        *url.URL
}

const (
	UserAgent = "docker-machine/v%d"
)

func NewClient(opts ClientOpts) (Client, error) {
	provider, err := openstack.AuthenticatedClient(opts.Credentials)
	if err != nil {
		return nil, err
	}

	blockStorageClient, err := openstack.NewBlockStorageV2(provider, opts.EndpointOpts)
	if err != nil {
		return nil, err
	}

	computeClient, err := openstack.NewComputeV2(provider, opts.EndpointOpts)
	if err != nil {
		return nil, err
	}

	networkClient, err := openstack.NewNetworkV2(provider, opts.EndpointOpts)
	if err != nil {
		return nil, err
	}

	imageClient, err := openstack.NewImageServiceV2(provider, opts.EndpointOpts)
	if err != nil {
		return nil, err
	}

	provider.HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(opts.Proxy)}
	blockStorageClient.HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(opts.Proxy)}
	computeClient.HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(opts.Proxy)}
	networkClient.HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(opts.Proxy)}
	imageClient.HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(opts.Proxy)}

	provider.UserAgent.Prepend(fmt.Sprintf(UserAgent, version.APIVersion))
	return &GenericClient{
		Compute:      computeClient,
		BlockStorage: blockStorageClient,
		Network:      networkClient,
		Image:        imageClient,
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

func (client *GenericClient) AttachFirstFreeFloatingIP(serverID string) (string, error) {
	fips, err := client.GetAllFloatingIP()
	if err != nil {
		return "", err
	}
	serverIP := fips[0].FloatingIP

	if err := client.AttachFloatingIP(serverID, serverIP); err != nil {
		return "", err
	}
	return serverIP, nil
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

func (client *GenericClient) GetPublicKey(keyPairName string) ([]byte, error) {
	keyPair, err := keypairs.Get(client.Compute, keyPairName).Extract()
	if err != nil {
		return nil, err
	}
	return []byte(keyPair.PublicKey), nil
}

func (client *GenericClient) CreateKeyPair(name string, publicKey string) error {
	opts := keypairs.CreateOpts{
		Name:      name,
		PublicKey: publicKey,
	}
	return keypairs.Create(client.Compute, opts).Err
}

func (client *GenericClient) DeleteKeyPair(name string) error {
	return keypairs.Delete(client.Compute, name).Err

}

func (client *GenericClient) CreateFlavor(name string, cpu, ram int) (*flavors.Flavor, error) {
	// TODO (m.kalinin): extract for using local_gb
	diskValue := 0
	isPublic := false
	opts := flavors.CreateOpts{
		Name:     name,
		RAM:      ram,
		VCPUs:    cpu,
		IsPublic: &isPublic,
		Disk:     &diskValue,
	}
	return flavors.Create(client.Compute, opts).Extract()
}

func GetFlavorByName(client *GenericClient, name string) (*flavors.Flavor, error) {
	flavorID, err := flavors.IDFromName(client.Compute, name)
	if err != nil {
		return nil, err
	}

	flavor := flavors.Flavor{
		Name: name,
		ID:   flavorID,
	}
	return &flavor, nil
}

func (client *GenericClient) GetFlavorBy(name, id *string) (*flavors.Flavor, error) {
	if name == nil && id == nil {
		return nil, errors.New("flavor name and flavor id can't be null")
	}

	if name != nil {
		return GetFlavorByName(client, *name)
	}
	return flavors.Get(client.Compute, *id).Extract()
}

func GetImageByName(client *GenericClient, name string) (*images.Image, error) {
	imageID, err := images.IDFromName(client.Compute, name)
	if err != nil {
		return nil, err
	}

	image := images.Image{
		Name: name,
		ID:   imageID,
	}
	return &image, nil
}

func (client *GenericClient) GetImageBy(name, id *string) (*images.Image, error) {
	if name == nil && id == nil {
		return nil, errors.New("image name and image id can't be null")
	}

	if name != nil {
		return GetImageByName(client, *name)
	}
	return images.Get(client.Image, *id).Extract()
}
