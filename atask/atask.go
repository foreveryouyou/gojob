package atask

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

type RedisClientOpt = asynq.RedisClientOpt

// TaskManager 任务管理器
type TaskManager struct {
	logger   ILogger
	taskList []ITask

	asynqRedisOpt RedisClientOpt
	asynqServer   *asynq.Server
}

type ParamNewTM struct {
	RedisOpt RedisClientOpt
	Logger   ILogger
}

func NewTaskManager(param ParamNewTM) (tm *TaskManager) {
	tm = &TaskManager{}
	tm.taskList = make([]ITask, 0, 20)

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
	var producerList = make(chan ITask, 20)
	go func() {
		for _, v := range tm.taskList {
			if v.TaskHandler != nil {
				tm.logger.Info(logPrefix+" 添加任务: %s", v.TaskName())
				producerList <- v
			}
		}
	}()

	var runTask = func(t ITask) {
		tm.logger.Info(logPrefix+" 执行任务: %s", t.TaskName())
		if t.TaskHandler == nil {
			return
		}

		defer func() {
			if err := recover(); err != nil {
				tm.logger.Warn(logPrefix+" 任务发生 panic, 重新加入任务队列 [%s], %#v", t.TaskName(), err)
				time.Sleep(10 * time.Second)
				// 重新加入执行队列
				producerList <- t
			} else {
				tm.logger.Warn(logPrefix+" 任务正常退出: %s", t.TaskName())
			}
		}()

		tm.logger.Info(logPrefix+" 任务开始执行: %s", t.TaskName())
		t.TaskHandler(ctx)
	}

	for _task := range producerList {
		go runTask(_task)
	}
}
