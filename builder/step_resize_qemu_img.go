package builder
//"os"
//"github.com/mholt/archiver"
import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

)

// StepResizeQemuImage expand already partitioned image
type StepResizeQemuImage struct {
	FromKey string
}

// Run the step
func (s *StepResizeQemuImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	var err error
	var out []byte

	// step 1: resize image with qemu-img (as image)
	out, err = exec.Command("qemu-img", "resize" , config.ImageConfig.ImagePath, string(config.ImageConfig.ImageSize)).CombinedOutput()
	ui.Message(fmt.Sprintf("Resize the image file %v to %s",config.ImageConfig.ImagePath, string(config.ImageConfig.ImageSize)))
	if err != nil {
		ui.Error(fmt.Sprintf("error while resize the image %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue	
}

// Cleanup after step execution
func (s *StepResizeQemuImage) Cleanup(state multistep.StateBag) {}