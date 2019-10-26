package main

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer/packer"

	"github.com/hashicorp/packer/helper/multistep"
)

func ShellCommand(command string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", command)
}

func run(state multistep.StateBag, cmds string) error {
	ui := state.Get("ui").(packer.Ui)
	stderr := new(bytes.Buffer)
	cmd := ShellCommand(cmds)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error executing command '%s': %s\nStderr: %s", cmds, err, stderr.String())
		ui.Error(err.Error())
		return err
	}
	return nil
}
