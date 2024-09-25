package atask

import (
	"context"

	"github.com/hibiken/asynq"
)

// ITask 基本任务定义, 所有任务都必须实现该接口
type ITask interface {
	// ID 任务唯一标识, 如: "imageResize", "image.resize"
	ID() string

	// Name 任务显示名, 仅用于显示, 无业务逻辑, 如: "图片处理"
	Name() string

	// Schedule 任务调度配置
	Schedule() Schedule

	// Handle 任务执行逻辑
	Handle(ctx context.Context, args ...any) (err error)

	// TaskFunc 任务处理逻辑
	// TaskFunc() *TaskFunc

	// 任务队列配置, 不需要任务队列的话返回nil即可
	TaskQueue() *TaskQueue
}

type Schedule struct {
	// 调度类型: cron, fixedInterval
	Type ScheduleType

	// 调度配置
	//   调度类型为cron, 值为 "cron表达式"
	//   调度类型为fixedInterval, 值为 "执行间隔时间, 单位: 秒"
	Conf string

	// Cron 定时器表达式
	Cron string

	// Interval 执行间隔时间, 单位: 秒
	Interval int64
}

// RunReq 任务执行请求参数
type RunReq struct {
}

type TaskFunc struct {

	// HandleTask 任务执行逻辑
	HandleTask func(ctx context.Context) (err error)
}

// TaskQueue 任务队列配置
type TaskQueue struct {
	// 队列名称, 如: "default"
	Name string

	// 队列标识, 如: "image:resize"
	Pattern string

	// 优先级, 值越大优先级越高
	Priority int

	// 队列处理函数
	Handler func(ctx context.Context, t *asynq.Task) error
}
