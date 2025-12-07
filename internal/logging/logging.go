package logging

import (
	"fmt"

	"github.com/docker/machine/libmachine/log"
)

const (
	PrefixStep    = "  -> "  
	PrefixSubstep = "     " 
)

func Step(format string, args ...any) {
	log.Infof(PrefixStep+format, args...)
}

func Substep(format string, args ...any) {
	log.Infof(PrefixSubstep+format, args...)
}

func DebugStep(format string, args ...any) {
	log.Debugf(PrefixStep+format, args...)
}

func WarnStep(format string, args ...any) {
	log.Warnf(PrefixStep+format, args...)
}

func Server(name string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", name, id)
}

func Action(command string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", command, id)
}

func Key(name string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", name, id)
}
