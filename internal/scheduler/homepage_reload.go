package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/MrSnakeDoc/jump/internal/sources/homepage"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
)

// HomepageReloader handles periodic reloading of homepage services
type HomepageReloader struct {
	loader        *homepage.Loader
	mapper        *homepage.Mapper
	store         *redisstore.Store
	index         *index.MemoryIndex
	logger        logger.Logger
	interval      time.Duration
	stopCh        chan struct{}
	manualTrigger chan struct{}
}

// NewHomepageReloader creates a new homepage reloader
func NewHomepageReloader(
	serviceFile string,
	store *redisstore.Store,
	idx *index.MemoryIndex,
	log logger.Logger,
	interval time.Duration,
	manualTrigger chan struct{},
) *HomepageReloader {
	return &HomepageReloader{
		loader:        homepage.NewLoader(serviceFile),
		mapper:        homepage.NewMapper(),
		store:         store,
		index:         idx,
		logger:        log,
		interval:      interval,
		stopCh:        make(chan struct{}),
		manualTrigger: manualTrigger,
	}
}

// Start begins the periodic reload process
func (hr *HomepageReloader) Start(ctx context.Context) error {
	// Load immediately on start
	if err := hr.Reload(ctx); err != nil {
		return fmt.Errorf("initial reload failed: %w", err)
	}

	// Start periodic reload
	ticker := time.NewTicker(hr.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := hr.Reload(ctx); err != nil {
					hr.logger.Error("failed to reload services",
						logger.Error(err))
				}
			case <-hr.manualTrigger:
				hr.logger.Info("manual reload triggered")
				if err := hr.Reload(ctx); err != nil {
					hr.logger.Error("failed to reload services",
						logger.Error(err))
				}
			case <-hr.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop stops the reloader
func (hr *HomepageReloader) Stop() {
	close(hr.stopCh)
}

// Reload loads services from homepage and updates store + index
func (hr *HomepageReloader) Reload(ctx context.Context) error {
	hr.logger.Info("reloading services from homepage")

	// Load and parse services.yaml
	config, err := hr.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	// Map to domain services
	newServices, err := hr.mapper.MapServices(config)
	if err != nil {
		return fmt.Errorf("failed to map services: %w", err)
	}

	hr.logger.Info("loaded services from homepage",
		logger.Int("count", len(newServices)))

	// Get existing services from homepage source to detect removals
	existingServices := hr.getHomepageServices()

	// Build map of new service IDs for quick lookup
	newServiceIDs := make(map[string]bool, len(newServices))
	for _, svc := range newServices {
		newServiceIDs[svc.ID] = true
	}

	// Find services that were removed from homepage
	var disabledServices []*domain.Service
	for _, existing := range existingServices {
		if !newServiceIDs[existing.ID] {
			// Service no longer in homepage - mark as disabled
			existing.Disabled = true
			existing.UpdatedAt = time.Now()
			disabledServices = append(disabledServices, existing)
		}
	}

	if len(disabledServices) > 0 {
		hr.logger.Info("marking removed services as disabled",
			logger.Int("count", len(disabledServices)))
	}

	// Combine active and disabled services for storage
	newServices = append(newServices, disabledServices...)

	// Update memory index
	hr.index.UpdateServices(newServices)

	// Update Redis store (best effort)
	if hr.store != nil {
		if err := hr.store.SaveServicesMany(ctx, newServices); err != nil {
			hr.logger.Warn("failed to save services to redis",
				logger.Error(err))
			// Don't fail - memory index is the primary source
		} else {
			hr.logger.Info("services saved to redis")
		}
	}

	return nil
}

// getHomepageServices returns existing services that came from homepage source
func (hr *HomepageReloader) getHomepageServices() []*domain.Service {
	all := hr.index.GetAllServices()
	var homepageServices []*domain.Service

	for _, svc := range all {
		// Check if service has homepage in its sources
		for _, source := range svc.Sources {
			if source == "homepage" {
				homepageServices = append(homepageServices, svc)
				break
			}
		}
	}

	return homepageServices
}
