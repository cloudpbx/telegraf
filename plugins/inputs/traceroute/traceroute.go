package traceroute

import (
	//"fmt"
	"os/exec"
	"regexp"
	"strconv"
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

func hostTracerouter(timeout float64, args ...string) (string, error) {
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
	ColumnNum int // 0-based index of the column number (usually [0:2])
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

// processTracerouteHeaderLine parses the top line of traceroute output
// and outputs target fqdn & ip
func processTracerouteHeaderLine(line string) (string, string) {
	fqdn_re := regexp.MustCompile("[\\w-]+(\\.[\\w]+)+")
	fqdn := fqdn_re.FindString(line)

	ipv4_brackets_re := regexp.MustCompile("\\(\\d+(\\.\\d+){3}\\)")
	ip_brackets := ipv4_brackets_re.FindString(line)
	ipv4_re := regexp.MustCompile("\\d+(\\.\\d+){3}")
	ip := ipv4_re.FindString(ip_brackets)

	return fqdn, ip
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
	fqdn_re := regexp.MustCompile("[\\w-]+(\\.[\\w]+)+")
	fqdn := fqdn_re.FindString(entry)

	ipv4_brackets_re := regexp.MustCompile("\\(\\d+(\\.\\d+){3}\\)")
	ip_brackets := ipv4_brackets_re.FindString(entry)
	ipv4_re := regexp.MustCompile("\\d+(\\.\\d+){3}")
	ip := ipv4_re.FindString(ip_brackets)

	rtt_whole_re := regexp.MustCompile("\\d+\\.\\d+\\sms")
	rtt_phrase := rtt_whole_re.FindString(entry)
	rtt_re := regexp.MustCompile("\\d+\\.\\d+")
	rtt_string := rtt_re.FindString(rtt_phrase)
	rtt64, err := strconv.ParseFloat(rtt_string, 32)
	rtt := float32(rtt64)
	return fqdn, ip, rtt, err
}

func init() {
	inputs.Add("traceroute", func() telegraf.Input {
		return &Traceroute{
			tracerouteMethod: hostTracerouter,
		}
	})
}
