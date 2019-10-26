package config

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	RemoteFileConfig `mapstructure:",squash"`
	ImageConfig      `mapstructure:",squash"`

	ctx interpolate.Context
}

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

	return warnings, errors
}
