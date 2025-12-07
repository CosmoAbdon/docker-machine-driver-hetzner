package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/CosmoAbdon/docker-machine-driver-hetzner/internal/driver"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	var (
		versionFlag = flag.Bool("version", false, "Print version information")
		vFlag       = flag.Bool("v", false, "Print version (short)")
	)
	flag.Parse()

	if *versionFlag || *vFlag {
		printVersion()
		os.Exit(0)
	}

	plugin.RegisterDriver(driver.NewDriver(version))
}

func printVersion() {
	fmt.Printf("docker-machine-driver-hetzner\n")
	fmt.Printf("  Version:    %s\n", version)
	fmt.Printf("  Git commit: %s\n", gitCommit)
	fmt.Printf("  Built:      %s\n", buildDate)
}
