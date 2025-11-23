package utils_test

import (
	"testing"

	"jenkins-monitor/internal/utils"
)

func TestParseFloat(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{
			name:        "Valid float",
			input:       "123.45",
			expected:    123.45,
			expectError: false,
		},
		{
			name:        "Invalid float",
			input:       "abc",
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := utils.ParseFloat(tc.input)
			if (err != nil) != tc.expectError {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectError, err)
			}
			if actual != tc.expected {
				t.Errorf("Expected: %f, but got: %f", tc.expected, actual)
			}
		})
	}
}

func TestGetDir(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Simple path",
			path:     "/var/lib/jenkins-monitor/processes.csv",
			expected: "/var/lib/jenkins-monitor",
		},
		{
			name:     "Root path",
			path:     "/processes.csv",
			expected: "/",
		},
		{
			name:     "Current dir",
			path:     "processes.csv",
			expected: ".",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := utils.GetDir(tc.path)
			if actual != tc.expected {
				t.Errorf("Expected: %q, but got: %q", tc.expected, actual)
			}
		})
	}
}
