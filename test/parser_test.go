package test

import (
	"testing"

	"github.com/drxc00/sweepy/utils"
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
			staleness:   "1",
			expectedInt: 1,
			expectedErr: false,
		},
		{
			name:        "Parse 1 with units",
			staleness:   "1day",
			expectedInt: 0,
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
