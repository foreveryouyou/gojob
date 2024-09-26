package atask

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SchedulerInterval 固定间隔调度器
type SchedulerInterval struct {
	// tasks     []ITask
	add       chan ITask
	exit      chan struct{}
	running   bool
	runningMu sync.Mutex
}

func NewSchedulerInterval() (si *SchedulerInterval) {
	si = &SchedulerInterval{}
	si.add = make(chan ITask, 100)
	return
}

func (s *SchedulerInterval) AddTask(tasks ...ITask) {
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

func (s *SchedulerInterval) runTask(ctx context.Context, t ITask) {
	go func() {
		sched := t.Schedule()
		interval := sched.Interval
		for {
			select {
			case <-ctx.Done():
				return

			default:
				func() {
					var err error
					panicErr := PanicToError(func() {
						err = t.Handle(ctx)
					})
					if panicErr != nil {
						fmt.Printf("task %s panic: %s", t.Name(), panicErr)
					} else if err != nil {
						fmt.Printf("task %s error: %s", t.Name(), err)
					}
					time.Sleep(time.Second * time.Duration(interval))
				}()
			}
		}
	}()
}
