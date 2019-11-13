package builder

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

// StepSetupExtra creates filesystem on already partitioned image
type StepSetupExtra struct {
	FromKey string
}

func replaceVars(l []string, config *cfg.Config, imageMountpoint string) []string {
	newList := make([]string, len(l))
	defined := map[string]string{
		"$MOUNTPOINT": imageMountpoint,
		"$IMAGE_PATH": config.ImageConfig.ImagePath,
	}

	for i, v := range l {
		for key, _ := range defined {
			v = strings.ReplaceAll(v, key, defined[key])
		}
		newList[i] = v
	}

	return newList
}

// Run the step
func (s *StepSetupExtra) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*cfg.Config)
	imageMountpoint := state.Get(s.FromKey).(string)

	for _, cmd := range config.ImageConfig.ImageSetupExtra {
		cmd = replaceVars(cmd, config, imageMountpoint)

		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			ui.Error(fmt.Sprintf("error while executing cmd: %v: %v: %s", cmd, err, string(out)))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup after step execution
func (s *StepSetupExtra) Cleanup(state multistep.StateBag) {}
