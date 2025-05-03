package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

func ParseStalenessFlagValue(stalenessFlag string) (int64, error) {
	// Make sure that the staleness value is integer.
	pattern := `^\d+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(stalenessFlag) {
		return 0, fmt.Errorf("staleness flag must be an integer")
	}

	// Convert the staleness value to an integer
	staleness, err := strconv.ParseInt(stalenessFlag, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("staleness flag must be an integer")
	}

	return staleness, nil
}
