package step

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateEmptyImage struct {
	FromKey string
}

func (s *stepCreateEmptyImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	//rootfsArchive := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("Creating empty image %s", config.ImageConfig.ImagePath))
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
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	//state.Put(s.ResultKey, partitions)

	return multistep.ActionContinue
}

func (s *stepCreateEmptyImage) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)

	_, err := os.Stat(config.ImageConfig.ImagePath)
	if !os.IsNotExist(err) {
		//os.Remove(config.ImageConfig.ImagePath)
	}
}
