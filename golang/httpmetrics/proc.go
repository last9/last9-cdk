package httpmetrics

import (
	"os"
	"path/filepath"
	"sync"
)

const unknown = "unknown"

var (
	hostName           string = unknown
	programName        string = unknown
	progOnce, hostOnce sync.Once
)

func getHostname() string {
	hostOnce.Do(func() {
		x, err := os.Hostname()
		if err == nil {
			hostName = x
		}
	})

	return hostName
}

func getProgamName() string {
	progOnce.Do(func() {
		x, err := os.Executable()
		if err == nil {
			programName = filepath.Base(x)
		}
	})

	return programName
}
