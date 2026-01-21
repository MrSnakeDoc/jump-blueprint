package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MrSnakeDoc/jump/internal/config"
	"github.com/MrSnakeDoc/jump/internal/httpserver"
	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
	"github.com/MrSnakeDoc/jump/internal/index"
	"github.com/MrSnakeDoc/jump/internal/logger"
	"github.com/MrSnakeDoc/jump/internal/redis"
	"github.com/MrSnakeDoc/jump/internal/scheduler"
	redisstore "github.com/MrSnakeDoc/jump/internal/store/redis"
	"github.com/MrSnakeDoc/jump/internal/version"
)

type App struct {
	cfg              *config.Config
	logger           logger.Logger
	server           *httpserver.Server
	redisClient      *goredis.Client
	memIndex         *index.MemoryIndex
	reloader         *scheduler.HomepageReloader
	bookmarkReloader *scheduler.BookmarkReloader
	gc               *scheduler.GarbageCollector
}

func New() *App {
	cfg := config.Load()

	loggerClient := logger.New(cfg.LogLevel, cfg.PrettyLog)

	// Initialize Redis early - fail fast if unavailable
	loggerClient.Infof("Connecting to Redis at %s", cfg.RedisAddr)
	redisClient, err := redis.New(redis.ConnectOptions{
		Addr:           cfg.RedisAddr,
		User:           cfg.RedisUser,
		Password:       cfg.RedisPassword,
		RedisDB:        cfg.RedisDB,
		DialTimeout:    cfg.RedisDT,
		ReadTimeout:    cfg.RedisRT,
		WriteTimeout:   cfg.RedisWT,
		PoolSize:       cfg.RedisPoolSize,
		ConnectTimeout: cfg.RedisConnectTimeout,
		RetryInterval:  cfg.RedisRetryInterval,
		MaxWait:        cfg.RedisMaxWait,
		PingTimeout:    cfg.RedisPingTimeout,
		WarnThreshold:  cfg.RedisWarnThreshold,
	}, loggerClient)
	if err != nil {
		loggerClient.Errorf("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	loggerClient.Info("Redis initialized successfully")

	// Initialize memory index
	memIndex := index.NewMemoryIndex()

	// Initialize Redis store
	store := redisstore.NewStore(redisClient)

	// Try to sync services from Redis to memory on startup
	syncer := scheduler.NewRedisSyncer(store, memIndex, loggerClient)
	if err := syncer.Sync(context.Background()); err != nil {
		loggerClient.Warn("failed to sync from redis on startup, will load from homepage",
			logger.Error(err))
	}

	// Create manual reload trigger channel
	reloadTrigger := make(chan struct{}, 1)

	// Initialize homepage reloader
	reloader := scheduler.NewHomepageReloader(
		cfg.ServiceFile,
		store,
		memIndex,
		loggerClient,
		cfg.ReloadInterval,
		reloadTrigger,
	)

	// Initialize garbage collector
	gc := scheduler.NewGarbageCollector(
		store,
		memIndex,
		loggerClient,
		cfg.GCInterval,
		scheduler.DefaultGCThreshold,
	)

	// Initialize bookmark reloader (if bookmark file is configured)
	var bookmarkReloader *scheduler.BookmarkReloader
	var bookmarkReloadTrigger chan struct{}
	if cfg.BookmarkFile != "" {
		loggerClient.Info("bookmark file configured, initializing bookmark reloader",
			logger.String("file", cfg.BookmarkFile))
		bookmarkReloadTrigger = make(chan struct{}, 1)
		bookmarkReloader = scheduler.NewBookmarkReloader(
			cfg.BookmarkFile,
			store,
			memIndex,
			loggerClient,
			cfg.ReloadInterval,
			bookmarkReloadTrigger,
		)
	} else {
		loggerClient.Info("bookmark file not configured, bookmark search disabled")
	}

	// Dependencies passed to routes (extend as needed).
	d := deps.Deps{
		Logger:                loggerClient,
		StartTime:             time.Now(),
		Version:               version.Version,
		Commit:                version.Commit,
		BuildDate:             version.BuildDate,
		GoVersion:             version.GoVersion,
		TimeNow:               time.Now,
		AllowedHosts:          cfg.AllowedHosts,
		AllowedCIDRS:          cfg.AllowedCIDRS,
		TrustProxy:            cfg.TrustProxy,
		ServiceFile:           cfg.ServiceFile,
		RedisClient:           redisClient,
		MemoryIndex:           memIndex,
		HomepageURL:           cfg.HomepageURL,
		TLSTimeout:            cfg.TLSTimeout,
		SkipTLSValidation:     cfg.SkipTLSValidation,
		MaxCandidates:         cfg.MaxCandidates,
		AllowedDomains:        cfg.AllowedDomains,
		ReloadTrigger:         reloadTrigger,
		BookmarkReloadTrigger: bookmarkReloadTrigger,
	}

	server := httpserver.New(cfg, loggerClient, d)

	return &App{
		cfg:              cfg,
		logger:           loggerClient,
		server:           server,
		redisClient:      redisClient,
		memIndex:         memIndex,
		reloader:         reloader,
		bookmarkReloader: bookmarkReloader,
		gc:               gc,
	}
}

func (a *App) Run() error {
	a.logger.Infof("🚀 Starting Jump v%s on %s", version.Version, a.cfg.ListenPort)
	a.logger.Infof("Jump %s (commit=%s, built=%s, go=%s)",
		version.Version, version.Commit, version.BuildDate, version.GoVersion)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start homepage reloader (loads services and starts periodic refresh)
	if err := a.reloader.Start(ctx); err != nil {
		return fmt.Errorf("failed to start homepage reloader: %w", err)
	}
	a.logger.Info("homepage reloader started",
		logger.Duration("interval", a.cfg.ReloadInterval))

	// Start bookmark reloader (if enabled)
	if a.bookmarkReloader != nil {
		if err := a.bookmarkReloader.Start(ctx); err != nil {
			return fmt.Errorf("failed to start bookmark reloader: %w", err)
		}
		a.logger.Info("bookmark reloader started",
			logger.Duration("interval", a.cfg.ReloadInterval))
	}

	// Start garbage collector
	if err := a.gc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start garbage collector: %w", err)
	}
	a.logger.Info("garbage collector started",
		logger.Duration("interval", a.cfg.GCInterval))

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.Start(); err != nil {
			errCh <- fmt.Errorf("http server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("⏳ Shutting down gracefully...")
	case err := <-errCh:
		return err
	}

	// Stop reloader
	a.reloader.Stop()

	// Stop bookmark reloader
	if a.bookmarkReloader != nil {
		a.bookmarkReloader.Stop()
	}

	// Stop garbage collector
	a.gc.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()
	if err := a.server.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Warnf("failed to close redis: %v", err)
		} else {
			a.logger.Info("✅ Redis closed cleanly")
		}
	}

	a.logger.Info("✅ Jump stopped cleanly")
	return nil
}
