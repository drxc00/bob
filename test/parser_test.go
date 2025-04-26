package test

import (
	"testing"

	"github.com/drxc00/bob/utils"
)

func TestParseStalenessFlagValue(t *testing.T) {
	tests := []struct {
		name        string
		staleness   string
		expectedInt int64
		expectedErr bool
	}{
		{
			name:        "Parse 1 day",
			staleness:   "1d",
			expectedInt: 1,
			expectedErr: false,
		},
		{
			name:        "Parse 1 hour",
			staleness:   "1h",
			expectedInt: 0,
			expectedErr: false,
		},
		{
			name:        "Parse 48 hours",
			staleness:   "48h",
			expectedInt: 2,
			expectedErr: false,
		},
		{
			name:        "Parse 1 minute",
			staleness:   "1m",
			expectedInt: 0,
			expectedErr: false,
		},
		{
			name:        "Parse 1 second",
			staleness:   "1s",
			expectedInt: 0,
			expectedErr: false,
		},
		{
			name:        "No units specified",
			staleness:   "1",
			expectedInt: 1,
			expectedErr: true,
		},
		{
			name:        "No units specified",
			staleness:   "24",
			expectedInt: 24,
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualInt, err := utils.ParseStalenessFlagValue(test.staleness)

			if test.expectedErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if actualInt != test.expectedInt {
					t.Errorf("Expected %d, but got %d", test.expectedInt, actualInt)
				}
			}
		})
	}
}
