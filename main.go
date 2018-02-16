package main

import (
	"github.com/d2si-oss/terraform-provider-filesystem/filesystem"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: filesystem.Provider})
}
