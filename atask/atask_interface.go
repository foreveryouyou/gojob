package atask

import (
	"context"

	"github.com/hibiken/asynq"
)

// ITask 基本任务定义, 所有任务都必须实现该接口
type ITask interface {
	// TaskName 任务名, 如: "图片处理"
	TaskName() string

	// Cron 定时器表达式
	TaskCron() string

	// Interval 执行间隔时间, 单位: 秒
	TaskInterval() int64

	// TaskHandler 任务执行逻辑
	TaskHandler(ctx context.Context) (err error)

	// TaskFunc 任务处理逻辑
	// TaskFunc() *TaskFunc

	// 任务队列配置, 不需要任务队列的话返回nil即可
	TaskQueue() *TaskQueue
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
