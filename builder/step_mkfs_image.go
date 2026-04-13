package builder

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepMkfsImage creates filesystem on already partitioned image
type StepMkfsImage struct {
	FromKey string
}

// Run the step
func (s *StepMkfsImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	loopDevice := state.Get(s.FromKey).(string)

	for i, partition := range config.ImageConfig.ImagePartitions {
		if partition.SkipMkfs {
			ui.Message(fmt.Sprintf("skipping mkfs for partition #%d", i+1))
			continue
		}

		cmd := fmt.Sprintf("mkfs.%s", partition.Filesystem)
		args := append(partition.FilesystemMakeOptions, fmt.Sprintf("%sp%d", loopDevice, i+1))

		ui.Message(fmt.Sprintf("creating partition #%d: mkfs.%s %s", i+1, cmd, strings.Join(args, " ")))
		out, err := exec.Command(cmd, args...).CombinedOutput()

		if err != nil {
			ui.Error(fmt.Sprintf("error mkfs %v: %s", err, string(out)))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepMkfsImage) Cleanup(_ multistep.StateBag) {}
