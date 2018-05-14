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
	// TODO (m.kalinin): require OS vars and get status of VM
	panic("implement me")
}

func (d *Driver) Kill() error {
	// TODO (m.kalinin): if --remove-project and other Resell API vars are provided then remove project
	// TODO (m.kalinin): if OS vars are provided then just remove a VM
	panic("implement me")
}

func (d *Driver) Remove() error {
	panic("implement me")
}

func (d *Driver) Restart() error {
	// TODO (m.kalinin): require OS vars and power-on provided VM
	panic("implement me")
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.ResellAPIToken = opts.String("token")
	return checkDriver(d)
}

func (d *Driver) Start() error {
	// TODO (m.kalinin): require OS vars and power-on provided VM
	panic("implement me")
}

func (d *Driver) Stop() error {
	// TODO (m.kalinin): require OS vars and shutdown provided VM
	panic("implement me")
}
