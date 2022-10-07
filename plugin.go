package scheduler

import (
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	rdb "github.com/rumorsflow/contracts/redis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

const PluginName = "scheduler"

type Plugin struct {
	mu sync.RWMutex

	cfg     *Config
	manager *asynq.PeriodicTaskManager
	opts    asynq.PeriodicTaskManagerOpts
	log     *zap.Logger
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger, client redis.UniversalClient, provider asynq.PeriodicTaskConfigProvider) error {
	const op = errors.Op("scheduler plugin init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	var err error
	if err = cfg.UnmarshalKey(PluginName, &p.cfg); err != nil {
		return errors.E(op, errors.Init, err)
	}

	p.cfg.InitDefault()

	if !p.cfg.Enable {
		return errors.E(op, errors.Disabled)
	}

	p.log = log
	p.opts = asynq.PeriodicTaskManagerOpts{
		PeriodicTaskConfigProvider: provider,
		RedisConnOpt:               rdb.NewProxy(client),
		SyncInterval:               p.cfg.SyncInterval,
		SchedulerOpts: &asynq.SchedulerOpts{
			Logger:              log.Sugar(),
			LogLevel:            logLevel(zapcore.LevelOf(log.Core())),
			EnqueueErrorHandler: p.errorHandler,
		},
	}

	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("scheduler plugin serve")

	errCh := make(chan error, 1)

	var err error
	p.manager, err = asynq.NewPeriodicTaskManager(p.opts)
	if err != nil {
		errCh <- errors.E(op, errors.Serve, err)
		return errCh
	}

	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		if err := p.manager.Start(); err != nil {
			errCh <- errors.E(op, errors.Serve, err)
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.manager.Shutdown()

	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}
