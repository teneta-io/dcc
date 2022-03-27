package container

import (
	"context"
	"github.com/sarulabs/di"
	"github.com/teneta-io/dcc/internal/config"
	"github.com/teneta-io/dcc/internal/http"
	"github.com/teneta-io/dcc/internal/service"
	"github.com/teneta-io/dcc/pkg/rabbitmq"
	"github.com/teneta-io/dcc/pkg/redis"
	"go.uber.org/zap"
)

var container di.Container

func Build(ctx context.Context) di.Container {
	builder, _ := di.NewBuilder()

	if err := builder.Add([]di.Def{
		{
			Name: "Logger",
			Build: func(ctn di.Container) (i interface{}, e error) {
				return zap.NewProduction()
			},
			Close: func(obj interface{}) error {
				return obj.(*zap.Logger).Sync()
			},
		},
		{
			Name: "Config",
			Build: func(ctn di.Container) (interface{}, error) {
				return config.New()
			},
		},
		{
			Name: "Server",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)
				logger := ctn.Get("Logger").(*zap.Logger)
				taskService := ctn.Get("TaskService").(*service.TaskService)

				return http.New(ctx, &cfg.ServerConfig, logger, taskService), nil
			},
		},
		{
			Name: "Redis",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)

				return redis.New(ctx, &cfg.RedisConfig)
			},
		},
		{
			Name: "RabbitMQ",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)
				logger := ctn.Get("Logger").(*zap.Logger)

				return rabbitmq.NewClient(ctx, &cfg.RabbitMQConfig, logger)
			},
		},
		{
			Name: "TaskPublisher",
			Build: func(ctn di.Container) (interface{}, error) {
				client := ctn.Get("RabbitMQ").(*rabbitmq.RabbitMQ)
				return rabbitmq.NewTaskPublisher(client)
			},
		},
		{
			Name: "TaskService",
			Build: func(ctn di.Container) (interface{}, error) {
				taskPublisher := ctn.Get("TaskPublisher").(*rabbitmq.TaskPublisher)

				return service.NewTaskService(taskPublisher), nil
			},
		},
	}...); err != nil {
		panic(err)
	}
	container = builder.Build()

	return container
}
