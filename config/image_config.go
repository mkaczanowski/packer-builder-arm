package config

import (
	"errors"
	"os"

	"github.com/dustin/go-humanize"
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
	Index       int    `mapstructure:"int"`
	Name        string `mapstructure:"name"`
	Type        string `mapstructure:"type"`
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
	ImageType         string        `mapstructure:"image_type"`
	ImageMountPath    string        `mapstructure:"image_mount_path"`
	ImageBuildMethod  string        `mapstructure:"image_build_method"`
	ImageSizeBytes    uint64        `mapstructure:"image_size_bytes"`
	ImagePartitions   []Partition   `mapstructure:"image_partitions"`
	ImageChrootMounts []ChrootMount `mapstructure:"image_chroot_mounts"`
	AdditionalChrootMounts []ChrootMount `mapstructure:"additional_chroot_mounts"`
	ImageSetupExtra   [][]string    `mapstructure:"image_setup_extra"`
	ImageChrootEnv    []string      `mapstructure:"image_chroot_env"`
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

	if c.ImageType == "" {
		c.ImageType = "dos"
	}

	if !(c.ImageType == "dos" || c.ImageType == "gpt") {
		errs = append(errs, errors.New("supported image types are: gpt, dos"))
	}

	if c.ImageBuildMethod == "" {
		errs = append(errs, errors.New("image build method must be specified"))
	}

	if !(c.ImageBuildMethod == "new" || c.ImageBuildMethod == "reuse") {
		errs = append(errs, errors.New("invalid image build method specified (valid options: new, reuse)"))
	}

	if len(c.ImagePartitions) == 0 {
		errs = append(errs, errors.New("you need to specify at least one partition"))
	}

	if len(c.ImageChrootMounts) == 0 {
		c.ImageChrootMounts = []ChrootMount{
			{MountType: "proc", SourcePath: "proc", DestinationPath: "/proc"},
			{MountType: "sysfs", SourcePath: "sysfs", DestinationPath: "/sys"},
			{MountType: "bind", SourcePath: "/dev", DestinationPath: "/dev"},
			{MountType: "devpts", SourcePath: "/devpts", DestinationPath: "/dev/pts"},
			{MountType: "binfmt_misc", SourcePath: "binfmt_misc", DestinationPath: "/proc/sys/fs/binfmt_misc"},
		}
	}

	if len(c.AdditionalChrootMounts) > 0 {
		c.ImageChrootMounts = append(c.ImageChrootMounts, c.AdditionalChrootMounts...)
	}

	return warnings, errs
}
