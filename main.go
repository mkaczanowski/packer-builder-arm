package main

import (
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/mkaczanowski/packer-plugin-arm/builder"
	"github.com/mkaczanowski/packer-plugin-arm/version"
	"os"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(builder.Builder))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
