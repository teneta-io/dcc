package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/teneta-io/dcc/internal/entities"
	"github.com/teneta-io/dcc/pkg/rabbitmq"
	"github.com/teneta-io/dcc/utils"
	"io/ioutil"
	"os"
	"time"
)

type TaskService struct {
	taskPublisher *rabbitmq.TaskPublisher
	storage       *redis.Client
}

func NewTaskService(taskPublisher *rabbitmq.TaskPublisher, storage *redis.Client) *TaskService {
	return &TaskService{
		taskPublisher: taskPublisher,
		storage:       storage,
	}
}

func (s *TaskService) Proceed(taskPath, privateKeyName string) error {
	payload, err := s.parsePayload(taskPath)

	if err != nil {
		return err
	}

	privateKey, publicKey, err := s.loadKeys(privateKeyName)

	if err != nil {
		return err
	}

	signature, err := s.sign(privateKey, payload)

	if err != nil {
		return err
	}

	task := &entities.Task{
		UUID:         uuid.New(),
		DCCSign:      signature,
		DCCPublicKey: publicKey,
		Payload:      payload,
		CreatedAt:    time.Now(),
		Status:       entities.TaskStatusNew,
	}

	data, err := json.Marshal(task)

	s.taskPublisher.Publish(data)
	s.storage.SetNX(context.Background(), task.UUID.String(), data, 0)

	return nil
}

func (s *TaskService) parsePayload(path string) (*entities.TaskPayload, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bts, _ := ioutil.ReadAll(file)
	payload := &entities.TaskPayload{}
	if err = json.Unmarshal(bts, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (s *TaskService) loadKeys(name string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := utils.LoadPrivateKeyFile(name)

	if err != nil {
		return nil, nil, err
	}

	publicKey, err := utils.LoadPublicKeyFile(name)

	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

func (s *TaskService) sign(privateKey *rsa.PrivateKey, payload *entities.TaskPayload) ([]byte, error) {
	bts, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	hash := sha512.New()
	_, err = hash.Write(bts)
	digest := hash.Sum(nil)
	if err != nil {
		return nil, err
	}

	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA512, digest)
}
