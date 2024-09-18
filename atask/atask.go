package atask

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

const (
	ProviderTypeDefault = iota // 默认, 由内置 cron 调度
	ProviderTypeXXLJob         // 由外部 xxl-job 调度
)

type RedisClientOpt = asynq.RedisClientOpt

// TaskManager 任务管理器
type TaskManager struct {
	logger   ILogger
	taskList []ITask

	asynqRedisOpt RedisClientOpt
	asynqServer   *asynq.Server

	providerType int // 任务调度实现方式
}

type ParamNewTM struct {
	ProviderType int // 任务调度实现方式

	RedisOpt RedisClientOpt
	Logger   ILogger
}

func NewTaskManager(param ParamNewTM) (tm *TaskManager) {
	tm = &TaskManager{}
	tm.taskList = make([]ITask, 0, 20)

	// provider
	tm.providerType = param.ProviderType
	if tm.providerType != ProviderTypeDefault && tm.providerType != ProviderTypeXXLJob {
		panic("atask provider type error")
	}

	// logger
	tm.logger = param.Logger
	if tm.logger == nil {
		tm.logger = &defaultLogger{}
	}

	// redis连接配置
	tm.asynqRedisOpt = param.RedisOpt

	return
}

func (tm *TaskManager) RedisOpt() RedisClientOpt {
	return tm.asynqRedisOpt
}

// AddTask 新增任务定义
func (tm *TaskManager) AddTask(tasks ...ITask) {
	for _, v := range tasks {
		if v == nil {
			continue
		}
		tm.taskList = append(tm.taskList, v)
	}
}

// Start 启动任务管理器
func (tm *TaskManager) Start(ctx context.Context) {
	tm.logger.Info("[TaskManager] 开始执行... %d 个任务", len(tm.taskList))

	// 启动任务处理
	go tm.handleTask(ctx)

	// 启动 asynqServer
	go tm.setupAsynqServer(ctx)

	// 等待退出信号
	go func() {
		<-ctx.Done()
		tm.logger.Info("asynq server shutdown")
		tm.asynqServer.Shutdown()
	}()
}

// 获取 asynq client 实例
func (tm *TaskManager) AsynqClient(ctx context.Context) *asynq.Client {
	client := asynq.NewClient(tm.asynqRedisOpt)
	return client
}

// 初始化 asynq server
func (tm *TaskManager) setupAsynqServer(ctx context.Context) {
	// queues 不同的任务队列
	//   With this above configuration:
	//     tasks in critical queue will be processed 60% of the time
	//     tasks in default queue will be processed 30% of the time
	//     tasks in low queue will be processed 10% of the time
	queues := map[string]int{
		"critical": 6,
		"default":  3,
		"low":      1,
	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	for _, t := range tm.taskList {
		if t == nil {
			continue
		}

		// 任务队列配置
		tq := t.TaskQueue()
		if tq == nil {
			continue
		}
		if _, ok := queues[tq.Name]; !ok {
			queues[tq.Name] = tq.Priority
		}

		// 任务队列处理器配置
		if hdl := tq.Handler; hdl == nil {
			continue
		}
		mux.HandleFunc(tq.Pattern, tq.Handler)
		tm.logger.Info("[TaskManager handler] 注册: %s, %s", t.TaskName(), tq.Pattern)
	}

	// 最大并发 workers 数
	var workers = 0
	for _, v := range queues {
		workers += v
	}
	if workers <= 0 {
		workers = 10
	}

	// 初始化 asynqServer
	tm.asynqServer = asynq.NewServer(
		tm.asynqRedisOpt,
		asynq.Config{
			BaseContext: func() context.Context { return ctx },
			// Specify how many concurrent workers to use
			Concurrency: workers,
			// Optionally specify multiple queues with different priority.
			Queues: queues,
			// Logger specifies the logger used by the server instance.
			Logger: &asynqLogger{
				logger: tm.logger,
			},
			// See the godoc for other configuration options
		},
	)

	if err := tm.asynqServer.Run(mux); err != nil {
		tm.logger.Warn("启动 asynq server 失败: %+v", err)
	}
}

func (tm *TaskManager) handleTask(ctx context.Context) {
	time.Sleep(time.Second * 1)
	logPrefix := "[TaskManager TaskHandler]"
	tm.logger.Info(logPrefix + " 开始执行...")

	switch tm.providerType {
	case ProviderTypeXXLJob:

	case ProviderTypeDefault:
		fallthrough
	default:
		hdlDefault := taskHandlerProviderDefault{
			tm: tm,
		}
		hdlDefault.handleTasks(ctx, tm.taskList)
	}
}
