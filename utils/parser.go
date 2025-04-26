package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

func ParseStalenessFlagValue(stalenessFlag string) (int64, error) {
	// We want to parse the value of the staleness flag as an int64
	// We accept the following formats:
	// - 1d
	// - 1h
	// - 1m
	// - 1s
	pattern := `^(\d+)(d|h|m|s)$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(stalenessFlag)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid staleness flag format: %s", stalenessFlag)
	}

	stalenessFlagInt, err := strconv.ParseInt(matches[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid staleness flag format: %s", stalenessFlag)
	}

	switch matches[2] {
	case "d":
		// Already in days, do nothing
	case "h":
		stalenessFlagInt /= 24 // Convert hours to days
	case "m":
		stalenessFlagInt /= (60 * 24) // Convert minutes to days
	case "s":
		stalenessFlagInt /= (60 * 60 * 24) // Convert seconds to days
	default:
		return 0, fmt.Errorf("invalid staleness flag format: %s", stalenessFlag)
	}

	return stalenessFlagInt, nil
}
