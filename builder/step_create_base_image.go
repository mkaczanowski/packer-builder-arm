package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCreateBaseImage creates the base image (empty file of given size via dd)
type StepCreateBaseImage struct{}

// Run the step
func (s *StepCreateBaseImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("creating an empty image %s", config.ImageConfig.ImagePath))
	out, err := exec.Command(
		"dd",
		"if=/dev/zero",
		fmt.Sprintf("of=%s", config.ImageConfig.ImagePath),
		"bs=1",
		"count=0",
		fmt.Sprintf("seek=%d", config.ImageConfig.ImageSizeBytes),
	).CombinedOutput()

	ui.Say(fmt.Sprintf("dd output: %s", string(out)))
	if err != nil {
		ui.Error(fmt.Sprintf("error dd %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepCreateBaseImage) Cleanup(_ multistep.StateBag) {}
