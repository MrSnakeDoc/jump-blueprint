package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
)

func TestGarbageCollector_Collect(t *testing.T) {
	log := logger.New("error", false)
	memIndex := index.NewMemoryIndex()

	// Add some test services
	now := time.Now()
	services := []*domain.Service{
		{
			ID:        "active-service.example.com",
			Hostname:  "active-service.example.com",
			Name:      "active-service",
			Sources:   []string{"homepage"},
			Disabled:  false,
			UpdatedAt: now,
		},
		{
			ID:        "recently-disabled.example.com",
			Hostname:  "recently-disabled.example.com",
			Name:      "recently-disabled",
			Sources:   []string{"homepage"},
			Disabled:  true,
			UpdatedAt: now.Add(-10 * 24 * time.Hour), // Disabled 10 days ago
		},
		{
			ID:        "old-disabled.example.com",
			Hostname:  "old-disabled.example.com",
			Name:      "old-disabled",
			Sources:   []string{"homepage"},
			Disabled:  true,
			UpdatedAt: now.Add(-35 * 24 * time.Hour), // Disabled 35 days ago
		},
	}

	memIndex.UpdateServices(services)

	// Create GC with 30 day threshold
	gc := NewGarbageCollector(
		nil, // no Redis store for this test
		memIndex,
		log,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Run collection
	err := gc.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	// Check results
	allServices := memIndex.GetAllServices()

	// Should have 2 services left (active + recently disabled)
	if len(allServices) != 2 {
		t.Errorf("Expected 2 services after GC, got %d", len(allServices))
	}

	// Check that active service is still there
	if _, ok := memIndex.GetService("active-service.example.com"); !ok {
		t.Error("Active service was incorrectly removed")
	}

	// Check that recently disabled is still there
	if _, ok := memIndex.GetService("recently-disabled.example.com"); !ok {
		t.Error("Recently disabled service was incorrectly removed")
	}

	// Check that old disabled service was removed
	if _, ok := memIndex.GetService("old-disabled.example.com"); ok {
		t.Error("Old disabled service was not removed")
	}
}
