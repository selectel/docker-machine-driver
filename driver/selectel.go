package driver

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	ResellAPIToken string
}

const (
	driverName = "selectel"
)

func NewDriver(machineName string, storePath string) *Driver {
	return &Driver{}
}

func (d *Driver) Create() error {
	return nil
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "SEL_TOKEN",
			Name:   "token",
			Usage:  "Token for VPC Resell API",
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	panic("implement me")
}

func (d *Driver) GetState() (state.State, error) {
	panic("implement me")
}

func (d *Driver) Kill() error {
	panic("implement me")
}

func (d *Driver) Remove() error {
	panic("implement me")
}

func (d *Driver) Restart() error {
	panic("implement me")
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.ResellAPIToken = opts.String("token")
	return checkDriver(d)
}

func (d *Driver) Start() error {
	panic("implement me")
}

func (d *Driver) Stop() error {
	panic("implement me")
}
