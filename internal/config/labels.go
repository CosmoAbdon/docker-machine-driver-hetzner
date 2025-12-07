package config

import "fmt"

const labelPrefix = "docker-machine-driver-hetzner/"

// LabelName returns a fully qualified label name with the driver prefix.
func LabelName(name string) string {
	return fmt.Sprintf("%s%s", labelPrefix, name)
}
