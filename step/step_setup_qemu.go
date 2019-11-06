package step

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

type StepSetupQemu struct {
	ImageMountPointKey string
}

func (s *StepSetupQemu) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	srcPath := config.QemuConfig.QemuBinarySourcePath
	dstPath := filepath.Join(imageMountpoint, config.QemuConfig.QemuBinaryDestinationPath)

	ui.Message(fmt.Sprintf("copying qemu binary from %s to: %s", srcPath, dstPath))
	out, err := exec.Command("cp", srcPath, dstPath).CombinedOutput()
	if err != nil {
		ui.Error(fmt.Sprintf("error while copying %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepSetupQemu) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*cfg.Config)
	imageMountpoint := state.Get(s.ImageMountPointKey).(string)
	dstPath := filepath.Join(imageMountpoint, config.QemuConfig.QemuBinaryDestinationPath)

	os.Remove(dstPath)
}
