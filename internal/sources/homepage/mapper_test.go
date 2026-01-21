package homepage

import (
	"testing"
)

func TestMapperMapServices(t *testing.T) {
	config := ServicesConfig{
		{
			"Infrastructure": []map[string]ServiceProps{
				{
					"AdGuard Home": {
						Icon:        "adguard-home.svg",
						Href:        "https://adguard.domain.ext",
						Description: "Network-wide ads blocking",
					},
				},
				{
					"Traefik": {
						Icon:        "traefik.svg",
						Href:        "https://traefik.domain.ext",
						Description: "Cloud Native Application Proxy",
					},
				},
			},
		},
	}

	mapper := NewMapper()
	services, err := mapper.MapServices(config)
	if err != nil {
		t.Fatalf("MapServices() error = %v", err)
	}

	if len(services) != 2 {
		t.Errorf("MapServices() returned %v services, want 2", len(services))
	}

	// Check first service
	found := false
	for _, svc := range services {
		if svc.Hostname == "adguard.domain.ext" {
			found = true
			if svc.Name != "adguard" {
				t.Errorf("service Name = %v, want adguard", svc.Name)
			}
		}
	}
	if !found {
		t.Error("MapServices() did not find adguard.domain.ext")
	}
}

func TestMapperMapServicesEmptyConfig(t *testing.T) {
	config := ServicesConfig{}
	mapper := NewMapper()
	services, err := mapper.MapServices(config)

	// Empty config should return an error
	if err == nil {
		t.Error("MapServices() with empty config should return error")
	}

	if services != nil {
		t.Errorf("MapServices() with empty config should return nil services, got %v", len(services))
	}
}

func TestMapperMapServicesInvalidURL(t *testing.T) {
	config := ServicesConfig{
		{
			"Test": []map[string]ServiceProps{
				{
					"Invalid Service": {
						Icon:        "test.svg",
						Href:        "not-a-valid-url",
						Description: "Invalid URL",
					},
				},
			},
		},
	}

	mapper := NewMapper()
	services, err := mapper.MapServices(config)

	// Should return error if no valid services
	if err == nil {
		t.Error("MapServices() should return error when no valid services found")
	}

	if services != nil {
		t.Errorf("MapServices() should return nil when no valid services, got %v services", len(services))
	}
}

func TestMapperMapServicesMultipleGroups(t *testing.T) {
	config := ServicesConfig{
		{
			"Group1": []map[string]ServiceProps{
				{
					"Service1": {
						Href: "https://service1.example.com",
					},
				},
			},
		},
		{
			"Group2": []map[string]ServiceProps{
				{
					"Service2": {
						Href: "https://service2.example.com",
					},
				},
			},
		},
	}

	mapper := NewMapper()
	services, err := mapper.MapServices(config)
	if err != nil {
		t.Fatalf("MapServices() error = %v", err)
	}

	if len(services) != 2 {
		t.Errorf("MapServices() returned %v services, want 2", len(services))
	}
}
