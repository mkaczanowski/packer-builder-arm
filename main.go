package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/mkaczanowski/packer-builder-arm/builder"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	if err := server.RegisterBuilder(builder.NewBuilder()); err != nil {
		panic(err)
	}
	server.Serve()
}
