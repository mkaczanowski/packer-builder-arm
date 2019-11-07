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

// Partition describes a single partition that is going to be
// created on raw image
type Partition struct {
	Name        string `mapstructure:"name"`
	Type        int    `mapstructure:"type"`
	Size        string `mapstructure:"size"`
	StartSector int    `mapstructure:"start_sector"`
	Filesystem  string `mapstructure:"filesystem"`
	Mountpoint  string `mapstructure:"mountpoint"`
}

// ChrootMount describes a mountpoint that is being setup
// as part of the chroot environment
type ChrootMount struct {
	MountType       string `mapstructure:"mount_type"`
	SourcePath      string `mapstructure:"source_path"`
	DestinationPath string `mapstructure:"destination_path"`
}

// ImageConfig describes the raw image properties
type ImageConfig struct {
	ImagePath         string        `mapstructure:"image_path" required:"true"`
	ImageSize         string        `mapstructure:"image_size"`
	ImageSizeBytes    uint64        `mapstructure:"image_size_bytes"`
	ImagePartitions   []Partition   `mapstructure:"image_partitions"`
	ImageChrootMounts []ChrootMount `mapstructure:"image_chroot_mounts"`
}

// Prepare image configuration
func (c *ImageConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	if c.ImageSize != "" && c.ImageSizeBytes != 0 {
		errs = append(errs, errors.New("only one of image_size or image_size_bytes can be specified"))
	}

	if c.ImageSize != "" {
		if got, err := humanize.ParseBytes(c.ImageSize); err != nil {
			errs = append(errs, err)
		} else {
			c.ImageSizeBytes = got
		}
	}

	if c.ImageSizeBytes == 0 {
		errs = append(errs, errors.New("one of image_size_bytes or image_size must be set"))
	}

	if len(c.ImagePartitions) == 0 {
		errs = append(errs, errors.New("you need to specify at least one partition"))
	}

	if len(c.ImageChrootMounts) == 0 {
		c.ImageChrootMounts = []ChrootMount{
			ChrootMount{MountType: "proc", SourcePath: "proc", DestinationPath: "/proc"},
			ChrootMount{MountType: "sysfs", SourcePath: "sysfs", DestinationPath: "/sys"},
			ChrootMount{MountType: "bind", SourcePath: "/dev", DestinationPath: "/dev"},
			ChrootMount{MountType: "devpts", SourcePath: "/devpts", DestinationPath: "/dev/pts"},
			ChrootMount{MountType: "binfmt_misc", SourcePath: "binfmt_misc", DestinationPath: "/proc/sys/fs/binfmt_misc"},
		}
	}

	return warnings, errs
}
