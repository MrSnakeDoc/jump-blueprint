package homepage

// BookmarkEntry represents a single bookmark entry in the YAML
type BookmarkEntry struct {
	Icon string `yaml:"icon"`
	Abbr string `yaml:"abbr"`
	Href string `yaml:"href"`
}

// BookmarkCategory represents a category with its bookmarks
// The YAML structure is: - CategoryName: { - BookmarkName: [{ icon, abbr, href }] }
// Each bookmark name maps to a list (array) with a single entry containing the properties
type BookmarkCategory map[string][]map[string][]BookmarkEntry

// BookmarksConfig is the root structure for bookmarks.yaml
type BookmarksConfig []BookmarkCategory
