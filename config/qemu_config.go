//go:generate mapstructure-to-hcl2 -type QemuConfig
package config

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// QemuConfig describes qemu configuration
type QemuConfig struct {
	QemuBinarySourcePath      string `mapstructure:"qemu_binary_source_path" required:"true"`
	QemuBinaryDestinationPath string `mapstructure:"qemu_binary_destination_path"`
}

// Prepare qemu configuration
func (q *QemuConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	// those paths are usually the same
	if q.QemuBinaryDestinationPath == "" {
		q.QemuBinaryDestinationPath = q.QemuBinarySourcePath
	}

	return warnings, errs
}
