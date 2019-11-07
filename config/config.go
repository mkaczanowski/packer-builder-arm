package config

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// Config top-level holder for more specific configurations used
// while building packer-builder-arm
type Config struct {
	RemoteFileConfig `mapstructure:",squash"`
	ImageConfig      `mapstructure:",squash"`
	QemuConfig       `mapstructure:",squash"`

	Ctx interpolate.Context
}

// Prepare relevant config structures
func (c *Config) Prepare(ctx *interpolate.Context) (warnings []string, errors []error) {
	var (
		warns []string
		errs  []error
	)

	warns, errs = c.RemoteFileConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	warns, errs = c.ImageConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	warns, errs = c.QemuConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	return warnings, errors
}
