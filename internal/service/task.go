package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"github.com/teneta-io/dcc/internal/entities"
	"github.com/teneta-io/dcc/pkg/rabbitmq"
	"github.com/teneta-io/dcc/utils"
	"io/ioutil"
	"os"
	"time"
)

type TaskService struct {
	taskPublisher *rabbitmq.TaskPublisher
}

func NewTaskService(taskPublisher *rabbitmq.TaskPublisher) *TaskService {
	return &TaskService{
		taskPublisher: taskPublisher,
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

	task, err := json.Marshal(&entities.Task{
		DCCSign:      signature,
		DCCPublicKey: publicKey,
		DCPSign:      "",
		DCPPublicKey: "",
		Payload:      payload,
		CreatedAt:    time.Now(),
	})

	s.taskPublisher.Publish(task)

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

//func (s *TaskService) Proceed(payload *entities.TaskPayload, publicKey, privateKey string) error {
//	bts, err := json.Marshal(&entities.Task{
//		DCCSign:      s.sign(payload, privateKey),
//		DCCPublicKey: publicKey,
//		DCPSign:      "",
//		DCPPublicKey: "",
//		Payload:      payload,
//	})
//
//	if err != nil {
//		return err
//	}
//
//	s.taskPublisher.Publish(bts)
//
//	return nil
//}

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
