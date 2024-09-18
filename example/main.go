package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/foreveryouyou/gojob/atask"
	"github.com/foreveryouyou/gojob/example/tasks"
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
	tasks.Setup(atask.RedisClientOpt{
		Addr:     "127.0.0.1:26379",
		DB:       0,
		Password: "",
	})
	tasks.Run(ctx)

	// 等待退出
	<-ctx.Done()
}
