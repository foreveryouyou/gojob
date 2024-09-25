package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/foreveryouyou/gojob/atask"
	"github.com/hibiken/asynq"
)

func init() {
	// 注册当前任务
	RegisterTask(&TaskDemo{})
}

// TaskDemo 任务定义, 实现了 atask.ITask 接口
type TaskDemo struct {
}

// ID 任务唯一标识, 如: "imageResize"
func (t *TaskDemo) ID() string {
	return "demo:task"
}

// Name 任务名称, 仅用于给人看的
func (t *TaskDemo) Name() string {
	return "demo测试任务"
}

// Schedule 任务调度配置
func (t *TaskDemo) Schedule() atask.Schedule {
	return atask.Schedule{
		Type: atask.ScheduleTypeCron,
		Conf: "* * * * * *",
	}
}

// Handle 任务执行逻辑
func (t *TaskDemo) Handle(ctx context.Context, args ...any) error {
	taskId := strconv.FormatInt(time.Now().UnixMilli(), 10)
	taskPayload := []byte("hello" + time.Now().Format(time.DateTime))

	tq := t.TaskQueue()
	if tq == nil {
		return errors.New("队列配置不完整")
	}

	asynqTask := asynq.NewTask(
		tq.Pattern,                     // 任务标识
		taskPayload,                    // 任务内容
		asynq.TaskID(taskId),           // 任务ID, 唯一标识, 这里不传则默认uuid, 如果业务有需要这里就传自定义的任务id, 重复的任务不会重复插入
		asynq.ProcessIn(time.Second*2), // 延迟2秒执行
		asynq.MaxRetry(3),              // 最大重试次数
		asynq.Timeout(0),               // 超时设置, 0 表示不超时, 如果超过超时时间, 自动自动重试
	)
	client := TaskManager().AsynqClient(ctx)
	defer client.Close()
	info, err := client.EnqueueContext(ctx, asynqTask, asynq.Queue(tq.Name))
	if err != nil {
		fmt.Printf("插入队列失败: %v\n", err)
	} else {
		fmt.Printf("插入队列成功: id=%s queue=%s]\n", info.ID, info.Queue)
	}

	time.Sleep(time.Second * 2)

	return nil
}

// TaskQueue 任务队列配置
func (t *TaskDemo) TaskQueue() *atask.TaskQueue {
	return &atask.TaskQueue{
		Name:     "demoQueue", // 任务队列名称
		Pattern:  "demo:task", // 队列标识
		Priority: 3,           // 任务队列优先级
		Handler:  t.queueHandler,
	}
}

// queueHandler 任务队列消费逻辑
func (t *TaskDemo) queueHandler(ctx context.Context, at *asynq.Task) (err error) {
	fmt.Printf("[%s] 处理队列任务, id=%s, payload=%s\n", t.Name(), at.ResultWriter().TaskID(), at.Payload())

	/* if time.Now().UnixMilli()%2 == 0 {
		// 返回错误的情况, 会自动重试, 直到用完重试次数
		return errors.New("测试失败的情况")
	} */

	// 没有错误, 则表示任务成功, 会自动从队列中清除
	time.Sleep(time.Second * 2)
	return nil
}
