package cloud

import (
	"path/filepath"
)

type CPI struct {
	JobPath     string
	JobsDir     string
	PackagesDir string
}

func (j CPI) ExecutablePath() string {
	return filepath.Join(j.JobPath, "bin", "cpi")
}
