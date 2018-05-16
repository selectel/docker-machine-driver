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

func (d *Driver) DriverName() string {
	return "selectel"
}

func (d *Driver) GetURL() (string, error) {
	panic("implement me")
}

func (d *Driver) GetState() (state.State, error) {
	// TODO (m.kalinin): require OS vars and get status of VM
	panic("implement me")

func (d *Driver) Start() error {
	return errors.New("selectel driver does not support start")
}

func (d *Driver) Stop() error {
	return errors.New("selectel driver does not support stop")
}

func (d *Driver) Save() error {
	return errors.New("selectel driver does not support save")
}

func (d *Driver) Kill() error {
	return errors.New("selectel driver does not support kill")
}

func (d *Driver) Remove() error {
	return errors.New("selectel driver does not support remove")
}

func (d *Driver) Restart() error {
	return errors.New("selectel driver does not support restart")
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.ResellAPIToken = opts.String("token")
	return checkDriver(d)
}

}

}
