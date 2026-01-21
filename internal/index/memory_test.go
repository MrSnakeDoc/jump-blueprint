package index

import (
	"sync"
	"testing"

	"github.com/MrSnakeDoc/jump/internal/domain"
)

func TestNewMemoryIndex(t *testing.T) {
	index := NewMemoryIndex()
	if index == nil {
		t.Fatal("NewMemoryIndex() returned nil")
	}
	services := index.GetAllServices()
	if len(services) != 0 {
		t.Errorf("NewMemoryIndex() should start with empty services, got %v", len(services))
	}
}

func TestUpdateServices(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "adguard", Name: "adguard", Hostname: "adguard.domain.ext"},
		{ID: "admin", Name: "admin", Hostname: "admin.domain.ext"},
	}

	index.UpdateServices(services)

	retrieved := index.GetAllServices()
	if len(retrieved) != 2 {
		t.Errorf("UpdateServices() stored %v services, want 2", len(retrieved))
	}
}

func TestUpdateServicesOverwrites(t *testing.T) {
	index := NewMemoryIndex()

	initial := []*domain.Service{
		{ID: "service1", Name: "service1", Hostname: "service1.example.com"},
	}
	index.UpdateServices(initial)

	updated := []*domain.Service{
		{ID: "service2", Name: "service2", Hostname: "service2.example.com"},
		{ID: "service3", Name: "service3", Hostname: "service3.example.com"},
	}
	index.UpdateServices(updated)

	retrieved := index.GetAllServices()
	if len(retrieved) != 2 {
		t.Errorf("UpdateServices() should overwrite, got %v services want 2", len(retrieved))
	}
}

func TestGetAllServices(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "service1", Name: "service1", Hostname: "service1.example.com"},
		{ID: "service2", Name: "service2", Hostname: "service2.example.com"},
	}
	index.UpdateServices(services)

	retrieved := index.GetAllServices()
	if len(retrieved) != len(services) {
		t.Errorf("GetAllServices() = %v services, want %v", len(retrieved), len(services))
	}
}

func TestIncrementCounter(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "adguard", Name: "adguard", Hostname: "adguard.domain.ext", Counter: 0},
	}
	index.UpdateServices(services)

	index.IncrementCounter("adguard")

	retrieved := index.GetAllServices()
	if len(retrieved) != 1 {
		t.Fatal("service not found")
	}
	if retrieved[0].Counter != 1 {
		t.Errorf("IncrementCounter() counter = %v, want 1", retrieved[0].Counter)
	}

	// Increment again
	index.IncrementCounter("adguard")
	retrieved = index.GetAllServices()
	if retrieved[0].Counter != 2 {
		t.Errorf("IncrementCounter() counter = %v, want 2", retrieved[0].Counter)
	}
}

func TestIncrementCounterNonExistent(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "adguard", Name: "adguard", Hostname: "adguard.domain.ext", Counter: 0},
	}
	index.UpdateServices(services)

	// Increment counter for non-existent service should not panic
	index.IncrementCounter("nonexistent")

	retrieved := index.GetAllServices()
	if retrieved[0].Counter != 0 {
		t.Errorf("IncrementCounter() on non-existent should not affect existing counter, got %v", retrieved[0].Counter)
	}
}

func TestConcurrentAccess(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "service1", Name: "service1", Hostname: "service1.example.com", Counter: 0},
		{ID: "service2", Name: "service2", Hostname: "service2.example.com", Counter: 0},
	}
	index.UpdateServices(services)

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = index.GetAllServices()
		}()
	}

	// Concurrent counter increments
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			index.IncrementCounter("service1")
		}()
	}

	wg.Wait()

	retrieved := index.GetAllServices()
	for _, svc := range retrieved {
		if svc.Hostname == "service1.example.com" {
			if svc.Counter != 100 {
				t.Errorf("Concurrent IncrementCounter() counter = %v, want 100", svc.Counter)
			}
		}
	}
}

func TestGetAllServicesReturnsSnapshot(t *testing.T) {
	index := NewMemoryIndex()

	services := []*domain.Service{
		{ID: "service1", Name: "service1", Hostname: "service1.example.com", Counter: 0},
	}
	index.UpdateServices(services)

	snapshot1 := index.GetAllServices()
	snapshot2 := index.GetAllServices()

	// The slices themselves should be different (not same memory address)
	// but the services they point to are the same
	if &snapshot1 == &snapshot2 {
		t.Error("GetAllServices() should return different slice instances")
	}

	// Verify both contain the same service
	if len(snapshot1) != 1 || len(snapshot2) != 1 {
		t.Fatal("both snapshots should contain 1 service")
	}

	// Since they point to the same service objects, modifying one affects the other
	// This is expected behavior - GetAllServices returns a new slice but same service pointers
	if snapshot1[0] != snapshot2[0] {
		t.Error("GetAllServices() should return references to the same service objects")
	}
}
