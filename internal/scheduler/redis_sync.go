package scheduler

import (
	"context"

	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
)

// RedisSyncer syncs services from Redis to memory index on startup
type RedisSyncer struct {
	store  *redisstore.Store
	index  *index.MemoryIndex
	logger logger.Logger
}

// NewRedisSyncer creates a new Redis syncer
func NewRedisSyncer(
	store *redisstore.Store,
	idx *index.MemoryIndex,
	log logger.Logger,
) *RedisSyncer {
	return &RedisSyncer{
		store:  store,
		index:  idx,
		logger: log,
	}
}

// Sync loads services from Redis and updates memory index
func (rs *RedisSyncer) Sync(ctx context.Context) error {
	rs.logger.Info("syncing services from redis to memory")

	services, err := rs.store.GetAllServices(ctx)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		rs.logger.Info("no services found in redis")
		return nil
	}

	rs.index.UpdateServices(services)

	rs.logger.Info("synced services from redis",
		logger.Int("count", len(services)))

	return nil
}
