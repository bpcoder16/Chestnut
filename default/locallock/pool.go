package locallock

import (
	"context"
	"time"

	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/bpcoder16/Chestnut/v2/modules/lock/localpool"
)

var pool *localpool.RWLockPool[string]

func Pool() *localpool.RWLockPool[string] {
	if pool == nil {
		panic("local pool not init")
	}
	return pool
}

func Run(ctx context.Context, configPath string) {
	config := loadConfig(configPath)
	pool = localpool.NewRWLockPool[string](
		time.Duration(config.TTLSec)*time.Second,
		time.Duration(config.SweepSec)*time.Second,
	)

	select {
	case <-ctx.Done():
		logit.Context(ctx).InfoW("localPool.Manager.Run", "Context cancelled, localPool preparing to shutdown")
	}

	logit.Context(ctx).InfoW("localPool.Manager.Run", "localPool shutdown...")
	pool.Close()
	logit.Context(ctx).InfoW("localPool.Manager.Run", "localPool shutdown completed, exited")
}
