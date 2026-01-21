package homepage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoad(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "services.yaml")

	yamlContent := `---
- Infrastructure:
    - AdGuard Home:
        icon: adguard-home.svg
        href: https://adguard.domain.ext
        description: Network-wide ads & trackers blocking DNS server
`

	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	loader := NewLoader(yamlPath)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(config) == 0 {
		t.Fatal("Load() returned empty config")
	}
}

func TestLoaderLoadWithTemplateVariables(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "services.yaml")

	yamlContent := `---
- Infrastructure:
    - AdGuard Home:
        icon: adguard-home.svg
        href: {{HOMEPAGE_VAR_ADGUARD_URL}}
        description: Test
`

	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	loader := NewLoader(yamlPath)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(config) == 0 {
		t.Fatal("Load() returned empty config")
	}
}

func TestLoaderLoadFileNotFound(t *testing.T) {
	loader := NewLoader("/nonexistent/path/services.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() with non-existent file should return error")
	}
}

func TestStripTemplateVariablesFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "single template variable",
			input:    []byte("url: {{HOMEPAGE_VAR_URL}}"),
			expected: "url: \"\"",
		},
		{
			name:     "no template variables",
			input:    []byte("plain text"),
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripTemplateVariables(tt.input)
			if string(result) != tt.expected {
				t.Errorf("stripTemplateVariables() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}
