package traceroute

import (
	"exec"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type HostTracerouter func(timeout float64, args ...string) (string, error)

// Traceroute struct should be named the same as the Plugin
type Traceroute struct {

	// urls to traceroute
	Urls []string

	// host traceroute function
	hostTraceroute HostTracerouter
}

// Description will appear directly above the plugin definition in the config file
func (t *Traceroute) Description() string {
	return "Traceroutes given url(s) and return statistics"
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
const sampleConfig = `
## List of urls to traceroute
urls = ["www.google.com"] # required
`

func (t *Traceroute) SampleConfig() string {
	return sampleConfig
}

// Gather defines what data the plugin will gather.
func (t *Traceroute) Gather(acc telegraf.Accumulator) error {

	return nil
}

func hostTraceRouter(timeout float64, args ...string) (string, error) {
	var c *exec.Cmd // FIXME
	out, err := internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout+5))
	return string(out), err
}

func (t *Traceroute) args(url string) []string {
	args := []string{""}
	return args
}

func processTraceroute(out string) error {
	return nil
}
