package test

import (
	"testing"

	"github.com/drxc00/sweepy/internal/cache"
)

func TestCacheGet(t *testing.T) {
	// In memory cache (Does not write to disk)
	cache := cache.NewCache[string]()
	cache.Set("test", "test")
	defer cache.Delete("test")

	cases := []struct {
		name       string
		key        string
		expectedOk bool
		expected   string
	}{
		{
			name:     "Get existing key",
			key:      "test",
			expected: "test",
		},
		{
			name:     "Get non-existing key",
			key:      "test2",
			expected: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, _ := cache.Get(c.key)

			if actual != c.expected {
				t.Errorf("Expected %s, got %s", c.expected, actual)
			}
		})
	}
}
