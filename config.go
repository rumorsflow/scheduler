package scheduler

import "time"

type Config struct {
	Enable       bool          `mapstructure:"enable"`
	SyncInterval time.Duration `mapstructure:"sync"`
}

func (cfg *Config) InitDefault() {
	if cfg.SyncInterval == 0 {
		cfg.SyncInterval = 15 * time.Minute
	}
}
