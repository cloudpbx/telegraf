package traceroute

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	tr_measurement  = "traceroute"
	hop_measurement = "traceroute_hop_data"
	header_line_len = 1
)

type MalformedHopLineError struct {
	line     string
	errorMsg string
}

func (e *MalformedHopLineError) Error() string {
	return fmt.Sprintf(`Hop line "%s" is malformed: %s`, e.line, e.errorMsg)
}

type HostTracerouter func(timeout float64, args ...string) (string, error)

// Traceroute struct should be named the same as the Plugin
type Traceroute struct {

	// URLs to traceroute
	Urls []string

	// Total timeout duration each traceroute call, in seconds. 0 means no timeout
	// Type: float
	// Default: 0.0
	ResponseTimeout float64 `toml:"response_timeout"`

	// Wait time per probe in seconds (traceroute -w <WAITTIME>)
	// Type: float
	// Default: 5.0 sec
	WaitTime float64 `toml:"waittime"`

	// Starting TTL of packet (traceroute -f <FIRST_TTL>)
	// Type: int
	// Default: 1
	FirstTTL int `toml:"first_ttl"`

	// Maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
	// Type: int
	// Default: 30
	MaxTTL int `toml:"max_ttl"`

	// Number of probe packets sent per hop (traceroute -q <NQUERIES>)
	// Type: int
	// Default: 3
	Nqueries int `toml:"nqueries"`

	// Do not try to map IP addresses to host names (traceroute -n)
	// Default: false
	NoHostname bool `toml:"no_host_name"`

	// Use ICMP packets (traceroute -I)
	// Default: false
	UseICMP bool `toml:"icmp"`

	// Source interface/address (traceroute -i <INTERFACE/SRC_ADDR>)
	// Type: string
	Interface string `toml:"interface"`

	// host traceroute function
	tracerouteMethod HostTracerouter
}

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

	for _, host_url := range t.Urls {
		wg.Add(1)
		go func(target_fqdn string) {
			defer wg.Done()
			tags := map[string]string{"target_fqdn": target_fqdn}
			fields := make(map[string]interface{})

			_, err := net.LookupHost(target_fqdn)
			if err != nil {
				acc.AddError(err)
				fields["result_code"] = 1
				acc.AddFields(tr_measurement, fields, tags)
				return
			}

			tr_args := t.args(target_fqdn)
			output, err := t.tracerouteMethod(t.ResponseTimeout, tr_args...)
			outputLines := strings.Split(strings.TrimSpace(output), "\n")

			var target_ip string
			for i, line := range outputLines {
				if i == 0 {
					_, target_ip = processTracerouteHeaderLine(line)
					tags["target_ip"] = target_ip
					fields["number_of_hops"] = len(outputLines) - header_line_len
				} else {
					hopNumber, hopInfo, err := processTracerouteHopLine(line)
					if err != nil {
						acc.AddError(&MalformedHopLineError{line, err.Error()})
					}
					for _, info := range hopInfo {
						hopTags := map[string]string{
							"target_fqdn":   target_fqdn,
							"target_ip":     target_ip,
							"column_number": strconv.Itoa(info.ColumnNum),
							"hop_fqdn":      info.Fqdn,
							"hop_ip":        info.Ip,
						}
						hopFields := map[string]interface{}{
							"hop_number": hopNumber,
							"hop_rtt_ms": info.RTT,
						}
						acc.AddFields(hop_measurement, hopFields, hopTags)
					}
				}
			}
			acc.AddFields(tr_measurement, fields, tags)

		}(host_url)
	}

	return nil
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

func executeWithoutTimeout(c *exec.Cmd) ([]byte, error) {
	var b bytes.Buffer
	c.Stderr = &b
	out, err := c.Output()
	if err != nil {
		out = b.Bytes()
	}
	return out, err
}

func (t *Traceroute) args(url string) []string {
	args := []string{url}
	//args = append(args, url)
	if t.WaitTime > 0.0 {
		args = append(args, "-w", strconv.FormatFloat(t.WaitTime, 'f', -1, 64))
	}
	if t.FirstTTL > 0 {
		args = append(args, "-f", strconv.Itoa(t.FirstTTL))
	}
	if t.MaxTTL > 0 && t.MaxTTL >= t.FirstTTL {
		args = append(args, "-m", strconv.Itoa(t.MaxTTL))
	}
	if t.Nqueries > 0 {
		args = append(args, "-q", strconv.Itoa(t.Nqueries))
	}
	if t.NoHostname {
		args = append(args, "-n")
	}
	if t.UseICMP {
		args = append(args, "-I")
	}
	if t.Interface != "" {
		args = append(args, "-i", t.Interface)
	}
	return args
}

