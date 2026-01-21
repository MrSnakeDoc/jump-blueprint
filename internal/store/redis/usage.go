package redis

import (
	"context"
	"fmt"
)

// IncrementUsage increments the usage counter for a service
func (s *Store) IncrementUsage(ctx context.Context, serviceID string) error {
	return s.UpdateServiceCounter(ctx, serviceID)
}

// GetUsageStats retrieves usage statistics for all services
func (s *Store) GetUsageStats(ctx context.Context) (map[string]int64, error) {
	services, err := s.GetAllServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	stats := make(map[string]int64, len(services))
	for _, service := range services {
		stats[service.ID] = service.Counter
	}

	return stats, nil
}
