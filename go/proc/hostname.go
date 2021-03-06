package proc

import (
	"os"
	"sync"
)

const (
	unknown       = "unknown"
	LabelHostname = "hostname"
)

var (
	hostName string = unknown
	hostOnce sync.Once
)

// GetHostname runs and saves a hostname, once
func GetHostname() string {
	hostOnce.Do(func() {
		x, err := os.Hostname()
		if err == nil {
			hostName = x
		}
	})

	return hostName
}
