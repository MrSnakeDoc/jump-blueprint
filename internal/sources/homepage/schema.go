package homepage

// ServicesConfig represents the top-level structure of services.yaml
// Homepage uses dynamic keys, so we parse as []map[string][]map[string]ServiceProps
type ServicesConfig []map[string][]map[string]ServiceProps

// ServiceProps contains the actual service properties
type ServiceProps struct {
	Href        string                 `yaml:"href"`
	Icon        string                 `yaml:"icon,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Target      string                 `yaml:"target,omitempty"`
	Ping        string                 `yaml:"ping,omitempty"`
	SiteMonitor string                 `yaml:"siteMonitor,omitempty"`
	Widget      map[string]interface{} `yaml:"widget,omitempty"`
}
