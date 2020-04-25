package builder
//"os"
//"github.com/mholt/archiver"
import (
	"context"
	"fmt"
	"strings"
	"os/exec"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

)

// StepResizepartitionKpartxImage expand already partitioned image
type StepResizepartitionKpartxImage struct {
	FromKey string
}

// Run the step
func (s *StepResizepartitionKpartxImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	var err error
	var out []byte

	// step 1: mount image for resize
	out, err = exec.Command("kpartx","-av", config.ImageConfig.ImagePath).CombinedOutput()
	ui.Message(fmt.Sprintf("Mount the image %v for resize ", config.ImageConfig.ImagePath))
	if err != nil {
		ui.Error(fmt.Sprintf("error while mount the image with parted %v: %s", err, out))
		return multistep.ActionHalt
	}


	var res []string = strings.SplitAfter(string(out), " ")
	var virtualdevice string = ""

	for _, word := range res {
		if (strings.Contains(word, "loop") && strings.Contains(word, fmt.Sprintf("p%d",config.ImageConfig.ImagePartitionExpand))) {
			virtualdevice = word
		}
		
	}

	// step 2: resize partition image with parted
	out, err = exec.Command("resize2fs", "-f", fmt.Sprintf("/dev/mapper/%v",strings.TrimSpace(virtualdevice))).CombinedOutput()
	ui.Message(fmt.Sprintf("Resize the image mounted in %v ", fmt.Sprintf("/dev/mapper/%v",virtualdevice)))
	if err != nil {
		ui.Error(fmt.Sprintf("error while resize the image with parted %v: %s", err, out))
		return multistep.ActionHalt
	}

	// step 3: dismount image for resize
	out, err = exec.Command("kpartx","-d", config.ImageConfig.ImagePath).CombinedOutput()
	ui.Message(fmt.Sprintf("Dismount the image %v ", config.ImageConfig.ImagePath))
	if err != nil {
		ui.Error(fmt.Sprintf("error while dismount the image with parted %v: %s", err, out))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue	
}

// Cleanup after step execution
func (s *StepResizepartitionKpartxImage) Cleanup(state multistep.StateBag) {}