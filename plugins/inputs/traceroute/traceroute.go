package traceroute

import (
	//"fmt"
	"bytes"
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

type HostTracerouter func(timeout float64, args ...string) (string, error)

// Traceroute struct should be named the same as the Plugin
type Traceroute struct {

	// URLs to traceroute
	Urls []string

	// Total timeout duration each traceroute call, in seconds. 0 means no timeout
	// Default: 0
	ResponseTimeout float64 `toml:"response_timeout"`

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
urls = ["www.google.com","0.0.0.0"] # required
`

func (t *Traceroute) SampleConfig() string {
	return sampleConfig
}

// Gather defines what data the plugin will gather.
func (t *Traceroute) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, host_url := range t.Urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			tags := map[string]string{"target_fqdn": url}
			fields := make(map[string]interface{})

			_, err := net.LookupHost(url)
			if err != nil {
				acc.AddError(err)
				fields["result_code"] = 1
				acc.AddFields(tr_measurement, fields, tags)
				return
			}

			tr_args := t.args(url)
			output, err := t.tracerouteMethod(t.ResponseTimeout, tr_args...)
			outputLines := strings.Split(strings.TrimSpace(output), "\n")

			var target_ip string
			for i, line := range outputLines {
				if i == 0 {
					_, target_ip = processTracerouteHeaderLine(line)
					tags["target_ip"] = target_ip
					fields["number_of_hops"] = len(outputLines) - header_line_len
				} else {

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

func findHopNumber(line string) (int, error) {
	re := regexp.MustCompile("^[\\d]+")
	hopNumString := re.FindString(line)
	return strconv.Atoi(hopNumString)
}

// findColumnEntries parses a line of traceroute output
// and finds column entries signified by "*", or "[fqdn]? ([ip])? ms"
func findColumnEntries(line string) []string {
	re := regexp.MustCompile("\\*|(([\\w-]+(\\.[\\w]+)+)\\s\\(\\d+(\\.\\d+){3}\\)\\s*)?(\\d+\\.\\d+\\sms)")
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
			tracerouteMethod: hostTracerouter,
		}
	})
}
