package cfgmodel

import (
	"errors"
	"testing"
)

func TestCheckDuration(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected error
	}{
		{
			name:     "valid",
			value:    "1s",
			expected: nil,
		},
		{
			name:     "invalid_type",
			value:    123,
			expected: errors.New("must be a string"),
		},
		{
			name:     "invalid_duration",
			value:    "invalid",
			expected: errors.New("invalid duration"),
		},
		{
			name:     "empty_string",
			value:    "",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkDuration(test.value)
			if test.expected != nil {
				if err == nil {
					t.Errorf("Expected error: %v, got nil", test.expected)
				} else if err.Error() != test.expected.Error() {
					t.Errorf("Expected error: %v, got: %v", test.expected, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
