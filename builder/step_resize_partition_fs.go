package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mkaczanowski/packer-builder-arm/config"
)

// StepResizePartitionFs expand already partitioned image
type StepResizePartitionFs struct {
	FromKey              string
	SelectedPartitionKey string
}

// Find partitions marked with ResizeFs that we explicitly want to resize
func findPartitionsToResize(imagePartitions []config.Partition) []int {
	var selectedPartitions []int

	for i, partition := range imagePartitions {
		if partition.ResizeFs {
			selectedPartitions = append(selectedPartitions, i+1)
		}
	}

	return selectedPartitions
}

// Run the step
func (s *StepResizePartitionFs) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui         = state.Get("ui").(packer.Ui)
		loopDevice = state.Get(s.FromKey).(string)
	)

	var selectedPartitions []int

	// If we're running in resize mode, we'll have a single selected partition
	// to resize from StepExpandPartition, and we can be sure we don't need to
	// expand any other partitions. If we're in repartition mode, we manually
	// choose which partitions to resize: only the user knows which ones
	// have been expanded.
	if s.SelectedPartitionKey != "" {
		selectedPartitions = append(selectedPartitions, state.Get(s.SelectedPartitionKey).(int))
	} else {
		config := state.Get("config").(*Config)
		selectedPartitions = findPartitionsToResize(config.ImagePartitions)
	}

	for _, partition := range selectedPartitions {
		device := fmt.Sprintf("%sp%d", loopDevice, partition)
		out, err := exec.Command("resize2fs", "-f", device).CombinedOutput()
		ui.Message(fmt.Sprintf("running resize2fs on %s ", device))
		if err != nil {
			ui.Error(fmt.Sprintf("error while resizing partition %v: %s", err, out))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepResizePartitionFs) Cleanup(_ multistep.StateBag) {}
