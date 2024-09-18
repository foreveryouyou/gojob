package tasks

import (
	"context"

	"github.com/foreveryouyou/gojob/atask"
)

var (
	taskManager *atask.TaskManager
	taskList    []atask.ITask = make([]atask.ITask, 0, 20)
)

func TaskList() []atask.ITask {
	return taskList
}

// RegisterTask 注册任务定义
func RegisterTask(t atask.ITask) {
	taskList = append(taskList, t)
}

func TaskManager() *atask.TaskManager {
	return taskManager
}

func Setup(opt atask.RedisClientOpt) {
	taskManager = atask.NewTaskManager(atask.ParamNewTM{
		RedisOpt: opt,
	})
	taskManager.AddTask(TaskList()...)
}

func Run(context context.Context) {
	taskManager.Start(context)
}
