package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func checkBinfmtMisc(srcPath string) (string, error) {
	files, err := os.ReadDir("/proc/sys/fs/binfmt_misc")
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/sys/fs/binfmt_misc directory: %v", err)
	}

	srcPathStat, _ := os.Stat(srcPath)
	for _, file := range files {
		if file.Name() == "register" || file.Name() == "status" {
			continue
		}

		pth := filepath.Join("/proc/sys/fs/binfmt_misc", file.Name())
		dat, err := os.ReadFile(pth)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %s, err: %v", file.Name(), err)
		}

		for _, line := range strings.Split(string(dat), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			if fields[0] == "interpreter" {
				fieldStat, _ := os.Stat(fields[1])
				// os.SameFile allows also comparing of sym- and relative symlinks.
				if os.SameFile(fieldStat, srcPathStat) {
					return pth, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Failed to find binfmt_misc for %s under /proc/sys/fs/binfmt_misc", srcPath)
}

// StepSetupQemu configures chroot environment to run binaries via qemu
type StepSetupQemu struct {
	ImageMountPointKey string
}

// Run the step
func (s *StepSetupQemu) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	srcPath := config.QemuConfig.QemuBinarySourcePath
	dstPath := filepath.Join(imageMountpoint, config.QemuConfig.QemuBinaryDestinationPath)

	// check if binfmt_misc is present
	binfmt, err := checkBinfmtMisc(srcPath)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("binfmt setup found at: %s", binfmt))

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		ui.Error(fmt.Sprintf("error while creating path: %v", err))
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("copying qemu binary from %s to: %s", srcPath, dstPath))
	out, err := exec.Command("cp", srcPath, dstPath).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error while copying %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepSetupQemu) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	dstPath := filepath.Join(imageMountpoint, config.QemuConfig.QemuBinaryDestinationPath)

	if err := os.Remove(dstPath); err != nil {
		ui.Error(err.Error())
	}
}
