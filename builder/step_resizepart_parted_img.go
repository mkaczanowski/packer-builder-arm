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

// StepResizepartPartedImage expand already partitioned image
type StepResizepartPartedImage struct {
	FromKey string
}

// Run the step
func (s *StepResizepartPartedImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	var err error
	var out []byte

	// step 1: resize partition image with parted
	out, err = exec.Command("parted", config.ImageConfig.ImagePath, "---pretend-input-tty","resizepart",fmt.Sprintf("%d",config.ImageConfig.ImagePartitionExpand), "100%").CombinedOutput()
	ui.Message(fmt.Sprintf("Set the resize %v image of partition %s to 100% of available size.", config.ImageConfig.ImagePath,fmt.Sprintf("%d",config.ImageConfig.ImagePartitionExpand)))
	if err != nil {
		ui.Error(fmt.Sprintf("error while resize the image with parted %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue	
}

// Cleanup after step execution
func (s *StepResizepartPartedImage) Cleanup(state multistep.StateBag) {}