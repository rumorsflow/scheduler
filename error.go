package scheduler

import (
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func (p *Plugin) errorHandler(task *asynq.Task, opts []asynq.Option, err error) {
	p.log.Error(
		"handle scheduler error",
		zap.Error(err),
		zap.String("task", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.Any("opts", opts),
	)
}
