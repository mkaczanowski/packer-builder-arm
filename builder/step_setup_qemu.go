package builder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

func checkBinfmtMisc(srcPath string) (string, error) {
	files, err := ioutil.ReadDir("/proc/sys/fs/binfmt_misc")
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/sys/fs/binfmt_misc directory: %v", err)
	}

	for _, file := range files {
		if file.Name() == "register" || file.Name() == "status" {
			continue
		}

		pth := filepath.Join("/proc/sys/fs/binfmt_misc", file.Name())
		dat, err := ioutil.ReadFile(pth)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %s, err: %v", file.Name(), err)
		}

		for _, line := range strings.Split(string(dat), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			if fields[0] == "interpreter" && fields[1] == srcPath {
				return pth, nil
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
	config := state.Get("config").(*cfg.Config)
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
	config := state.Get("config").(*cfg.Config)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	dstPath := filepath.Join(imageMountpoint, config.QemuConfig.QemuBinaryDestinationPath)

	if err := os.Remove(dstPath); err != nil {
		ui.Error(err.Error())
	}
}
