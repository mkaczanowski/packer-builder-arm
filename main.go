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
	server.RegisterBuilder(builder.NewBuilder())
	server.Serve()
}
