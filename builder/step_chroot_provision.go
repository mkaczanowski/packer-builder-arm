package builder

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/chroot"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepChrootProvision provisions the instance within a chroot
type StepChrootProvision struct {
	ImageMountPointKey string
	Hook               packer.Hook
	SetupQemu          bool
}

// Run the step
func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	comm := &chroot.Communicator{
		Chroot: imageMountpoint,
		CmdWrapper: func(cmd string) (string, error) {
			if s.SetupQemu {
				return fmt.Sprintf(
					"%s %s",
					strings.Join(config.ImageConfig.ImageChrootEnv, " "),
					cmd,
				), nil
			} else {
				return fmt.Sprintf(
					"%s",
					cmd,
				), nil
			}
		},
	}
	hookData := commonsteps.PopulateProvisionHookData(state)

	ui.Message("running the provision hook")
	if err := s.Hook.Run(ctx, packer.HookProvision, ui, comm, hookData); err != nil {
		ui.Error(fmt.Sprintf("error while running provision hook: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
