package builder

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepExpandPartition expand already partitioned image
type StepExpandPartition struct {
	ResultKey string
}

func findExpandablePartition(config *Config) (int, error) {
	partitions := []int{}

	for i, partition := range config.ImageConfig.ImagePartitions {
		// resizefs works only for ext partition family
		if partition.Size == "0" && strings.HasPrefix(partition.Filesystem, "ext") {
			partitions = append(partitions, i+1)
		}
	}

	if len(partitions) > 1 {
		return 0, fmt.Errorf(
			"found %d resizable partions (%v), but we can expand only one",
			len(partitions), partitions,
		)
	}

	if len(partitions) == 0 {
		return 0, errors.New("couldn't find any partition to expand")
	}

	return partitions[0], nil
}

// Run the step
func (s *StepExpandPartition) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	partitionIndex, err := findExpandablePartition(config)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	out, err := exec.Command("parted", config.ImageConfig.ImagePath, "---pretend-input-tty", "resizepart", strconv.Itoa(partitionIndex), "100%").CombinedOutput()
	ui.Message(fmt.Sprintf("expanding partition no. %d on the image %s", partitionIndex, config.ImageConfig.ImagePath))
	if err != nil {
		ui.Error(fmt.Sprintf("error while expanding partition %v: %s", err, out))
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, partitionIndex)

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepExpandPartition) Cleanup(state multistep.StateBag) {}
