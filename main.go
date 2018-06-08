package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/chavalc/terraform-provider-proxmox/proxmox"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: proxmox.Provider,
	})
}
