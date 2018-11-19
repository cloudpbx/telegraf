package traceroute

import (
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/cloudpbx/ive-measurement/iplookup"
	tr "github.com/cloudpbx/ive-measurement/msm/traceroute"
	"github.com/cloudpbx/ive-measurement/msm/traceroute/metric"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	tr_measurement  = "traceroute"
	hop_measurement = "traceroute_hop_data"
	version         = "2.000"
)

// Description will appear directly above the plugin definition in the config file
func (t *Traceroute) Description() string {
	return "Traceroutes given url(s) and return statistics"
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
const sampleConfig = `
# NOTE: this plugin forks the traceroute command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/traceroute
  #
  ## List of urls to traceroute
  urls = ["www.google.com"] # required
  ## per-traceroute timeout, in s. 0 == no timeout
  ## it is highly recommended to set this value to match the telegraf interval
  # response_timeout = 0.0
  ## wait time per probe in seconds (traceroute -w <WAITTIME>)
  # waittime = 5.0
  ## starting TTL of packet (traceroute -f <FIRST_TTL>)
  # first_ttl = 1
  ## maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
  # max_ttl = 30
  ## number of probe packets sent per hop (traceroute -q <NQUERIES>)
  # nqueries = 3
  ## do not try to map IP addresses to host names (traceroute -n)
  # no_host_name = false
  ## use ICMP packets (traceroute -I)
  # icmp = false
  ## Lookup AS path in routes (traceroute -A)
  # as_path_lookups = false
  ## source interface/address to traceroute from (traceroute -i <INTERFACE/SRC_ADDR>)
  # interface = ""
`

func (t *Traceroute) SampleConfig() string {
	return sampleConfig
}

// Gather defines what data the plugin will gather.

func (t *Traceroute) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	netID := iplookup.FindNetID()
	pacc := &pluginAccumulator{acc: acc}
	dummyHost := "will be deleted"
	dummyTime := time.Now()
	for i, host_url := range t.Urls {
		wg.Add(1)
		go func(target_fqdn string) {
			defer wg.Done()
			tags := map[string]string{"target_fqdn": target_fqdn}
			fields := make(map[string]interface{})

			_, err := net.LookupHost(target_fqdn)
			if err != nil {
				pacc.AddError(err)
				fields["result_code"] = tr.HostNotFound
				pacc.Add(tr_measurement, tags, fields, dummyTime)
				return
			}

			tr_args := t.args(target_fqdn)
			rawOutput, err := t.tracerouteMethod(t.ResponseTimeout, tr_args...)
			if err != nil {
				pacc.AddError(err)
				return
			}

			output, err := tr.ParseTracerouteResults(rawOutput)
			if err != nil {
				pacc.AddError(err)
				return
			}
			metric.ParseTROutput(pacc, output, netID, dummyHost, dummyTime)

		}(host_url)
		if (i%50) == 0 && (i > 0) {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return nil
}

type pluginAccumulator struct {
	sync.Mutex
	acc telegraf.Accumulator
}

func (pacc *pluginAccumulator) Add(name string,
	tags map[string]string,
	fields map[string]interface{},
	_ time.Time,
) error {
	pacc.Lock()
	defer pacc.Unlock()
	delete(tags, "host")
	pacc.acc.AddFields(name, fields, tags)
	return nil
}

func (pacc *pluginAccumulator) AddError(err error) {
	pacc.Lock()
	defer pacc.Unlock()
	pacc.acc.AddError(err)
	return
}

func hostTracerouter(timeout float64, args ...string) (string, error) {
	var out []byte
	bin, err := exec.LookPath("traceroute")
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	if timeout == float64(0) {
		out, err = executeWithoutTimeout(c)
	} else {
		out, err = internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout+5))
	}
	return string(out), err
}

func init() {
	inputs.Add("traceroute", func() telegraf.Input {
		return &Traceroute{
			ResponseTimeout:  0,
			WaitTime:         5.0,
			FirstTTL:         1,
			MaxTTL:           30,
			Nqueries:         3,
			NoHostname:       false,
			UseICMP:          false,
			tracerouteMethod: hostTracerouter,
		}
	})
}
