package main

import (
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/mkaczanowski/packer-builder-arm/builder"
	"os"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, builder.NewBuilder())
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
