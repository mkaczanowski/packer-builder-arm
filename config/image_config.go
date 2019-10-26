package config

import (
	"errors"

	humanize "github.com/dustin/go-humanize"
	"github.com/hashicorp/packer/template/interpolate"
)

type ImageConfig struct {
	// where image is going to be saved
	ImagePath string `mapstructure:"image_path" required:"true"`
	// An URL to a checksum file containing a checksum for the ISO file. At
	// least one of `file+checksum` and `file+checksum_url` must be defined.
	// `file+checksum_url` will be ignored if `file+checksum` is non empty.
	ImageSize      string `mapstructure:"image_size"`
	ImageSizeBytes uint64 `mapstructure:"image_size_bytes"`
}

func (c *ImageConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	if c.ImageSize != "" && c.ImageSizeBytes != 0 {
		errs = append(errs, errors.New("Only one of image_size or image_size_bytes can be specified"))
	}

	if c.ImageSize != "" {
		if got, err := humanize.ParseBytes(c.ImageSize); err != nil {
			errs = append(errs, err)
		} else {
			c.ImageSizeBytes = got
		}
	}

	if c.ImageSizeBytes == 0 {
		errs = append(errs, errors.New("One of image_size_bytes or image_size must be set"))
	}

	return warnings, errs
}
