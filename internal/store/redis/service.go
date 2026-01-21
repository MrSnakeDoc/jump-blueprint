package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MrSnakeDoc/jump/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultServiceTTL is the default TTL for service entries (48 hours)
	DefaultServiceTTL = 48 * time.Hour
	// DefaultCacheTTL is the default TTL for cached resolutions (24 hours)
	DefaultCacheTTL = 24 * time.Hour
)

// Store handles Redis operations for services and cache
type Store struct {
	client *redis.Client
}

// NewStore creates a new Redis store
func NewStore(client *redis.Client) *Store {
	return &Store{
		client: client,
	}
}

// SaveService stores a service in Redis
func (s *Store) SaveService(ctx context.Context, service *domain.Service) error {
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %w", err)
	}

	key := ServiceKey(service.ID)

	// Store service data
	if err := s.client.Set(ctx, key, data, DefaultServiceTTL).Err(); err != nil {
		return fmt.Errorf("failed to save service: %w", err)
	}

	// Add to set of all services
	if err := s.client.SAdd(ctx, AllServicesKey(), service.ID).Err(); err != nil {
		return fmt.Errorf("failed to add service to set: %w", err)
	}

	return nil
}

// GetService retrieves a service from Redis by ID
func (s *Store) GetService(ctx context.Context, id string) (*domain.Service, error) {
	key := ServiceKey(id)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("service not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	var service domain.Service
	if err := json.Unmarshal(data, &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service: %w", err)
	}

	return &service, nil
}

// GetAllServices retrieves all services from Redis
func (s *Store) GetAllServices(ctx context.Context) ([]*domain.Service, error) {
	// Get all service IDs
	ids, err := s.client.SMembers(ctx, AllServicesKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service IDs: %w", err)
	}

	if len(ids) == 0 {
		return []*domain.Service{}, nil
	}

	// Retrieve all services in parallel
	services := make([]*domain.Service, 0, len(ids))
	for _, id := range ids {
		service, err := s.GetService(ctx, id)
		if err != nil {
			// Skip services that couldn't be retrieved
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

// DeleteService removes a service from Redis
func (s *Store) DeleteService(ctx context.Context, id string) error {
	key := ServiceKey(id)

	// Delete service data
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Remove from set of all services
	if err := s.client.SRem(ctx, AllServicesKey(), id).Err(); err != nil {
		return fmt.Errorf("failed to remove service from set: %w", err)
	}

	return nil
}

// UpdateServiceCounter increments the usage counter for a service
func (s *Store) UpdateServiceCounter(ctx context.Context, id string) error {
	service, err := s.GetService(ctx, id)
	if err != nil {
		return err
	}

	service.Counter++
	service.LastSeenAt = time.Now()

	return s.SaveService(ctx, service)
}

// SaveServicesMany stores multiple services in Redis (bulk operation)
func (s *Store) SaveServicesMany(ctx context.Context, services []*domain.Service) error {
	pipe := s.client.Pipeline()

	for _, service := range services {
		data, err := json.Marshal(service)
		if err != nil {
			return fmt.Errorf("failed to marshal service %s: %w", service.ID, err)
		}

		key := ServiceKey(service.ID)
		pipe.Set(ctx, key, data, DefaultServiceTTL)
		pipe.SAdd(ctx, AllServicesKey(), service.ID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save services: %w", err)
	}

	return nil
}
