package traceroute

import (
	//"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	LinuxTracerouteOutput = `
traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets
 1  165.227.32.254 (165.227.32.254)  1.206 ms 165.227.32.253 (165.227.32.253)  1.188 ms 165.227.32.254 (165.227.32.254)  1.143 ms
 2  138.197.249.78 (138.197.249.78)  0.985 ms 138.197.249.86 (138.197.249.86)  0.939 ms 138.197.249.90 (138.197.249.90)  1.181 ms
 3  72.14.219.10 (72.14.219.10)  0.818 ms 162.243.190.33 (162.243.190.33)  0.952 ms  0.941 ms
 4  108.170.250.225 (108.170.250.225)  0.825 ms  0.970 ms 108.170.250.241 (108.170.250.241)  1.007 ms
 5  108.170.226.217 (108.170.226.217)  0.995 ms 108.170.226.219 (108.170.226.219)  1.033 ms 108.170.226.217 (108.170.226.217)  1.003 ms
 6  dfw06s38-in-f14.1e100.net (172.217.0.238)  1.187 ms  0.722 ms  0.545 ms
`
	LinuxTracerouteHeader = `traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets`
	LinuxTracerouteFqdn   = `google.com`
	LinuxTracerouteIp     = `172.217.0.238`
)

func MockHostTracerouter(timeout float64, args ...string) (string, error) {
	return LinuxTracerouteOutput, nil
}

func TestConstructorSimple(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{
		tracerouteMethod: MockHostTracerouter,
	}
	sampleConfig := tr.SampleConfig()
	//fmt.Println(sampleConfig)
	assert.NotNil(t, sampleConfig)
}

/*
func TestHostTracerouterSimple(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{}
	args := tr.args("google.com")
	_, err := hostTracerouter(0, args...)
	if err != nil {
		t.Fatal("call failed:", err)
	}
}
*/

func TestHostTracerouteBadHost(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{}
	args := tr.args("badhost")
	_, err := hostTracerouter(3, args...)
	assert.Error(t, err)
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

var (
	NormalTracerouteLine      = `6  yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms  1.244 ms  0.417 ms`
	NormalTracerouteEntries   = []string{"yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms", "1.244 ms", "0.417 ms"}
	NormalTracerouteHopNumber = 6
	NormalTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			ColumnNum: 0,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			RTT:       1.480,
		},
		TracerouteHopInfo{
			ColumnNum: 1,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			RTT:       1.244,
		},
		TracerouteHopInfo{
			ColumnNum: 2,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			RTT:       0.417,
		},
	}
)
var (
	SomeVoidTracerouteLine      = `14  54.239.110.152 (54.239.110.152)  27.198 ms * 54.239.110.247 (54.239.110.247)  37.625 ms`
	SomeVoidTracerouteEntries   = []string{"54.239.110.152 (54.239.110.152)  27.198 ms", "*", "54.239.110.247 (54.239.110.247)  37.625 ms"}
	SomeVoidTracerouteHopNumber = 14
	SomeVoidTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			ColumnNum: 0,
			Fqdn:      "54.239.110.152",
			Ip:        "54.239.110.152",
			RTT:       27.198,
		},
		TracerouteHopInfo{
			ColumnNum: 2,
			Fqdn:      "54.239.110.247",
			Ip:        "54.239.110.247",
			RTT:       37.625,
		},
	}
)
var AllVoidTracerouteLine = `5  * * *`
var AllVoidTracerouteEntries = []string{"*", "*", "*"}
var AllVoidTracerouteHopNumber = 5

func TestFindColumnEntries(t *testing.T) {
	var entries []string
	entries = findColumnEntries(NormalTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(NormalTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, NormalTracerouteEntries)

	entries = findColumnEntries(SomeVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(SomeVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, SomeVoidTracerouteEntries)

	entries = findColumnEntries(AllVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(AllVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, AllVoidTracerouteEntries)

}

var IpFqdnColumnEntry = `12  54.239.110.174 (54.239.110.174)  22.609 ms`
var HttpFqdnColumnEntry = `yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms`
var CarryOverColumnEntry = `0.417 ms`

func TestProcessTracerouteColumnEntry(t *testing.T) {
	var fqdn, ip string
	var rtt float32
	var err error
	acceptableDelta := 0.0005

	fqdn, ip, rtt, err = processTracerouteColumnEntry(IpFqdnColumnEntry, 0, "", "")
	assert.NoError(t, err)
	assert.Equal(t, "54.239.110.174", fqdn, "fqdn")
	assert.Equal(t, "54.239.110.174", ip, "ip")
	assert.InDelta(t, 22.609, rtt, acceptableDelta, "rtt")

	fqdn, ip, rtt, err = processTracerouteColumnEntry(HttpFqdnColumnEntry, 3, "something.not.useful.org", "255.255.255.255")
	assert.NoError(t, err)
	assert.Equal(t, "yyz10s03-in-f3.1e100.net", fqdn, "fqdn")
	assert.Equal(t, "172.217.0.227", ip, "ip")
	assert.InDelta(t, 1.480, rtt, acceptableDelta, "rtt")

	carryOverFqdn := "wildmadagascar.org"
	carryOverIp := "75.101.140.9"
	fqdn, ip, rtt, err = processTracerouteColumnEntry(CarryOverColumnEntry, 1, carryOverFqdn, carryOverIp)
	assert.NoError(t, err)
	assert.Equal(t, carryOverFqdn, fqdn, "fqdn")
	assert.Equal(t, carryOverIp, ip, "ip")
	assert.InDelta(t, 0.417, rtt, acceptableDelta, "rtt")
}

func TestProcessTracerouteHopLine(t *testing.T) {
	var (
		hopNumber int
		hopInfo   []TracerouteHopInfo
		err       error
	)

	hopNumber, hopInfo, err = processTracerouteHopLine(NormalTracerouteLine)
	assert.NoError(t, err)
	assert.Equal(t, NormalTracerouteHopNumber, hopNumber, "hopNumber")
	assert.True(t, reflect.DeepEqual(NormalTracerouteHopInfo, hopInfo))

	hopNumber, hopInfo, err = processTracerouteHopLine(SomeVoidTracerouteLine)
	assert.NoError(t, err)
	assert.Equal(t, SomeVoidTracerouteHopNumber, hopNumber, "hopNumber")
	assert.True(t, reflect.DeepEqual(SomeVoidTracerouteHopInfo, hopInfo))

	hopNumber, hopInfo, err = processTracerouteHopLine(AllVoidTracerouteLine)
	assert.NoError(t, err)
	assert.Equal(t, AllVoidTracerouteHopNumber, hopNumber, "hopNumber")
	assert.True(t, reflect.DeepEqual([]TracerouteHopInfo{}, hopInfo))
}

func TestProcessTracerouteHeaderLine(t *testing.T) {
	fqdn, ip := processTracerouteHeaderLine(LinuxTracerouteHeader)
	assert.Equal(t, LinuxTracerouteFqdn, fqdn, "fqdn")
	assert.Equal(t, LinuxTracerouteIp, ip, "ip")
}
