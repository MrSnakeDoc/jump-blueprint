package homepage

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing of Homepage services.yaml
type Loader struct {
	filePath string
}

// NewLoader creates a new Homepage loader
func NewLoader(filePath string) *Loader {
	return &Loader{
		filePath: filePath,
	}
}

// Load reads and parses the services.yaml file
func (l *Loader) Load() (ServicesConfig, error) {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read services file: %w", err)
	}

	// Strip Homepage template variables ({{HOMEPAGE_VAR_...}})
	// These are not needed for Jump's purposes
	data = stripTemplateVariables(data)

	var config ServicesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse services yaml: %w", err)
	}

	return config, nil
}

// stripTemplateVariables removes Homepage template variables from YAML
// Example: {{HOMEPAGE_VAR_ADGUARD_USER}} -> ""
func stripTemplateVariables(data []byte) []byte {
	// Match {{...}} patterns
	re := regexp.MustCompile(`\{\{[^}]+\}\}`)
	return re.ReplaceAll(data, []byte(`""`))
}
