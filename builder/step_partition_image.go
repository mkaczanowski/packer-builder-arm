package builder

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func partitionGPT(ui packer.Ui, config *Config) multistep.StepAction {
	ui.Message(
		fmt.Sprintf("creating %d GPT partitions on %s",
			len(config.ImageConfig.ImagePartitions),
			config.ImageConfig.ImagePath),
	)
	for _, partition := range config.ImageConfig.ImagePartitions {
		cmd := []string{
			"sgdisk",
			"-n",
			fmt.Sprintf("0:%d:%s", partition.StartSector, partition.Size),
			"-t",
			fmt.Sprintf("0:%s", partition.Type),
			"-c",
			fmt.Sprintf("0:%s", partition.Name),
			config.ImageConfig.ImagePath,
		}

		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error sgdisk %v: %s", err, string(out)))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func partitionDOS(ui packer.Ui, config *Config) multistep.StepAction {
	lines := []string{
		"label: dos",
		fmt.Sprintf("device: %s", config.ImageConfig.ImagePath),
		"unit: sectors",
	}

	ui.Message(
		fmt.Sprintf("creating %d dos partitions on %s",
			len(config.ImageConfig.ImagePartitions),
			config.ImageConfig.ImagePath),
	)
	for i, partition := range config.ImageConfig.ImagePartitions {
		line := fmt.Sprintf(
			"%s%d: type=%s",
			config.ImageConfig.ImagePath,
			i+1,
			partition.Type,
		)

		if partition.StartSector != 0 {
			line = line + fmt.Sprintf(", start=%d", partition.StartSector)
		}

		if partition.Size != "" && partition.Size != "0" {
			line = line + fmt.Sprintf(", size=%s", partition.Size)
		}

		lines = append(lines, line)
	}

	cmd := exec.Command("sfdisk", config.ImageConfig.ImagePath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ui.Error(fmt.Sprintf("error while getting stdout pipe %v", err))
		return multistep.ActionHalt
	}
	defer stdout.Close()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		ui.Error(fmt.Sprintf("error while getting stdin pipe %v", err))
		return multistep.ActionHalt
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()

		ui.Error(fmt.Sprintf("error while spawning up command %v", err))
		return multistep.ActionHalt
	}

	io.WriteString(stdin, strings.Join(lines, "\n"))
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		out, _ := ioutil.ReadAll(stdout)
		ui.Error(fmt.Sprintf("error sfdisk %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// StepPartitionImage creates partitions on raw image
type StepPartitionImage struct{}

// Run the step
func (s *StepPartitionImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	zeroCmd := []string{"sgdisk", "-Z", config.ImageConfig.ImagePath}

	out, err := exec.Command(zeroCmd[0], zeroCmd[1:]...).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error sgdisk -Z %v: %s", err, string(out)))
		return multistep.ActionHalt
	}

	if config.ImageConfig.ImageType == "dos" {
		return partitionDOS(ui, config)
	}

	if config.ImageConfig.ImageType == "gpt" {
		return partitionGPT(ui, config)
	}

	return multistep.ActionHalt
}

// Cleanup after step execution
func (s *StepPartitionImage) Cleanup(state multistep.StateBag) {}
