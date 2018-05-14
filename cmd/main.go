package main

import (
	"github.com/selectel/docker-machine-driver/driver"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}