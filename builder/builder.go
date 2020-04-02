//go:generate mapstructure-to-hcl2 -type Config
package builder

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// Config top-level holder for more specific configurations used
// while building packer-builder-arm
type Config struct {
	cfg.RemoteFileConfig `mapstructure:",squash"`
	cfg.ImageConfig      `mapstructure:",squash"`
	cfg.QemuConfig       `mapstructure:",squash"`

	ctx interpolate.Context
}

// Builder builds (or modifies) arm system images
type Builder struct {
	config  Config
	context context.Context
	cancel  context.CancelFunc

	runner *multistep.BasicRunner
}

// NewBuilder default Builder constructor
func NewBuilder() *Builder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Builder{
		context: ctx,
		cancel:  cancel,
	}
}

// ConfigSpec returns the config spec
func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

// InitConfig prepares relevant config structures
func (b *Builder) InitConfig(ctx *interpolate.Context) (warnings []string, errors []error) {
	var (
		warns []string
		errs  []error
	)

	warns, errs = b.config.RemoteFileConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	warns, errs = b.config.ImageConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	warns, errs = b.config.QemuConfig.Prepare(ctx)
	warnings = append(warnings, warns...)
	errors = append(errors, errs...)

	return warnings, errors
}

// Prepare setup configuration (ex. ImageConfig)
func (b *Builder) Prepare(args ...interface{}) ([]string, []string, error) {
	var (
		errs     *packer.MultiError
		warnings []string
	)

	if err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, args...); err != nil {
		return nil, nil, err
	}

	fileWarns, fileErrs := b.InitConfig(&b.config.ctx)
	warnings = append(fileWarns, fileWarns...)
	errs = packer.MultiErrorAppend(errs, fileErrs...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

// Run executes steps in order to produce the system image
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&common.StepDownload{
			Checksum:     b.config.FileChecksum,
			ChecksumType: b.config.FileChecksumType,
			Description:  "rootfs_archive",
			ResultKey:    "rootfs_archive_path",
			Url:          b.config.FileUrls,
			Extension:    b.config.TargetExtension,
			TargetPath:   b.config.TargetPath,
		},
	}

	if b.config.ImageConfig.ImageBuildMethod == "new" {
		steps = append(
			steps,
			&StepCreateBaseImage{},
			&StepPartitionImage{},
			&StepMapImage{ResultKey: "image_loop_device"},
			&StepMkfsImage{FromKey: "image_loop_device"},
			&StepMountImage{FromKey: "image_loop_device", ResultKey: "image_mountpoint", MouthPath: b.config.ImageMountPath},
			&StepPopulateFilesystem{RootfsArchiveKey: "rootfs_archive_path", ImageMountPointKey: "image_mountpoint"},
		)
	} else if b.config.ImageConfig.ImageBuildMethod == "reuse" {
		steps = append(
			steps,
			&StepExtractAndCopyImage{FromKey: "rootfs_archive_path"},
			&StepMapImage{ResultKey: "image_loop_device"},
			&StepMountImage{FromKey: "image_loop_device", ResultKey: "image_mountpoint", MouthPath: b.config.ImageMountPath},
		)
	} else {
		return nil, errors.New("invalid build method")
	}

	steps = append(
		steps,
		&StepSetupExtra{FromKey: "image_mountpoint"},
		&StepSetupChroot{ImageMountPointKey: "image_mountpoint"},
		&StepSetupQemu{ImageMountPointKey: "image_mountpoint"},
		&StepChrootProvision{ImageMountPointKey: "image_mountpoint", Hook: hook},
		&StepCompressArtifact{ImageMountPointKey: "image_mountpoint"},
	)

	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("build was halted")
	}

	return &Artifact{b.config.ImageConfig.ImagePath}, nil
}
