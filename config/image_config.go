package config

import (
	"errors"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/hashicorp/packer/template/interpolate"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type Partition struct {
	Name        string `mapstructure:"name"`
	Type        int    `mapstructure:"type"`
	Size        string `mapstructure:"size"`
	StartSector int    `mapstructure:"start_sector"`
	Filesystem  string `mapstructure:"filesystem"`
	Mountpoint  string `mapstructure:"mountpoint"`
}

type ChrootMount struct {
	MountType       string `mapstructure:"mount_type"`
	SourcePath      string `mapstructure:"source_path"`
	DestinationPath string `mapstructure:"destination_path"`
}

type ImageConfig struct {
	// where image is going to be saved
	ImagePath string `mapstructure:"image_path" required:"true"`
	// An URL to a checksum file containing a checksum for the ISO file. At
	// least one of `file+checksum` and `file+checksum_url` must be defined.
	// `file+checksum_url` will be ignored if `file+checksum` is non empty.
	ImageSize         string        `mapstructure:"image_size"`
	ImageSizeBytes    uint64        `mapstructure:"image_size_bytes"`
	ImagePartitions   []Partition   `mapstructure:"image_partitions"`
	ImageChrootMounts []ChrootMount `mapstructure:"image_chroot_mounts"`
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
