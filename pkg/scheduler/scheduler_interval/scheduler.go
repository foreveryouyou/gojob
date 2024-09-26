package scheduler_interval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/foreveryouyou/gojob/pkg/logger"
	"github.com/foreveryouyou/gojob/pkg/utils"
)

// SchedulerInterval 固定间隔调度器
type SchedulerInterval struct {
	// tasks     []Task
	add       chan Task
	exit      chan struct{}
	running   bool
	runningMu sync.Mutex

	logger logger.ILogger
}

// Task 任务
type Task struct {
	// ID 任务唯一标识, 如: "image.resize"
	ID string

	// Handle 任务处理函数
	//   返回 err 不为 nil 表示本次调度失败, 会根据 RetryMax 进行重试
	Handle func(ctx context.Context) (err error)

	// Interval 调度间隔
	Interval time.Duration

	// RetryMax 重试次数
	RetryMax int
	// retryTimes 重试次数
	retryTimes int
}

type Option func(si *SchedulerInterval)

func WithLogger(logger logger.ILogger) Option {
	return func(si *SchedulerInterval) {
		si.logger = logger
	}
}

func NewSchedulerInterval(opts ...Option) (si *SchedulerInterval) {
	si = &SchedulerInterval{}
	si.add = make(chan Task, 100)

	for _, opt := range opts {
		opt(si)
	}
	return
}

func (s *SchedulerInterval) AddTask(tasks ...Task) {
	go func() {
		for _, t := range tasks {
			s.add <- t
		}
	}()
}

// Start the scheduler in its own goroutine, or no-op if already started.
func (s *SchedulerInterval) Start(ctx context.Context) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		return
	}
	s.running = true
	go s.run(ctx)
}

// Run the scheduler, or no-op if already running.
func (s *SchedulerInterval) Run(ctx context.Context) {
	s.runningMu.Lock()
	if s.running {
		s.runningMu.Unlock()
		return
	}
	s.running = true
	s.runningMu.Unlock()
	s.run(ctx)
}

func (s *SchedulerInterval) Stop() {
	<-s.exit
}

func (s *SchedulerInterval) run(ctx context.Context) {
	defer func() {
		s.exit <- struct{}{}
	}()

	for {
		select {
		case t := <-s.add:
			s.runTask(ctx, t)

		case <-ctx.Done():
			return
		}
	}
}

func (s *SchedulerInterval) runTask(ctx context.Context, t Task) {
	go func() {
		interval := t.Interval
		if interval <= 0 {
			interval = time.Second
		}

		for {
			select {
			case <-ctx.Done():
				return

			default:
				func() {
					s.logger.Info("task %s handle", t.ID)

					var err error
					panicErr := utils.PanicToError(func() {
						err = t.Handle(ctx)
					})
					if panicErr != nil {
						fmt.Printf("task %s panic: %s", t.ID, panicErr)
						t.retryTimes++
					} else if err != nil {
						fmt.Printf("task %s error: %s", t.ID, err)
						t.retryTimes++
					}
					if t.retryTimes >= t.RetryMax {
						s.logger.Warn("task %s retry max %d times, stop", t.ID, t.RetryMax)
						return
					}

					time.Sleep(interval)
				}()
			}
		}
	}()
}
