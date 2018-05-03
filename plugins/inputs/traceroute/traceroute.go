package traceroute

import (
	//"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type HostTracerouter func(timeout float64, args ...string) (string, error)

// Traceroute struct should be named the same as the Plugin
type Traceroute struct {

	// urls to traceroute
	Urls []string

	// host traceroute function
	tracerouteMethod HostTracerouter
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
	bin, err := exec.LookPath("traceroute")
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	out, err := internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout+5))
	return string(out), err
}

func (t *Traceroute) args(url string) []string {
	args := []string{url}
	//args = append(args, url)
	return args
}

type TracerouteHopInfo struct {
	PacketNum int // 1-based index of the column number (usually [1:3])
	Fqdn      string
	Ip        string
	RTT       float32 //milliseconds
}

func processTracerouteOutput(out string) (int, error) {
	var numHops int = -1
	lines := strings.Split(strings.TrimSpace(out), "\n")
	numHops = len(lines) - 1
	return numHops, nil
}

// processTracerouteHopLine parses
func processTracerouteHopLine(line string) (int, []TracerouteHopInfo, error) {
	hopInfo := []TracerouteHopInfo{}
	return 0, hopInfo, nil
}

func init() {
	inputs.Add("traceroute", func() telegraf.Input {
		return &Traceroute{
			tracerouteMethod: hostTraceRouter,
		}
	})
}
