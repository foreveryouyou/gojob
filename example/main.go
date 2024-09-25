package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/foreveryouyou/gojob/atask"
	"github.com/foreveryouyou/gojob/example/tasks"
	"github.com/xxl-job/xxl-job-executor-go"
)

func main() {

	// 监听关停信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	// 监听外部终止程序的信号
	go func() {
		sig := <-sigs
		fmt.Printf("%s, waiting...\n", sig)
		cancel()
	}()

	// 任务管理器
	tasks.Setup(atask.ParamNewTM{
		ProviderType: atask.ProviderTypeXXLJob,
		RedisOpt: atask.RedisClientOpt{
			Addr:     "127.0.0.1:26379",
			DB:       0,
			Password: "",
		},
		XXLJobExcutor: func() xxl.Executor {
			return xxl.NewExecutor(
				xxl.ServerAddr("http://127.0.0.1:58080/xxl-job-admin"),
				xxl.AccessToken("abcdefg"), //请求令牌(默认为空)
				// xxl.ExecutorIp("127.0.0.1"),    //可自动获取
				xxl.ExecutorPort("9998"),       //默认9999（非必填）
				xxl.RegistryKey("golang-jobs"), //执行器名称
				xxl.SetLogger(&logger{}),       //自定义日志
			)
		},
	})
	tasks.Run(ctx)

	// 等待退出
	<-ctx.Done()
}

// xxl.Logger接口实现
type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	log.Printf("自定义日志 - "+format, a...)
}