type TracerouteHopInfo struct {
	ColumnNum int // 0-based index of the column number (usually [0:2])
	Fqdn      string
	Ip        string
	RTT       float32 //milliseconds
}

var fqdn_re = regexp.MustCompile("[\\w-]+(\\.[\\w]+)+")
var ipv4_with_brackets_re = regexp.MustCompile("\\(\\d+(\\.\\d+){3}\\)")
var ipv4_re = regexp.MustCompile("\\d+(\\.\\d+){3}")
var rtt_with_ms_re = regexp.MustCompile("\\d+\\.\\d+\\sms")
var rtt_re = regexp.MustCompile("\\d+\\.\\d+")

// processTracerouteHeaderLine parses the top line of traceroute output
// and outputs target fqdn & ip
func processTracerouteHeaderLine(line string) (string, string) {
	fqdn := fqdn_re.FindString(line)

	ip_brackets := ipv4_with_brackets_re.FindString(line)
	ip := ipv4_re.FindString(ip_brackets)

	return fqdn, ip
}

func findNumberOfHops(out string) int {
	var numHops int = -1
	lines := strings.Split(strings.TrimSpace(out), "\n")
	numHops = len(lines) - 1
	return numHops
}

// processTracerouteHopLine parses hop information
// present after the header line outputted by traceroute
func processTracerouteHopLine(line string) (int, []TracerouteHopInfo, error) {
	var err error
	hopInfo := []TracerouteHopInfo{}
	hopNumber, err := findHopNumber(line)
	if err != nil {
		return hopNumber, hopInfo, err
	}

	colEntries := findColumnEntries(line)

	var fqdn, ip string
	var rtt float32
	for i, entry := range colEntries {
		if entry != "*" {
			fqdn, ip, rtt, err = processTracerouteColumnEntry(entry, i, fqdn, ip)
			if err != nil {
				return hopNumber, hopInfo, err
			}
			if ip == "" {
				ip = fqdn
			}
			hopInfo = append(hopInfo, TracerouteHopInfo{
				ColumnNum: i,
				Fqdn:      fqdn,
				Ip:        ip,
				RTT:       rtt,
			})
		}
	}

	return hopNumber, hopInfo, err
}

func findHopNumber(rawline string) (int, error) {
	line := strings.TrimSpace(rawline)
	re := regexp.MustCompile("^[\\d]+")
	hopNumString := re.FindString(line)
	return strconv.Atoi(hopNumString)
}

// findColumnEntries parses a line of traceroute output
// and finds column entries signified by "*", or "[fqdn]? ([ip])? ms"
func findColumnEntries(line string) []string {
	re := regexp.MustCompile("\\*|(([\\w-]+(\\.[\\w]+)+)\\s(\\(\\d+(\\.\\d+){0,3}\\))?\\s*)?(\\d+\\.\\d+\\sms)")
	return re.FindAllString(line, -1)
}

// processTracerouteColumnEntry parses column entry
// and extracts fqdn, ip, rtt if available
// in the case where the fqdn & ip are "carried over", the inputted fqdn, ip are used
func processTracerouteColumnEntry(entry string, columnNum int, last_fqdn, last_ip string) (string, string, float32, error) {
	fqdn, ip, rtt, err := processTracerouteColumnEntryHelper(entry)
	if (fqdn == "" || ip == "") && columnNum > 0 {
		fqdn = last_fqdn
		ip = last_ip
	}
	return fqdn, ip, rtt, err
}

func processTracerouteColumnEntryHelper(entry string) (string, string, float32, error) {
	fqdn := fqdn_re.FindString(entry)

	ip_brackets := ipv4_with_brackets_re.FindString(entry)
	ip := ipv4_re.FindString(ip_brackets)

	rtt_phrase := rtt_with_ms_re.FindString(entry)
	rtt_string := rtt_re.FindString(rtt_phrase)
	rtt64, err := strconv.ParseFloat(rtt_string, 32)
	rtt := float32(rtt64)
	return fqdn, ip, rtt, err
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
