package atask

import (
	"context"

	"github.com/robfig/cron/v3"
)

// taskHandlerProviderDefault 默认任务处理器
type taskHandlerProviderDefault struct {
	cron *cron.Cron
	tm   *TaskManager
}

func (hdl *taskHandlerProviderDefault) handleTasks(ctx context.Context, taskList []ITask) {

	hdl.cron = cron.New(cron.WithSeconds())

	for _, v := range taskList {
		hdl.tm.logger.Info("添加任务: %s", v.Name())
		sched := v.Schedule()
		// todo 非cron类型的处理
		hdl.cron.AddFunc(sched.Cron, func() {
			v.Handle(ctx)
		})
	}

	hdl.cron.Start()
}
