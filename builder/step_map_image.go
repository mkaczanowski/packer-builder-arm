package builder

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepMapImage maps system image to /dev/loopX
type StepMapImage struct {
	ResultKey  string
	loopDevice string
}

// Run the step
func (s *StepMapImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	image := config.ImageConfig.ImagePath

	// find empty loop device
	ui.Message(fmt.Sprintf("searching for empty loop device (to map %s)", image))
	out, err := exec.Command("losetup", "-f").Output()

	if err != nil {
		ui.Error(fmt.Sprintf("error losetup -f %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	s.loopDevice = strings.Trim(string(out), "\n")

	// map image with losetup
	ui.Message(fmt.Sprintf("mapping image %s to %s", image, s.loopDevice))
	out, err = exec.Command("losetup", "-P", s.loopDevice, image).CombinedOutput()

	if err != nil {
		ui.Error(fmt.Sprintf("error losetup -P %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, s.loopDevice)

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepMapImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	out, err := exec.Command("losetup", "-d", s.loopDevice).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error while unmounting loop device %v: %s", err, string(out)))
	}
}
