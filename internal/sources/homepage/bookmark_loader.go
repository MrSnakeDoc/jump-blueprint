package homepage

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BookmarkLoader handles loading and parsing of Homepage bookmarks.yaml
type BookmarkLoader struct {
	filePath string
}

// NewBookmarkLoader creates a new Homepage bookmark loader
func NewBookmarkLoader(filePath string) *BookmarkLoader {
	return &BookmarkLoader{
		filePath: filePath,
	}
}

// Load reads and parses the bookmarks.yaml file
func (l *BookmarkLoader) Load() (BookmarksConfig, error) {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	// Strip Homepage template variables ({{HOMEPAGE_VAR_...}})
	data = stripTemplateVariables(data)

	var config BookmarksConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks yaml: %w", err)
	}

	return config, nil
}
