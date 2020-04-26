package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepResizePartitionFs expand already partitioned image
type StepResizePartitionFs struct {
	FromKey              string
	SelectedPartitionKey string
}

// Run the step
func (s *StepResizePartitionFs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui = state.Get("ui").(packer.Ui)

		loopDevice        = state.Get(s.FromKey).(string)
		selectedPartition = state.Get(s.SelectedPartitionKey).(int)
		device            = fmt.Sprintf("%sp%d", loopDevice, selectedPartition)
	)

	out, err := exec.Command("resize2fs", "-f", device).CombinedOutput()
	ui.Message(fmt.Sprintf("running resize2fs on %s ", device))
	if err != nil {
		ui.Error(fmt.Sprintf("error while resizing partition %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepResizePartitionFs) Cleanup(state multistep.StateBag) {}
