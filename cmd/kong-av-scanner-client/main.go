package main

import (
	"github.com/IntegrationWorks/kong-plugin-virus-scan/internal/avclient"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

func main() {
	server.StartServer(New, Version, Priority)
}

// Version is the version of the plugin.
var Version = "0.2"

// Priority is the priority of the plugin.
var Priority = 1

func New() interface{} {
	return &Config{}
}

type Config struct {
	// ScannerURL references the URL required to hit the ICAP scanning server / AV scanner. E.g. icap://myscannerhost:1344/avscan
	ScannerURL string
}

/* Access method */
func (conf Config) Access(kong *pdk.PDK) {
	avclient.DoAccess(kong, conf.ScannerURL)
}
