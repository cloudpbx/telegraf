package traceroute

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// headerLineLen is the length of traceroute header.
const headerLineLen = 1

// VoidASNValue is the symbol for a void ASN character.
const VoidASNValue = "*"

// MalformedHopLineError occurs when a traceroute hop line is malformed.
type MalformedHopLineError struct {
	line     string
	errorMsg string
}

func (e *MalformedHopLineError) Error() string {
	return fmt.Sprintf(`Hop line "%s" is malformed: %s`, e.line, e.errorMsg)
}

// parseTracerouteResults parses string traceroute output into OutputData.
func parseTracerouteResults(output string) (OutputData, error) {
	var targetFQDN, targetIP string
	response := OutputData{}
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	numberOfHops := len(outputLines) - headerLineLen
	hopInfo := []HopInfo{}
	for i, line := range outputLines {
		if i == 0 {
			targetFQDN, targetIP = processTracerouteHeaderLine(line)
			response.TargetFQDN = targetFQDN
			response.TargetIP = targetIP
		} else {
			lineHopInfo, err := processTracerouteHopLine(line)
			if err != nil {
				return response, err
			}
			hopInfo = append(hopInfo, lineHopInfo...)
		}
	}
	response.NumberOfHops = numberOfHops
	response.HopInfo = hopInfo
	return response, nil
}

// OutputData is the structure of a processed traceroute.
type OutputData struct {
	TargetFQDN   string
	TargetIP     string
	NumberOfHops int
	HopInfo      []HopInfo
}

// HopInfo is the structure of a hop in a traceroute.
type HopInfo struct {
	HopNumber int // nth hop from root (ex. 1st hop = 1)
	ColumnNum int // 0-based index of the column number (usually [0:2])
	FQDN      string
	IP        string
	ASN       string
	RTT       float32 //milliseconds
}

var fqdnRe = regexp.MustCompile("([\\w-]+(\\.[\\w-]+)*(\\.[a-z]{2,63}))|(\\d+(\\.\\d+){3})")
var ipv4WithBracketsRe = regexp.MustCompile("\\(\\d+(\\.\\d+){3}\\)")
var ipv4Re = regexp.MustCompile("\\d+(\\.\\d+){3}")
var rttWithMSRe = regexp.MustCompile("\\d+\\.\\d+\\sms")
var rttRe = regexp.MustCompile("\\d+\\.\\d+")
var asnWithBracketsRe = regexp.MustCompile("\\[(\\*|((AS|as)[\\d]+)(\\/((AS|as)[\\d]+))?)\\]")
var asnRe = regexp.MustCompile("(\\*|((AS|as)[\\d]+)(\\/((AS|as)[\\d]+))?)")

// processTracerouteHeaderLine parses the top line of traceroute output
// and outputs target fqdn & ip.
func processTracerouteHeaderLine(line string) (string, string) {
	fqdn := fqdnRe.FindString(line)

	ipBrackets := ipv4WithBracketsRe.FindString(line)
	ip := ipv4Re.FindString(ipBrackets)

	return fqdn, ip
}

// findNumberOfHops returns the number of hops to a traceroute.
// In the case of any empty traceroute, 0 is returned.
func findNumberOfHops(out string) int {
	numHops := -1
	lines := strings.Split(strings.TrimSpace(out), "\n")
	numHops = len(lines) - 1
	return numHops
}

// processTracerouteHopLine parses hop information
// present after the header line outputted by traceroute.
func processTracerouteHopLine(line string) ([]HopInfo, error) {
	var err error
	hopInfo := []HopInfo{}
	hopNumber, err := findHopNumber(line)
	if err != nil {
		return hopInfo, err
	}

	colEntries := findColumnEntries(line)

	var fqdn, ip, asn string
	var rtt float32
	for i, entry := range colEntries {
		if entry != "*" {
			fqdn, ip, asn, rtt, err = processTracerouteColumnEntry(entry, i, fqdn, ip, asn)
			if err != nil {
				return hopInfo, err
			}
			if ip == "" {
				ip = fqdn
			}

			hopInfo = append(hopInfo, HopInfo{
				HopNumber: hopNumber,
				ColumnNum: i,
				FQDN:      fqdn,
				IP:        ip,
				ASN:       asn,
				RTT:       rtt,
			})
		}
	}

	return hopInfo, err
}

// findHopNumber parses the hop number
func findHopNumber(rawline string) (int, error) {
	line := strings.TrimSpace(rawline)
	re := regexp.MustCompile("^[\\d]+")
	hopNumString := re.FindString(line)
	return strconv.Atoi(hopNumString)
}

var columnEntryRe = regexp.MustCompile("\\*|(([\\w-]+(\\.[\\w-]+)+)\\s(\\(\\d+(\\.\\d+){0,3}\\))?\\s*(\\[(\\*|(((AS|as)[\\d]+)(\\/(AS|as)[\\d]+)*))\\])?\\s*)?(\\d+\\.\\d+\\sms)")

// findColumnEntries parses a line of traceroute output
// and finds column entries signified by "*", or "[fqdn]? ([ip])? ms".
func findColumnEntries(line string) []string {
	return columnEntryRe.FindAllString(line, -1)
}

// processTracerouteColumnEntry parses column entry
// and extracts fqdn, ip, rtt if available
// in the case where the fqdn & ip are "carried over", the inputted fqdn, ip are used.
func processTracerouteColumnEntry(entry string, columnNum int, lastFQDN, lastIP string, lastASN string) (string, string, string, float32, error) {
	fqdn, ip, asn, rtt, err := processTracerouteColumnEntryHelper(entry)
	if (fqdn == "" && ip == "") && columnNum > 0 {
		fqdn = lastFQDN
		ip = lastIP
		asn = lastASN
	}
	return fqdn, ip, asn, rtt, err
}

// processTracerouteColumnEntryHelper is a helper function for parsing a traceroute column
func processTracerouteColumnEntryHelper(entry string) (string, string, string, float32, error) {
	fqdn := fqdnRe.FindString(entry)

	ipBrackets := ipv4WithBracketsRe.FindString(entry)
	ip := ipv4Re.FindString(ipBrackets)

	asnBrackets := asnWithBracketsRe.FindString(entry)
	asn := asnRe.FindString(asnBrackets)
	if asn == VoidASNValue {
		asn = ""
	}

	rttPhrase := rttWithMSRe.FindString(entry)
	rttString := rttRe.FindString(rttPhrase)
	rtt64, err := strconv.ParseFloat(rttString, 32)
	rtt := float32(rtt64)
	return fqdn, ip, asn, rtt, err
}
