package step

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

type StepMapImage struct {
	ResultKey  string
	loopDevice string
}

func (s *StepMapImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)
	image := config.ImageConfig.ImagePath

	// Find empty loop device
	ui.Message(fmt.Sprintf("searching for empty loop device (to map %s)", image))
	out, err := exec.Command("losetup", "-f").Output()

	if err != nil {
		ui.Error(fmt.Sprintf("error losetup -f %v: %s", err, string(out)))
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	s.loopDevice = strings.Trim(string(out), "\n")

	// Map image
	ui.Message(fmt.Sprintf("mapping image %s to %s", image, s.loopDevice))
	out, err = exec.Command("losetup", "-P", s.loopDevice, image).CombinedOutput()

	if err != nil {
		ui.Error(fmt.Sprintf("error losetup -P %v: %s", err, string(out)))
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, s.loopDevice)

	return multistep.ActionContinue
}

func (s *StepMapImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	out, err := exec.Command("losetup", "-d", s.loopDevice).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error while unmounting loop device %v: %s", err, string(out)))
	}
}
