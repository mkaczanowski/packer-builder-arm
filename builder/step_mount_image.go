package builder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

func sortMountablePartitions(partitions []cfg.Partition, reverse bool) []cfg.Partition {
	mountable := []cfg.Partition{}

	for i, partition := range partitions {
		partition.Index = i + 1
		if partition.Mountpoint != "" {
			mountable = append(mountable, partition)
		}
	}

	sort.Slice(mountable, func(i, j int) bool {
		if reverse {
			return mountable[i].Mountpoint > mountable[j].Mountpoint
		}
		return mountable[i].Mountpoint < mountable[j].Mountpoint
	})

	return mountable
}

// StepMountImage mounts partition to selected mountpoints
type StepMountImage struct {
	FromKey     string
	ResultKey   string
	tempdir     string
	mountpoints []string
}

// Run the step
func (s *StepMountImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	loopDevice := state.Get(s.FromKey).(string)

	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.tempdir = tempdir

	partitions := sortMountablePartitions(config.ImageConfig.ImagePartitions, false)
	for _, partition := range partitions {
		mountpoint := filepath.Join(s.tempdir, partition.Mountpoint)
		device := fmt.Sprintf("%sp%d", loopDevice, partition.Index)

		if err := os.MkdirAll(mountpoint, 0755); err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("mounting %s to %s", device, mountpoint))
		_, err := exec.Command("mount", device, mountpoint).CombinedOutput()
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.mountpoints = append(s.mountpoints, mountpoint)
	}

	state.Put(s.ResultKey, s.tempdir)

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepMountImage) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	if s.tempdir != "" {
		partitions := sortMountablePartitions(config.ImageConfig.ImagePartitions, true)
		for _, partition := range partitions {
			mountpoint := filepath.Join(s.tempdir, partition.Mountpoint)

			_, err := exec.Command("umount", mountpoint).CombinedOutput()
			if err != nil {
				ui.Error(err.Error())
			}
		}
		s.mountpoints = nil

		if err := os.Remove(s.tempdir); err != nil {
			ui.Error(err.Error())
		}

		s.tempdir = ""
	}
}
