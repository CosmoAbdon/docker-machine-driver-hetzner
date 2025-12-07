//go:build instrumented

package driver

import (
	"encoding/json"
	"os"
	"runtime/debug"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/docker/machine/libmachine/log"
)

const runningInstrumented = false

func instrumented[T any](input T) T {
	j, err := json.Marshal(input)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	log.Debugf("%v\n%v\n", string(debug.Stack()), string(j))
	return input
}

type debugLogWriter struct {
}

func (x debugLogWriter) Write(data []byte) (int, error) {
	log.Debug(string(data))
	return len(data), nil
}

func (d *Driver) getClientInstrumentationOpts() []hcloud.ClientOption {
	if os.Getenv("HETZNER_DRIVER_HTTP_DEBUG") == "42" {
		return []hcloud.ClientOption{hcloud.WithDebugWriter(debugLogWriter{})}
	}
	return nil
}
