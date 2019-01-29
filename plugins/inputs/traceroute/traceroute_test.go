package traceroute

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TRLineTest is a struct for testing traceroute line output.
type TRLineTest struct {
	Line         string
	Entries      []string
	NumberOfHops int
	HopInfo      []HopInfo
}

// TRColumnTest is a struct for testing traceroute column output.
type TRColumnTest struct {
	Text      string
	CarryOver [3]string
	FQDN      string
	IP        string
	ASN       string
	Rtt       float64
}

// MockHostTracerouter is a mock HostTracerouter that returns mock output
func MockHostTracerouter(timeout float64, args ...string) (string, error) {
	return LinuxTracerouteOutput, nil
}

func TestFindNumberOfHops(t *testing.T) {
	numHops := findNumberOfHops(LinuxTracerouteOutput)
	assert.Equal(t, 6, numHops, "6 hops made by packet")
}

var SampleTracerouteLine = `12  54.239.110.174 (54.239.110.174)  22.609 ms 54.239.110.130 (54.239.110.130)  26.629 ms 54.239.110.183 (54.239.110.183)  34.258 ms`

func TestGetHopNumber(t *testing.T) {
	hopNum, err := findHopNumber(SampleTracerouteLine)
	assert.NoError(t, err)
	assert.Equal(t, 12, hopNum, "Traceroute line is the 12th hop")
}

var TRLineTestSuite = []TRLineTest{
	NormalTRLineTest,
	SomeVoidTRLineTest,
	NoHostTRLineTest,
	AllVoidTRLineTest,
	NoHostASN1TRLineTest,
	LowerCaseASTRLineTest,
}

