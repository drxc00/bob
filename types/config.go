package types

import (
	"fmt"
	"os"

	"github.com/drxc00/bob/utils"
)

type ScanContext struct {
	Staleness  int64
	NoCache    bool
	ResetCache bool
	Path       string
}

func NewScanContext(path string, staleness string, noCache bool, resetCache bool) ScanContext {
	var stalenessFlagInt int64
	var err error

	if staleness == "0" {
		stalenessFlagInt = 0
	} else {
		stalenessFlagInt, err = utils.ParseStalenessFlagValue(staleness)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing staleness flag: %v\n", err)
			os.Exit(1)
		}
	}

	return ScanContext{
		Path:       path,
		Staleness:  stalenessFlagInt,
		NoCache:    noCache,
		ResetCache: resetCache,
	}
}
