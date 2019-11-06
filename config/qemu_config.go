package config

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type QemuConfig struct {
	QemuBinarySourcePath      string   `mapstructure:"qemu_binary_source_path"`
	QemuBinaryDestinationPath string   `mapstructure:"qemu_binary_destination_path"`
	QemuArgs                  []string `mapstructure:"qemu_args"`
}

func (c *QemuConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	return warnings, errs
}
