package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepChrootProvision provisions the instance within a chroot
type StepChrootProvision struct {
	ImageMountPointKey string
	Hook               packer.Hook
}

// Run the step
func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	comm := &Communicator{
		Chroot: imageMountpoint,
	}

	ui.Message("running the provision hook")
	if err := s.Hook.Run(ctx, packer.HookProvision, ui, comm, nil); err != nil {
		ui.Error(fmt.Sprintf("error while running provision hook: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
