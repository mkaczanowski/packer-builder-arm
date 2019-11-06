package step

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	cfg "github.com/mkaczanowski/packer-builder-arm/config"
)

func renderAndExecuteTemplate(tmplName string, tmpl *template.Template, state multistep.StateBag) (multistep.StepAction, error) {
	config := state.Get("config").(*cfg.Config)
	ui := state.Get("ui").(packer.Ui)

	// Create temporary config file
	tmpfile, err := ioutil.TempFile("", tmplName)
	if err != nil {
		return multistep.ActionHalt, fmt.Errorf("error while creating %s config file: %v", tmplName, err)
	}
	defer os.Remove(tmpfile.Name())

	// Render template
	ui.Message(fmt.Sprintf("rendering template %s", tmpfile.Name()))
	if err = tmpl.Execute(tmpfile, config.ImageConfig); err != nil {
		return multistep.ActionHalt, fmt.Errorf("error while rendering template %s: %v", tmplName, err)
	}

	// Add execution bit
	tmpfile.Close()
	os.Chmod(tmpfile.Name(), 0755)

	// Partition the image
	out, err := exec.Command(
		tmpfile.Name(),
		config.ImageConfig.ImagePath,
	).CombinedOutput()

	ui.Say(fmt.Sprintf("execution output: %s", string(out)))
	if err != nil {
		return multistep.ActionHalt, fmt.Errorf("error while executing template %s %v: %s", tmplName, err, string(out))
	}

	return multistep.ActionContinue, nil
}
