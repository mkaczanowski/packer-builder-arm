package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// StepCreateBaseImage creates the base image (empty file of given size via dd)
type StepCreateBaseImage struct{}

// Run the step
func (s *StepCreateBaseImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("creating an empty image %s", config.ImageConfig.ImagePath))
	out, err := exec.Command(
		"dd",
		"if=/dev/zero",
		fmt.Sprintf("of=%s", config.ImageConfig.ImagePath),
		fmt.Sprintf("bs=%d", config.ImageConfig.ImageSizeBytes),
		"count=1",
	).CombinedOutput()

	ui.Say(fmt.Sprintf("dd output: %s", string(out)))
	if err != nil {
		ui.Error(fmt.Sprintf("error dd %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepCreateBaseImage) Cleanup(state multistep.StateBag) {}
