package atask

import (
	"context"

	"github.com/xxl-job/xxl-job-executor-go"
)

// taskHandlerProviderXXLJob xxl-job任务处理器
type taskHandlerProviderXXLJob struct {
	xxlJobExcutor xxl.Executor
	tm            *TaskManager
}

func (hdl *taskHandlerProviderXXLJob) handleTasks(ctx context.Context, taskList []ITask) {

	exec := hdl.xxlJobExcutor
	exec.Init()

	for _, v := range taskList {
		hdl.tm.logger.Info("添加任务: %s", v.TaskName())
		exec.RegTask(v.TaskName(), func(cxt context.Context, param *xxl.RunReq) (msg string) {
			v.TaskHandler(ctx)
			return
		})
	}

	exec.Run()
}
