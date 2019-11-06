package step

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

type StepSetupChroot struct {
	ImageMountPointKey string
}

func sortMountpoints(chrootMounts []cfg.ChrootMount, reverse bool) []cfg.ChrootMount {
	mounts := make([]cfg.ChrootMount, len(chrootMounts))
	copy(mounts, chrootMounts)

	sort.Slice(mounts, func(i, j int) bool {
		if reverse {
			return mounts[i].DestinationPath > mounts[j].DestinationPath
		}
		return mounts[i].DestinationPath < mounts[j].DestinationPath
	})

	return mounts
}

func prepareCmd(chrootMount cfg.ChrootMount, mountpoint string) []string {
	cmd := []string{
		"mount",
	}

	if chrootMount.MountType == "bind" {
		cmd = append(cmd, "--bind")
	} else {
		cmd = append(cmd, "-t", chrootMount.MountType)
	}

	return append(cmd, chrootMount.SourcePath, mountpoint)
}

func getMounts() (map[string]bool, error) {
	dat, err := ioutil.ReadFile("/etc/mtab")
	if err != nil {
		return nil, err
	}

	selected := make(map[string]bool)
	all := strings.Split(string(dat), "\n")

	for _, line := range all {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		selected[string(fields[1])] = true
	}

	return selected, nil
}

func (s *StepSetupChroot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	chrootMounts := sortMountpoints(config.ImageConfig.ImageChrootMounts, false)

	for _, chrootMount := range chrootMounts {
		mountpoint := filepath.Join(imageMountpoint, chrootMount.DestinationPath)
		cmd := prepareCmd(chrootMount, mountpoint)

		if err := os.MkdirAll(mountpoint, 0755); err != nil {
			err := fmt.Errorf("Error creating mount directory: %s", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("mounting %s with: %s", chrootMount.SourcePath, cmd))
		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error while mounting %v: %s", err, out))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepSetupChroot) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	chrootMounts := sortMountpoints(config.ImageConfig.ImageChrootMounts, true)
	imageMountpoint := state.Get(s.ImageMountPointKey).(string)

	mounted, err := getMounts()
	if err != nil {
		ui.Error(fmt.Sprintf("unable to read mtab: %v", err))
	}

	for _, chrootMount := range chrootMounts {
		mountpoint := filepath.Join(imageMountpoint, chrootMount.DestinationPath)
		if _, ok := mounted[mountpoint]; !ok {
			continue
		}

		out, err := exec.Command("umount", mountpoint).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error while unmounting %v: %s", err, out))
		}
	}
}
