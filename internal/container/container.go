package container

import (
	"context"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/sarulabs/di"
	"github.com/teneta-io/dcc/internal/config"
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
				l, err := zap.NewDevelopment()

				if err != nil {
					return nil, err
				}

				zap.ReplaceGlobals(l)

				return l, nil
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
			Name: "Redis",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)

				return redis.New(ctx, cfg.RedisConfig)
			},
		},
		{
			Name: "RabbitMQ",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)

				return rabbitmq.NewClient(ctx, cfg.RabbitMQConfig)
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
				client := ctn.Get("Redis").(*redis2.Client)

				return service.NewTaskService(taskPublisher, client), nil
			},
		},
	}...); err != nil {
		panic(err)
	}
	container = builder.Build()
	container.Get("Logger")

	return container
}
