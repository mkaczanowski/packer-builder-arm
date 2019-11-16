package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// StepMkfsImage creates filesystem on already partitioned image
type StepMkfsImage struct {
	FromKey string
}

// Run the step
func (s *StepMkfsImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*cfg.Config)
	loopDevice := state.Get(s.FromKey).(string)

	ui.Message("running mkfs")
	for i, partition := range config.ImageConfig.ImagePartitions {
		out, err := exec.Command(
			fmt.Sprintf("mkfs.%s", partition.Filesystem),
			fmt.Sprintf("%sp%d", loopDevice, i+1),
		).CombinedOutput()

		if err != nil {
			ui.Error(fmt.Sprintf("error mkfs %v: %s", err, string(out)))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepMkfsImage) Cleanup(state multistep.StateBag) {}
