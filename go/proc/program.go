package proc

import (
	"os"
	"path/filepath"
	"sync"
)

const LabelProgram = "program"

var (
	programName string = unknown
	progOnce    sync.Once
)

func GetProgamName() string {
	progOnce.Do(func() {
		x, err := os.Executable()
		if err == nil {
			programName = filepath.Base(x)
		}
	})

	return programName
}
