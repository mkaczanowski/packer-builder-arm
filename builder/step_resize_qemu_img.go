package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepResizeQemuImage expand already partitioned image
type StepResizeQemuImage struct {
}

// Run the step
func (s *StepResizeQemuImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	// resize base image (.img) with qemu tooling
	out, err := exec.Command("qemu-img", "resize", config.ImageConfig.ImagePath, string(config.ImageConfig.ImageSize)).CombinedOutput()
	ui.Message(fmt.Sprintf("resizing the image file %v to %s", config.ImageConfig.ImagePath, string(config.ImageConfig.ImageSize)))
	if err != nil {
		ui.Error(fmt.Sprintf("error while resizing the image %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepResizeQemuImage) Cleanup(state multistep.StateBag) {}
