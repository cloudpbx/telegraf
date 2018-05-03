package traceroute

import (
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var LinuxTracerouteOutput = `
traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets
 1  165.227.32.254 (165.227.32.254)  1.206 ms 165.227.32.253 (165.227.32.253)  1.188 ms 165.227.32.254 (165.227.32.254)  1.143 ms
 2  138.197.249.78 (138.197.249.78)  0.985 ms 138.197.249.86 (138.197.249.86)  0.939 ms 138.197.249.90 (138.197.249.90)  1.181 ms
 3  72.14.219.10 (72.14.219.10)  0.818 ms 162.243.190.33 (162.243.190.33)  0.952 ms  0.941 ms
 4  108.170.250.225 (108.170.250.225)  0.825 ms  0.970 ms 108.170.250.241 (108.170.250.241)  1.007 ms
 5  108.170.226.217 (108.170.226.217)  0.995 ms 108.170.226.219 (108.170.226.219)  1.033 ms 108.170.226.217 (108.170.226.217)  1.003 ms
 6  dfw06s38-in-f14.1e100.net (172.217.0.238)  1.187 ms  0.722 ms  0.545 ms
`

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
func TestHostTraceRouterSimple(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{}
	args := tr.args("google.com")
	_, err := hostTraceRouter(3, args...)
	if err != nil {
		t.Fatal("call failed:", err)
	}
}
*/

func TestHostTraceRouteBadHost(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{}
	args := tr.args("badhost")
	_, err := hostTraceRouter(3, args...)
	assert.Error(t, err)
}

func TestProcessOutputSimple(t *testing.T) {
	numHops, err := processTracerouteOutput(LinuxTracerouteOutput)
	assert.NoError(t, err)
	assert.Equal(t, 6, numHops, "6 hops made by packet")
}
