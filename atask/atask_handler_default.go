package atask

import (
	"context"
	"time"

	"github.com/foreveryouyou/gojob/pkg/scheduler/scheduler_interval"
	"github.com/robfig/cron/v3"
)

// taskHandlerProviderDefault 默认任务处理器
type taskHandlerProviderDefault struct {
	cron *cron.Cron
	tm   *TaskManager

	si *scheduler_interval.SchedulerInterval
}

func (hdl *taskHandlerProviderDefault) handleTasks(ctx context.Context, taskList []ITask) {

	hdl.cron = cron.New(cron.WithSeconds())
	hdl.si = scheduler_interval.NewSchedulerInterval()
	hdl.si.Start(ctx)

	for _, v := range taskList {
		hdl.tm.logger.Info("添加任务: %s", v.Name())
		sched := v.Schedule()
		switch sched.Type {
		case ScheduleTypeCron:
			hdl.tm.logger.Info("添加任务: %s, cron: %s", v.Name(), sched.Cron)
			hdl.cron.AddFunc(sched.Cron, func() {
				v.Handle(ctx)
			})
		case ScheduleTypeFixedInterval:
			hdl.tm.logger.Info("添加任务: %s, interval: %d s", v.Name(), sched.Interval)
			entry := scheduler_interval.Task{
				Handle: func(ctx context.Context) (err error) {
					return v.Handle(ctx)
				},
				ID:       v.ID(),
				Interval: time.Second * time.Duration(sched.Interval),
				RetryMax: 0,
			}
			hdl.si.AddTask(entry)
		default:
			hdl.tm.logger.Error("不支持的任务调度类型: %d", sched.Type)
		}
	}

	hdl.cron.Start()
}
