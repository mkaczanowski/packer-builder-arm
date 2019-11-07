package builder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// StepPartitionImage creates partitions on raw image
type StepPartitionImage struct{}

// Run the step
func (s *StepPartitionImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*cfg.Config)
	zeroCmd := []string{"sgdisk", "-Z", config.ImageConfig.ImagePath}

	out, err := exec.Command(zeroCmd[0], zeroCmd[1:]...).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error sgdisk -Z %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	for _, partition := range config.ImageConfig.ImagePartitions {
		cmd := []string{
			"sgdisk",
			"-n",
			fmt.Sprintf("0:%d:%s", partition.StartSector, partition.Size),
			"-t",
			fmt.Sprintf("0:%d", partition.Type),
			"-c",
			fmt.Sprintf("0:%s", partition.Name),
			config.ImageConfig.ImagePath,
		}

		out, err = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error sgdisk %v: %s", err, string(out)))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepPartitionImage) Cleanup(state multistep.StateBag) {}
