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
	hdl.si = scheduler_interval.NewSchedulerInterval(
		scheduler_interval.WithLogger(hdl.tm.logger),
	)
	hdl.si.Start(ctx)

	hdl.tm.logger.Info("[TaskHandlerProviderDefault] Start...")

	for _, v := range taskList {
		sched := v.Schedule()
		switch sched.Type {
		case ScheduleTypeCron:
			hdl.tm.logger.Info("添加任务: [%s], cron: [%s]", v.Name(), sched.Conf)
			cronExpr, err := sched.Cron()
			if err != nil {
				hdl.tm.logger.Error("无效的cron表达式: [%s], err: %s", sched.Conf, err.Error())
				continue
			}
			hdl.cron.AddFunc(cronExpr, func() {
				v.Handle(ctx)
			})

		case ScheduleTypeFixedInterval:
			hdl.tm.logger.Info("添加任务: [%s], interval: [%s]", v.Name(), sched.Conf)
			interval, err := sched.Interval()
			if err != nil {
				hdl.tm.logger.Error("无效的interval: [%s], err: %s", sched.Conf, err.Error())
				continue
			}
			entry := scheduler_interval.Task{
				Handle: func(ctx context.Context) (err error) {
					return v.Handle(ctx)
				},
				ID:       v.ID(),
				Interval: time.Second * time.Duration(interval),
				RetryMax: 0,
			}
			hdl.si.AddTask(entry)

		default:
			hdl.tm.logger.Error("不支持的任务调度类型: [%d]", sched.Type)
		}
	}

	hdl.cron.Start()
}
