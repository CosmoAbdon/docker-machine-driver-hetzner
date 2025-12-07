package driver

import (
	"fmt"

	"github.com/docker/machine/libmachine/log"
)

// Log level prefixes for consistent formatting
const (
	logPrefixStep    = "  -> "  // Primary action step
	logPrefixSubstep = "     " // Sub-action or detail (continuation)
)

// logStep logs an info-level message for a primary action step.
// Use this for main operations like "Creating server...", "Destroying key..."
func logStep(format string, args ...any) {
	log.Infof(logPrefixStep+format, args...)
}

// logSubstep logs an info-level message for a sub-action or detail.
// Use this for follow-up details under a main step.
func logSubstep(format string, args ...any) {
	log.Infof(logPrefixSubstep+format, args...)
}

// logDebugStep logs a debug-level message for a primary action step.
func logDebugStep(format string, args ...any) {
	log.Debugf(logPrefixStep+format, args...)
}

// logWarnStep logs a warning-level message for a primary action step.
func logWarnStep(format string, args ...any) {
	log.Warnf(logPrefixStep+format, args...)
}

// logServer formats server information consistently: "ServerName [ID: 123]"
func logServer(name string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", name, id)
}

// logAction formats action information consistently: "command [ID: 123]"
func logAction(command string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", command, id)
}

// logKey formats SSH key information consistently: "KeyName [ID: 123]"
func logKey(name string, id int64) string {
	return fmt.Sprintf("%s [ID: %d]", name, id)
}
