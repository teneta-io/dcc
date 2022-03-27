package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/teneta-io/dcc/internal/config"
	"github.com/teneta-io/dcc/internal/entities"
	"github.com/teneta-io/dcc/internal/http/requests"
	"github.com/teneta-io/dcc/internal/service"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	ctx         context.Context
	logger      *zap.Logger
	server      *http.Server
	taskService *service.TaskService
}

func New(ctx context.Context, cfg *config.ServerConfig, logger *zap.Logger, taskService *service.TaskService) *Server {
	return &Server{
		ctx:    ctx,
		logger: logger.Named("HTTP"),
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			ReadTimeout:       0,
			ReadHeaderTimeout: 0,
			WriteTimeout:      0,
			IdleTimeout:       0,
			MaxHeaderBytes:    0,
		},
		taskService: taskService,
	}
}

func (s *Server) Run() {
	s.logger.Info("Starting http server...", zap.Any("address", s.server.Addr))

	http.HandleFunc("/healthcheck", s.healthCheckHandler)
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/task", s.taskHandler)

	if err := http.ListenAndServe(s.server.Addr, nil); err != nil {
		s.logger.Fatal("server error %v", zap.Error(err))
	}
}

func (s *Server) Shutdown() error {
	s.logger.Info("Shutdown HTTP server...")

	if err := s.server.Shutdown(s.ctx); err != nil {
		s.logger.Fatal("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server stopped.")
	return nil
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func (s *Server) taskHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	request := &requests.TaskRequest{}

	if err := decoder.Decode(&request); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.taskService.Proceed(&entities.TaskPayload{
		Link:       request.Link,
		PriceLimit: request.PriceLimit,
		Requirements: entities.Requirements{
			VCPU:    request.Requirements.VCPU,
			RAM:     request.Requirements.RAM,
			Storage: request.Requirements.Storage,
			Network: request.Requirements.Network,
		},
		Status:    entities.TaskStatusNew,
		CreatedAt: time.Now(),
		ExpiredAt: request.ExpiredAt,
	}, request.PublicKey, request.PrivateKey); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
