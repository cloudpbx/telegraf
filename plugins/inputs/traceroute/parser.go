package traceroute

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const header_line_len = 1

type MalformedHopLineError struct {
	line     string
	errorMsg string
}

func (e *MalformedHopLineError) Error() string {
	return fmt.Sprintf(`Hop line "%s" is malformed: %s`, e.line, e.errorMsg)
}

func parseTracerouteResults(output string) (string, int, []TracerouteHopInfo, error) {
	var target_ip string
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	number_of_hops := len(outputLines) - header_line_len
	hop_info := []TracerouteHopInfo{}
	for i, line := range outputLines {
		if i == 0 {
			_, target_ip = processTracerouteHeaderLine(line)
		} else {
			lineHopInfo, err := processTracerouteHopLine(line)
			if err != nil {
				return target_ip, number_of_hops, hop_info, err
			}
			hop_info = append(hop_info, lineHopInfo...)
		}
	}

	return target_ip, number_of_hops, hop_info, nil
}

type TracerouteHopInfo struct {
	HopNumber int // nth hop from root (ex. 1st hop = 1)
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
func processTracerouteHopLine(line string) ([]TracerouteHopInfo, error) {
	var err error
	hopInfo := []TracerouteHopInfo{}
	hopNumber, err := findHopNumber(line)
	if err != nil {
		return hopInfo, err
	}

	colEntries := findColumnEntries(line)

	var fqdn, ip string
	var rtt float32
	for i, entry := range colEntries {
		if entry != "*" {
			fqdn, ip, rtt, err = processTracerouteColumnEntry(entry, i, fqdn, ip)
			if err != nil {
				return hopInfo, err
			}
			if ip == "" {
				ip = fqdn
			}
			hopInfo = append(hopInfo, TracerouteHopInfo{
				HopNumber: hopNumber,
				ColumnNum: i,
				Fqdn:      fqdn,
				Ip:        ip,
				RTT:       rtt,
			})
		}
	}

	return hopInfo, err
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
