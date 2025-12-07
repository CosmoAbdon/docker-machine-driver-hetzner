package config

import "fmt"

const labelPrefix = "docker-machine-driver-hetzner/"

func LabelName(name string) string {
	return fmt.Sprintf("%s%s", labelPrefix, name)
}
