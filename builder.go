package main

import (
	"context"
	"errors"

	. "config"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "builder-arm"

type Builder struct {
	config  configs.Config
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

	fileWarns, fileErrs := b.config.Prepare(&b.config.ctx)
	warnings = append(fileWarns, fileWarns...)
	errs = packer.MultiErrorAppend(errs, fileErrs...)

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	state := new(multistep.BasicStateBag)
	//state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("hook", hook)
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
		&stepCreateEmptyImage{FromKey: "rootfs_archive_path"},
		//&stepCopyImage{FromKey: "iso_path", ResultKey: "imagefile", ImageOpener: image.NewImageOpener(ui)},
	}

	b.runner = &multistep.BasicRunner{Steps: steps}

	// Executes the steps
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	// check if it is ok
	_, canceled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if canceled || halted {
		return nil, errors.New("step canceled or halted")
	}

	return &Artifact{"test"}, nil
}
