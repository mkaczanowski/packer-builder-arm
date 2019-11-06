package builder

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
	"github.com/mkaczanowski/packer-builder-arm/step"
)

const BuilderId = "builder-arm"

type Builder struct {
	config  cfg.Config
	context context.Context
	cancel  context.CancelFunc

	runner *multistep.BasicRunner
}

func NewBuilder() *Builder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Builder{
		context: ctx,
		cancel:  cancel,
	}
}

func (b *Builder) Prepare(args ...interface{}) ([]string, error) {
	var (
		errs     *packer.MultiError
		warnings []string
	)

	if err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, args...); err != nil {
		return nil, err
	}

	fileWarns, fileErrs := b.config.Prepare(&b.config.Ctx)
	warnings = append(fileWarns, fileWarns...)
	errs = packer.MultiErrorAppend(errs, fileErrs...)

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

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
		&step.StepCreateBaseImage{},
		&step.StepPartitionImage{},
		&step.StepMkfsImage{},
		&step.StepMapImage{ResultKey: "image_loop_device"},
		&step.StepMountImage{FromKey: "image_loop_device", ResultKey: "image_mountpoint"},
		&step.StepPopulateFilesystem{RootfsArchiveKey: "rootfs_archive_path", ImageMountPointKey: "image_mountpoint"},
		&step.StepSetupChroot{ImageMountPointKey: "image_mountpoint"},
		&step.StepSetupQemu{ImageMountPointKey: "image_mountpoint"},
		&step.StepChrootProvision{ImageMountPointKey: "image_mountpoint", Hook: hook},
	}

	b.runner = &multistep.BasicRunner{Steps: steps}

	// Executes the steps
	b.runner.Run(ctx, state)

	// check if it is ok
	_, canceled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if canceled || halted {
		return nil, errors.New("step canceled or halted")
	}

	return &Artifact{b.config.ImageConfig.ImagePath}, nil
}
