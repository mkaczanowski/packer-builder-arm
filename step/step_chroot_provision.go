package step

// taken from here: https://github.com/hashicorp/packer/blob/81522dced0b25084a824e79efda02483b12dc7cd/builder/amazon/chroot/step_chroot_provision.go

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/mkaczanowski/packer-builder-arm/communicator"
)

// StepChrootProvision provisions the instance within a chroot.
type StepChrootProvision struct {
	ImageMountPointKey string
	Hook               packer.Hook
}

func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	comm := &communicator.Communicator{
		Chroot: imageMountpoint,
	}

	ui.Message("running the provision hook")
	if err := s.Hook.Run(ctx, packer.HookProvision, ui, comm, nil); err != nil {
		ui.Error(fmt.Sprintf("error while running provision hook: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
