package config

import (
	"os"
	"testing"
	"time"
)

func TestRequireEnv(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		shouldSet bool
		wantPanic bool
	}{
		{
			name:      "variable set",
			key:       "TEST_VAR",
			value:     "test_value",
			shouldSet: true,
			wantPanic: false,
		},
		{
			name:      "variable not set",
			key:       "TEST_VAR_MISSING",
			shouldSet: false,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSet {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("failed to unset env var: %v", err)
					}
				}()
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("requireEnv() should have panicked")
					}
				}()
			}

			result := requireEnv(tt.key)
			if !tt.wantPanic && result != tt.value {
				t.Errorf("requireEnv() = %v, want %v", result, tt.value)
			}
		})
	}
}

func TestRequireEnvInt(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		expected  int
		wantPanic bool
	}{
		{
			name:      "valid integer",
			key:       "TEST_INT",
			value:     "42",
			expected:  42,
			wantPanic: false,
		},
		{
			name:      "invalid integer",
			key:       "TEST_INT_INVALID",
			value:     "not_a_number",
			wantPanic: true,
		},
		{
			name:      "missing variable",
			key:       "TEST_INT_MISSING",
			value:     "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("failed to unset env var: %v", err)
					}
				}()
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("requireEnvInt() should have panicked")
					}
				}()
			}

			result := requireEnvInt(tt.key)
			if !tt.wantPanic && result != tt.expected {
				t.Errorf("requireEnvInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRequireEnvSlice(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		expected  []string
		wantPanic bool
	}{
		{
			name:      "single value",
			key:       "TEST_SLICE",
			value:     "value1",
			expected:  []string{"value1"},
			wantPanic: false,
		},
		{
			name:      "multiple values",
			key:       "TEST_SLICE_MULTI",
			value:     "value1, value2, value3",
			expected:  []string{"value1", "value2", "value3"},
			wantPanic: false,
		},
		{
			name:      "missing variable",
			key:       "TEST_SLICE_MISSING",
			value:     "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("failed to unset env var: %v", err)
					}
				}()
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("requireEnvSlice() should have panicked")
					}
				}()
			}

			result := requireEnvSlice(tt.key)
			if !tt.wantPanic {
				if len(result) != len(tt.expected) {
					t.Errorf("requireEnvSlice() length = %v, want %v", len(result), len(tt.expected))
				}
				for i := range result {
					if result[i] != tt.expected[i] {
						t.Errorf("requireEnvSlice()[%d] = %v, want %v", i, result[i], tt.expected[i])
					}
				}
			}
		})
	}
}

func TestExtractDomains(t *testing.T) {
	tests := []struct {
		name     string
		hosts    []string
		expected []string
	}{
		{
			name:     "simple hostname",
			hosts:    []string{"jump.domain.ext"},
			expected: []string{"jump.domain.ext", "domain.ext"},
		},
		{
			name:     "hostname with port",
			hosts:    []string{"10.70.80.2:8080"},
			expected: []string{"10.70.80.2", "70.80.2"},
		},
		{
			name:     "multiple hostnames",
			hosts:    []string{"jump.domain.ext", "api.domain.ext"},
			expected: []string{"jump.domain.ext", "domain.ext", "api.domain.ext"},
		},
		{
			name:     "empty slice",
			hosts:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomains(tt.hosts)
			if len(result) != len(tt.expected) {
				t.Errorf("extractDomains() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			// Check that all expected domains are present (order may vary)
			for _, exp := range tt.expected {
				found := false
				for _, res := range result {
					if res == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("extractDomains() missing expected domain: %v", exp)
				}
			}
		})
	}
}

func TestMustDuration(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		def      time.Duration
		expected time.Duration
	}{
		{
			name:     "valid duration",
			key:      "TEST_DURATION",
			value:    "5s",
			def:      1 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "invalid duration uses default",
			key:      "TEST_DURATION_INVALID",
			value:    "invalid",
			def:      10 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "missing variable uses default",
			key:      "TEST_DURATION_MISSING",
			value:    "",
			def:      15 * time.Second,
			expected: 15 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("failed to unset env var: %v", err)
					}
				}()
			}

			result := mustDuration(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("mustDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMustBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		def      bool
		expected bool
	}{
		{
			name:     "true value",
			key:      "TEST_BOOL",
			value:    "true",
			def:      false,
			expected: true,
		},
		{
			name:     "false value",
			key:      "TEST_BOOL_FALSE",
			value:    "false",
			def:      true,
			expected: false,
		},
		{
			name:     "invalid value uses default",
			key:      "TEST_BOOL_INVALID",
			value:    "invalid",
			def:      true,
			expected: true,
		},
		{
			name:     "missing variable uses default",
			key:      "TEST_BOOL_MISSING",
			value:    "",
			def:      false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Errorf("failed to unset env var: %v", err)
					}
				}()
			}

			result := mustBool(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("mustBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}
