package main

import (
	"context"
	"flag"
	"github.com/teneta-io/dcc/internal/container"
	"github.com/teneta-io/dcc/internal/service"
	"go.uber.org/zap"
)

var taskPath, privateKeyName string

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
	}()

	c := container.Build(ctx)

	flag.StringVar(&taskPath, "task", "", "json file path")
	flag.StringVar(&privateKeyName, "private-key", "", "private key path")
	flag.Parse()

	if taskPath == "" {
		zap.L().Fatal("invalid task json path")
	}

	if privateKeyName == "" {
		zap.L().Fatal("invalid private key path")
	}

	zap.S().Info("task creating...")

	s := c.Get("TaskService").(*service.TaskService)

	if err := s.Proceed(taskPath, privateKeyName); err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Info("done.")
}
