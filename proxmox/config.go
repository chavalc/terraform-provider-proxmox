package proxmox

import (
	"log"

	"github.com/chavalc/goproxmox"
)

type Config struct {
	Host     string
	Username string
	Password string
}

// Client() returns a new client for accessing Arubacloud.
func (c *Config) Client() (*goproxmox.Client, error) {
	client := goproxmox.NewClient(c.Host, c.Username, c.Password)

	log.Printf("[INFO] Proxmox client configured for URL: %s", client.BaseURL.String())

	return client, nil
}
