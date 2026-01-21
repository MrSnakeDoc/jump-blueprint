package homepage

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
)

// Mapper converts Homepage services to domain.Service entities
type Mapper struct{}

// NewMapper creates a new mapper instance
func NewMapper() *Mapper {
	return &Mapper{}
}

// MapServices converts Homepage ServicesConfig to []domain.Service
func (m *Mapper) MapServices(config ServicesConfig) ([]*domain.Service, error) {
	var services []*domain.Service
	now := time.Now()

	// Iterate through groups
	for _, groupMap := range config {
		for groupName, servicesList := range groupMap {
			_ = groupName // Group name available if needed

			// Iterate through services in this group
			for _, serviceMap := range servicesList {
				for serviceName, props := range serviceMap {
					_ = serviceName // Service name available if needed

					// Skip services without href
					if props.Href == "" {
						continue
					}

					// Parse URL to extract hostname
					parsedURL, err := url.Parse(props.Href)
					if err != nil {
						// Skip invalid URLs
						continue
					}

					hostname := parsedURL.Hostname()
					if hostname == "" {
						continue
					}

					// Extract service name from first DNS label (subdomain)
					name := extractServiceName(hostname)

					service := &domain.Service{
						ID:         hostname,
						Hostname:   hostname,
						Name:       name,
						Sources:    []string{"homepage"},
						LastSeenAt: now,
						Counter:    0,
					}

					services = append(services, service)
				}
			}
		}
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no valid services found in homepage config")
	}

	return services, nil
}

// extractServiceName extracts the first DNS label as service name
// Example: "jellyfin.domain.ext" -> "jellyfin"
func extractServiceName(hostname string) string {
	parts := strings.Split(hostname, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return hostname
}
