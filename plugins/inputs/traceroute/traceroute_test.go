package traceroute

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var LinuxTracerouteOutput = ``

func MockHostTracerouter(timeout float64, args ...string) (string, error) {
	return LinuxTracerouteOutput, nil
}

func TestSanity(t *testing.T) {
	var tr *Traceroute
	tr = &Traceroute{
		hostTraceRoute: MockHostTracerouter,
	}
	sampleConfig := tr.SampleConfig()
	assert.NotNil(t, sampleConfig)
	fmt.Println(sampleConfig)
}