func TestFindColumnEntries(t *testing.T) {
	var entries []string
	for _, trLineTest := range TRLineTestSuite {
		entries = findColumnEntries(trLineTest.Line)
		assert.Equal(t, len(trLineTest.Entries), len(entries), "# entries")
		assert.True(t, reflect.DeepEqual(trLineTest.Entries, entries), "Expected: %s, Actual: %s", trLineTest.Entries, entries)
	}

	entries = findColumnEntries(NormalTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(NormalTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, NormalTracerouteEntries)

	entries = findColumnEntries(SomeVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(SomeVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, SomeVoidTracerouteEntries)

	entries = findColumnEntries(NoHostTracerouteLine)
	assert.Equal(t, 1, len(entries), "1 entry")
	assert.True(t, reflect.DeepEqual(NoHostTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, NoHostTracerouteEntries)

	entries = findColumnEntries(AllVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(AllVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, AllVoidTracerouteEntries)

}

var TRColumnTestSuite = []TRColumnTest{
	IPFQDNTRColumnTest,
	HTTPFQDNTRColumnTest,
	CarryOverTRColumnTest,
	NoHostTRColumnTest,
	ASN1TRColumnTest,
}

func TestProcessTracerouteColumnEntry(t *testing.T) {
	var fqdn, ip, asn string
	var rtt float32
	var err error
	acceptableDelta := 0.0005

	for _, trColumnTest := range TRColumnTestSuite {
		fqdn, ip, asn, rtt, err = processTracerouteColumnEntry(trColumnTest.Text, 1, trColumnTest.CarryOver[0], trColumnTest.CarryOver[1], trColumnTest.CarryOver[2])
		assert.NoError(t, err)
		assert.Equal(t, trColumnTest.FQDN, fqdn, "fqdn")
		assert.Equal(t, trColumnTest.IP, ip, "ip")
		assert.Equal(t, trColumnTest.ASN, asn, "asn")
		assert.InDelta(t, trColumnTest.Rtt, rtt, acceptableDelta, "rtt")
	}
}

func TestProcessTracerouteHopLine(t *testing.T) {
	var (
		hopInfo []HopInfo
		err     error
	)
	for _, trLineTest := range TRLineTestSuite {
		hopInfo, err = processTracerouteHopLine(trLineTest.Line)
		assert.NoError(t, err)
		expectedHopInfo := trLineTest.HopInfo
		assert.True(t, reflect.DeepEqual(expectedHopInfo, hopInfo), "Expected: %s Actual: %s", expectedHopInfo, hopInfo)
	}

}

func TestProcessTracerouteHeaderLine(t *testing.T) {
	fqdn, ip := processTracerouteHeaderLine(LinuxTracerouteHeader)
	assert.Equal(t, LinuxTracerouteFQDN, fqdn, "fqdn")
	assert.Equal(t, LinuxTracerouteIP, ip, "ip")
}

var (
	// LinuxTracerouteOutput is sample output from linux traceroute.
	LinuxTracerouteOutput = `
traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets
 1  165.227.32.254 (165.227.32.254)  1.206 ms 165.227.32.253 (165.227.32.253)  1.188 ms 165.227.32.254 (165.227.32.254)  1.143 msg
 2  138.197.249.78 (138.197.249.78)  0.985 ms 138.197.249.86 (138.197.249.86)  0.939 ms 138.197.249.90 (138.197.249.90)  1.181 ms
 3  72.14.219.10 (72.14.219.10)  0.818 ms 162.243.190.33 (162.243.190.33)  0.952 ms  0.941 ms
 4  108.170.250.225 (108.170.250.225)  0.825 ms  0.970 ms 108.170.250.241 (108.170.250.241)  1.007 ms
 5  108.170.226.217 (108.170.226.217)  0.995 ms 108.170.226.219 (108.170.226.219)  1.033 ms 108.170.226.217 (108.170.226.217)  1.003 ms
 6  dfw06s38-in-f14.1e100.net (172.217.0.238)  1.187 ms  0.722 ms  0.545 ms
`
	// LinuxTracerouteHeader is the header of LinuxTracerouteOutput.
	LinuxTracerouteHeader = `traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets`
	// LinuxTracerouteFQDN is the FQDN of LinuxTracerouteOutput.
	LinuxTracerouteFQDN = `google.com`
	// LinuxTracerouteIP is the IP of of LinuxTracerouteOutput.
	LinuxTracerouteIP = `172.217.0.238`
)

var (
	NormalTracerouteLine      = ` 6  yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms  1.244 ms  0.417 ms`
	NormalTracerouteEntries   = []string{"yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms", "1.244 ms", "0.417 ms"}
	NormalTracerouteHopNumber = 6
	NormalHopInfo             = []HopInfo{
		HopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 0,
			FQDN:      "yyz10s03-in-f3.1e100.net",
			IP:        "172.217.0.227",
			ASN:       "",
			RTT:       1.480,
		},
		HopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 1,
			FQDN:      "yyz10s03-in-f3.1e100.net",
			IP:        "172.217.0.227",
			ASN:       "",
			RTT:       1.244,
		},
		HopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 2,
			FQDN:      "yyz10s03-in-f3.1e100.net",
			IP:        "172.217.0.227",
			ASN:       "",
			RTT:       0.417,
		},
	}
	NormalTRLineTest = TRLineTest{
		Line:         NormalTracerouteLine,
		Entries:      NormalTracerouteEntries,
		NumberOfHops: NormalTracerouteHopNumber,
		HopInfo:      NormalHopInfo,
	}
)
var (
	SomeVoidTracerouteLine      = `14  54.239.110.152 (54.239.110.152)  27.198 ms * 54.239.110.247 (54.239.110.247)  37.625 ms`
	SomeVoidTracerouteEntries   = []string{"54.239.110.152 (54.239.110.152)  27.198 ms", "*", "54.239.110.247 (54.239.110.247)  37.625 ms"}
	SomeVoidTracerouteHopNumber = 14
	SomeVoidHopInfo             = []HopInfo{
		HopInfo{
			HopNumber: SomeVoidTracerouteHopNumber,
			ColumnNum: 0,
			FQDN:      "54.239.110.152",
			IP:        "54.239.110.152",
			ASN:       "",
			RTT:       27.198,
		},
		HopInfo{
			HopNumber: SomeVoidTracerouteHopNumber,
			ColumnNum: 2,
			FQDN:      "54.239.110.247",
			IP:        "54.239.110.247",
			ASN:       "",
			RTT:       37.625,
		},
	}
	SomeVoidTRLineTest = TRLineTest{
		Line:         SomeVoidTracerouteLine,
		Entries:      SomeVoidTracerouteEntries,
		NumberOfHops: SomeVoidTracerouteHopNumber,
		HopInfo:      SomeVoidHopInfo,
	}
)

var (
	NoHostTracerouteLine      = `10  129.250.2.81  186.767 ms`
	NoHostTracerouteEntries   = []string{"129.250.2.81  186.767 ms"}
	NoHostTracerouteHopNumber = 10
	NoHostHopInfo             = []HopInfo{
		HopInfo{
			HopNumber: NoHostTracerouteHopNumber,
			ColumnNum: 0,
			FQDN:      "129.250.2.81",
			IP:        "129.250.2.81",
			ASN:       "",
			RTT:       186.767,
		},
	}
	NoHostTRLineTest = TRLineTest{
		Line:         NoHostTracerouteLine,
		Entries:      NoHostTracerouteEntries,
		NumberOfHops: NoHostTracerouteHopNumber,
		HopInfo:      NoHostHopInfo,
	}
)

var (
	NoHostASN1TRLine      = `15  77.238.190.3 [AS34010]  155.664 ms 77.238.190.2 [AS34010]  155.539 ms 77.238.190.5 [AS34010]  157.304 ms`
	NoHostASN1TREntries   = []string{"77.238.190.3 [AS34010]  155.664 ms", "77.238.190.2 [AS34010]  155.539 ms", "77.238.190.5 [AS34010]  157.304 ms"}
	NoHostASN1TRHopNumber = 15
	NoHostASN1TRHopInfo   = []HopInfo{
		HopInfo{
			HopNumber: NoHostASN1TRHopNumber,
			ColumnNum: 0,
			FQDN:      "77.238.190.3",
			IP:        "77.238.190.3",
			ASN:       "AS34010",
			RTT:       155.664,
		},
		HopInfo{
			HopNumber: NoHostASN1TRHopNumber,
			ColumnNum: 1,
			FQDN:      "77.238.190.2",
			IP:        "77.238.190.2",
			ASN:       "AS34010",
			RTT:       155.539,
		},
		HopInfo{
			HopNumber: NoHostASN1TRHopNumber,
			ColumnNum: 2,
			FQDN:      "77.238.190.5",
			IP:        "77.238.190.5",
			ASN:       "AS34010",
			RTT:       157.304,
		},
	}
	NoHostASN1TRLineTest = TRLineTest{
		Line:         NoHostASN1TRLine,
		Entries:      NoHostASN1TREntries,
		NumberOfHops: NoHostASN1TRHopNumber,
		HopInfo:      NoHostASN1TRHopInfo,
	}
	NoHostASN2NumberOfHops = 14
	NoHostASN2TRLineTest   = TRLineTest{
		Line: "14  49.255.198.125 [*]  188.903 ms 101.0.127.233 [AS38880/AS38220/AS55803]  187.293 ms  182.836 ms",
		Entries: []string{
			"49.255.198.125 [*]  188.903 ms",
			"101.0.127.233 [AS38880/AS38220/AS55803]  187.293 m",
			"182.836 ms",
		},
		NumberOfHops: 14,
		HopInfo: []HopInfo{
			HopInfo{
				HopNumber: NoHostASN2NumberOfHops,
				ColumnNum: 0,
				FQDN:      "49.255.198.125",
				IP:        "49.255.198.125",
				ASN:       "",
				RTT:       188.903,
			},
			HopInfo{
				HopNumber: NoHostASN2NumberOfHops,
				ColumnNum: 1,
				FQDN:      "101.0.127.233",
				IP:        "101.0.127.233",
				ASN:       "AS38880/AS38220/AS55803",
				RTT:       187.293,
			},
			HopInfo{
				HopNumber: NoHostASN2NumberOfHops,
				ColumnNum: 2,
				FQDN:      "101.0.127.233",
				IP:        "101.0.127.233",
				ASN:       "AS38880/AS38220/AS55803",
				RTT:       187.293,
			},
		},
	}
)

var (
	NoHostASNTR2Line      = `17  101.0.127.49 [AS38880/AS38220/AS55803]  183.849 ms 101.0.126.74 [AS55803/AS38880/AS38220]  184.038 ms  180.053 ms`
	NoHostASNTR2Entries   = []string{"101.0.127.49 [AS38880/AS38220/AS55803]  183.849 ms", "101.0.126.74 [AS55803/AS38880/AS38220]  184.038 ms", "180.053 ms"}
	NoHostASNTR2HopNumber = 17
	NoHostASNTR2HopInfo   = []HopInfo{
		HopInfo{
			HopNumber: NoHostASNTR2HopNumber,
			ColumnNum: 0,
			FQDN:      "101.0.127.49",
			IP:        "101.0.127.49",
			ASN:       "AS38880/AS38220/AS55803",
			RTT:       183.849,
		},
		HopInfo{
			HopNumber: NoHostASNTR2HopNumber,
			ColumnNum: 1,
			FQDN:      "101.0.126.74",
			IP:        "101.0.126.74",
			ASN:       "AS55803/AS38880/AS38220",
			RTT:       184.038,
		},
		HopInfo{
			HopNumber: NoHostASNTR2HopNumber,
			ColumnNum: 2,
			FQDN:      "101.0.126.74",
			IP:        "101.0.126.74",
			ASN:       "AS55803/AS38880/AS38220",
			RTT:       180.053,
		},
	}
)

var (
	AllVoidTracerouteLine      = `5  * * *`
	AllVoidTracerouteEntries   = []string{"*", "*", "*"}
	AllVoidTracerouteHopNumber = 5
	AllVoidTRLineTest          = TRLineTest{
		Line:         AllVoidTracerouteLine,
		Entries:      AllVoidTracerouteEntries,
		NumberOfHops: AllVoidTracerouteHopNumber,
		HopInfo:      []HopInfo{},
	}
)

var (
	LowerCaseAsTracerouteLine      = `6  206.248.155.168 [as13768]  86.202 ms  68.356 ms  68.281 ms`
	LowerCaseAsTracerouteEntries   = []string{"206.248.155.168 [as13768]  86.202 ms", "68.356 ms", "68.281 ms"}
	LowerCaseAsTracerouteHopNumber = 6
	LowerCaseAsHopInfo             = []HopInfo{
		HopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 0,
			FQDN:      "206.248.155.168",
			IP:        "206.248.155.168",
			ASN:       "as13768",
			RTT:       86.202,
		},
		HopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 1,
			FQDN:      "206.248.155.168",
			IP:        "206.248.155.168",
			ASN:       "as13768",
			RTT:       68.356,
		},
		HopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 2,
			FQDN:      "206.248.155.168",
			IP:        "206.248.155.168",
			ASN:       "as13768",
			RTT:       68.281,
		},
	}
	LowerCaseASTRLineTest = TRLineTest{
		Line:         LowerCaseAsTracerouteLine,
		Entries:      LowerCaseAsTracerouteEntries,
		NumberOfHops: LowerCaseAsTracerouteHopNumber,
		HopInfo:      LowerCaseAsHopInfo,
	}
)
var (
	IPFQDNColumnEntry  = `12  54.239.110.174 (54.239.110.174)  22.609 ms`
	IPFQDNTRColumnTest = TRColumnTest{
		Text:      IPFQDNColumnEntry,
		CarryOver: [3]string{"", "", ""},
		FQDN:      "54.239.110.174",
		IP:        "54.239.110.174",
		ASN:       "",
		Rtt:       22.609,
	}
	HTTPFQDNColumnEntry  = `yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms`
	HTTPFQDNTRColumnTest = TRColumnTest{
		Text:      HTTPFQDNColumnEntry,
		CarryOver: [3]string{"some", "thing", "inconsequential"},
		FQDN:      "yyz10s03-in-f3.1e100.net",
		IP:        "172.217.0.227",
		ASN:       "",
		Rtt:       1.480,
	}
	CarryOverColumnEntry  = `0.417 ms`
	CarryOverParams       = [3]string{"wildmadagascar.org", "75.101.140.9", "AS16509"}
	CarryOverTRColumnTest = TRColumnTest{
		Text:      CarryOverColumnEntry,
		CarryOver: CarryOverParams,
		FQDN:      CarryOverParams[0],
		IP:        CarryOverParams[1],
		ASN:       CarryOverParams[2],
		Rtt:       0.417,
	}
	NoHostColumnEntry  = `3  192.168.1.1  2.854 ms`
	NoHostTRColumnTest = TRColumnTest{
		Text:      NoHostColumnEntry,
		CarryOver: [3]string{"Ja", "Pa", "Dog"},
		FQDN:      "192.168.1.1",
		IP:        "",
		ASN:       "",
		Rtt:       2.854,
	}
	ASN1TRColumnTest = TRColumnTest{
		Text:      "66.163.66.70 (66.163.66.70) [AS6327]  65.254 ms",
		CarryOver: [3]string{"", "", ""},
		FQDN:      "66.163.66.70",
		IP:        "66.163.66.70",
		ASN:       "AS6327",
		Rtt:       65.254,
	}
)
