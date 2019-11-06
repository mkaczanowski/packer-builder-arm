package step

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepPartitionImage struct{}

func (s *StepPartitionImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	action, err := renderAndExecuteTemplate("sgdisk-config", diskSchemaTemplate, state)
	if err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return action
	}

	return action
}

func (s *StepPartitionImage) Cleanup(state multistep.StateBag) {}
