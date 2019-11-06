package step

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepMkfsImage struct{}

func (s *StepMkfsImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	action, err := renderAndExecuteTemplate("mkfs-config", filesystemSchemaTemplate, state)
	if err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return action
	}

	return action
}

func (s *StepMkfsImage) Cleanup(state multistep.StateBag) {}
