package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/redis/go-redis/v9"
)

// ConnectOptions defines Redis connection retry behavior.
type ConnectOptions struct {
	Addr           string        // Redis address (ex: "localhost:6379")
	User           string        // Optional username
	Password       string        // Optional password
	RedisDB        int           // Redis DB number
	DialTimeout    time.Duration // Redis dial timeout
	ReadTimeout    time.Duration // Redis read timeout
	WriteTimeout   time.Duration // Redis write timeout
	PoolSize       int           // Redis connection pool size
	ConnectTimeout time.Duration // Total time allowed for connection attempts (ex: 30s)
	RetryInterval  time.Duration // Initial wait between retries (ex: 2s, grows exponentially)
	MaxWait        time.Duration // max wait between retries (ex: 10s)
	PingTimeout    time.Duration // timeout for each ping attempt (ex: 2s)
	WarnThreshold  int           // warn after this many attempts
}

// retryConfig holds retry policy settings.
type retryConfig struct {
	maxWait       time.Duration
	pingTimeout   time.Duration
	initialWait   time.Duration
	totalTimeout  time.Duration
	warnThreshold int // warn after this many attempts
}

// connectionLogger handles all Redis connection logging.
type connectionLogger struct {
	logger logger.Logger
}

func (cl *connectionLogger) logConnectionStart(addr string, timeout time.Duration) {
	cl.logger.Info("connecting to redis",
		logger.String("addr", addr),
		logger.Duration("timeout", timeout))
}

func (cl *connectionLogger) logSuccess(addr string, attempts int, elapsed time.Duration) {
	if attempts > 1 {
		cl.logger.Warn("connected to redis after retry",
			logger.String("addr", addr),
			logger.Int("attempts", attempts),
			logger.Duration("elapsed", elapsed))
	} else {
		cl.logger.Info("connected to redis",
			logger.String("addr", addr))
	}
}

func (cl *connectionLogger) logTimeout(addr string, attempts int, timeout time.Duration, err error) {
	cl.logger.Error("redis unavailable - failed to connect after timeout",
		logger.String("addr", addr),
		logger.Int("attempts", attempts),
		logger.Duration("timeout", timeout),
		logger.Error(err))
}

func (cl *connectionLogger) logRetry(addr string, attempt int, remaining time.Duration, nextRetry time.Duration, warnThreshold int, err error) {
	switch {
	case remaining < 10*time.Second:
		cl.logger.Error("redis still down - retrying but timeout approaching",
			logger.String("addr", addr),
			logger.Int("attempt", attempt),
			logger.Duration("remaining", remaining),
			logger.Duration("next_retry_in", nextRetry),
			logger.Error(err))
	case attempt <= warnThreshold:
		cl.logger.Warn("redis connection failed, retrying",
			logger.String("addr", addr),
			logger.Int("attempt", attempt),
			logger.Duration("next_retry_in", nextRetry),
			logger.Error(err))
	default:
		cl.logger.Error("redis still unavailable - connection attempts failing",
			logger.String("addr", addr),
			logger.Int("attempt", attempt),
			logger.Duration("next_retry_in", nextRetry),
			logger.Error(err))
	}
}

// validateOptions ensures all required configuration values are valid.
func (cl *connectionLogger) validateOptions(opts ConnectOptions) error {
	if opts.ConnectTimeout <= 0 {
		cl.logger.Error("invalid ConnectTimeout", logger.Duration("value", opts.ConnectTimeout))
		return fmt.Errorf("ConnectTimeout must be > 0, got %v", opts.ConnectTimeout)
	}
	if opts.RetryInterval <= 0 {
		cl.logger.Error("invalid RetryInterval", logger.Duration("value", opts.RetryInterval))
		return fmt.Errorf("RetryInterval must be > 0, got %v", opts.RetryInterval)
	}
	if opts.MaxWait <= 0 {
		cl.logger.Error("invalid MaxWait", logger.Duration("value", opts.MaxWait))
		return fmt.Errorf("MaxWait must be > 0, got %v", opts.MaxWait)
	}
	if opts.PingTimeout <= 0 {
		cl.logger.Error("invalid PingTimeout", logger.Duration("value", opts.PingTimeout))
		return fmt.Errorf("PingTimeout must be > 0, got %v", opts.PingTimeout)
	}
	if opts.WarnThreshold < 0 {
		cl.logger.Error("invalid WarnThreshold", logger.Int("value", opts.WarnThreshold))
		return fmt.Errorf("WarnThreshold must be >= 0, got %d", opts.WarnThreshold)
	}
	return nil
}

// New creates a new Redis client with retry logic and exponential backoff.
// It will keep retrying until ConnectTimeout is reached, logging warnings for each failed attempt.
// Returns error if connection cannot be established within the timeout.
func New(opts ConnectOptions, log logger.Logger) (*redis.Client, error) {
	connLogger := &connectionLogger{logger: log}
	if err := connLogger.validateOptions(opts); err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Username:     opts.User,
		Password:     opts.Password,
		DB:           opts.RedisDB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
	})

	retry := retryConfig{
		maxWait:       opts.MaxWait,
		pingTimeout:   opts.PingTimeout,
		initialWait:   opts.RetryInterval,
		totalTimeout:  opts.ConnectTimeout,
		warnThreshold: opts.WarnThreshold,
	}

	return connectWithRetry(client, opts.Addr, retry, connLogger)
}

// connectWithRetry handles the retry loop with exponential backoff.
func connectWithRetry(client *redis.Client, addr string, retry retryConfig, log *connectionLogger) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), retry.totalTimeout)
	defer cancel()

	log.logConnectionStart(addr, retry.totalTimeout)
	attempt := 0
	wait := retry.initialWait

	for {
		attempt++

		// Attempt connection
		pingCtx, pingCancel := context.WithTimeout(ctx, retry.pingTimeout)
		err := client.Ping(pingCtx).Err()
		pingCancel()

		if err == nil {
			elapsed := retry.totalTimeout - timeLeft(ctx)
			log.logSuccess(addr, attempt, elapsed)
			return client, nil
		}

		// Check if timeout exhausted
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			log.logTimeout(addr, attempt, retry.totalTimeout, err)
			return nil, fmt.Errorf("redis unavailable at %s after %d attempts (timeout: %v): %w",
				addr, attempt, retry.totalTimeout, err)

		case <-timer.C:
			remaining := timeLeft(ctx)
			log.logRetry(addr, attempt, remaining, wait, retry.warnThreshold, err)
			// Exponential backoff with cap
			wait *= 2
			if wait > retry.maxWait {
				wait = retry.maxWait
			}
		}
	}
}

// timeLeft returns the remaining time before context deadline.
func timeLeft(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	return time.Until(deadline)
}
