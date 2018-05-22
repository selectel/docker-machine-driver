package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/selectel/docker-machine-driver/openstack"
)

const (
	errorMandatoryEnvOrOption = "%s must be specified either using the environment variable %s or the CLI option %s"
	errorMandatoryOption      = "%s must be specified using the CLI option %s"
	errorExclusiveOptions     = "Either %s or %s must be specified, not both"
)

func requireFreeFloatingIP(client openstack.Client) error {
	fips, err := client.GetAllFloatingIP()
	if err != nil {
		panic(err)
	}

	if len(fips) < 1 {
		return errors.New("no free floating ip in project")
	}
	return nil
}

func createPublicKeyIfNeeded(client openstack.Client, keyName, keyPath string) error {
	_, err := client.GetPublicKey(keyName)
	// user has a ssh-key pair with given name
	if err == nil {
		return nil
	}

	log.Infof("No ssh-key with name '%s' exists", keyName)
	publicKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	// create a new ssh-key for user with local ssh-key
	log.Info("Trying to add a new ssh-key for user...")
	return client.CreateKeyPair(keyName, string(publicKey))
}

func (d *Driver) checkConfig() error {
	if d.AuthUrl == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Authentication URL", "OS_AUTH_URL", "--os-auth-url")
	}
	if d.DomainName == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Domain name", "OS_PROJECT_DOMAIN_NAME", "--os-domain-name")
	}
	if d.Username == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Username", "OS_USERNAME", "--os-username")
	}
	if d.Password == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Password", "OS_PASSWORD", "--os-password")
	}
	if d.ProjectID == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Project id", "OS_PROJECT_ID", "--os-project-id")
	}
	if d.AvailabilityZone == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Availability Zone", "OS_AVAILABILITY_ZONE", "--os-availability-zone")
	}

	if d.FlavorName != "" && d.FlavorID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Flavor name", "Flavor id")
	}

	if d.ImageName == "" && d.ImageID == "" {
		return fmt.Errorf(errorMandatoryOption, "Image name or Image id", "--os-image-name or --os-image-id")
	}
	if d.ImageName != "" && d.ImageID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Image name", "Image id")
	}
	if d.NetworkID == "" {
		return fmt.Errorf(errorMandatoryOption, "Network id", "--os-net-id")
	}
	if _, err := os.Stat(d.SSHKeyPath); err != nil {
		return fmt.Errorf(errorMandatoryEnvOrOption, "KeyPairPath", "SEL_SSH_PRIVATE_KEY_PATH", "--sel-ssh-private-key-path")
	}
	return nil
}

func (d *Driver) resolveNamesAndIds() error {
	if d.FlavorID != "" {
		log.Info("FlavorID was provided. Validating...")
		if _, err := d.client.GetFlavorBy(nil, &d.FlavorID); err != nil {
			return err
		}

		// avoid next check
		d.FlavorName = ""
	}

	if d.FlavorName != "" {
		log.Infof("FlavorName was provided. Getting ID bases on '%s' name...", d.FlavorName)
		flavor, err := d.client.GetFlavorBy(&d.FlavorName, nil)
		if err != nil {
			return err
		}

		d.FlavorID = flavor.ID
		log.Info("Got flavorID", d.FlavorID)
	}

	if d.FlavorID == "" && d.FlavorName == "" {
		// todo: is RAM % 2 ?
		d.FlavorName = mcnutils.GenerateRandomID()[0:31]
		log.Info("No any information about flavor was provided.")
		log.Infof("Creating flavor with CPU/RAM values %d/%d and name %s", d.CPU, d.RAM, d.FlavorName)

		flavor, err := d.client.CreateFlavor(d.FlavorName, d.CPU, d.RAM)
		if err != nil {
			return err
		}

		d.FlavorID = flavor.ID
	}

	if d.ImageName != "" {
		log.Infof("ImageName was provided. Getting ID bases on '%s'", d.ImageName)
		image, err := d.client.GetImageBy(&d.ImageName, nil)
		if err != nil {
			return err
		}

		d.ImageID = image.ID
		log.Info("Got id", d.ImageID)
	}
	return nil
}
