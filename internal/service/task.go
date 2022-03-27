package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/teneta-io/dcc/internal/entities"
	"github.com/teneta-io/dcc/pkg/rabbitmq"
	"strings"
)

type TaskService struct {
	taskPublisher *rabbitmq.TaskPublisher
}

func NewTaskService(taskPublisher *rabbitmq.TaskPublisher) *TaskService {
	return &TaskService{
		taskPublisher: taskPublisher,
	}
}

func (s *TaskService) Proceed(payload *entities.TaskPayload, publicKey, privateKey string) error {
	bts, err := json.Marshal(&entities.Task{
		DCCSign:      s.sign(payload, privateKey),
		DCCPublicKey: publicKey,
		DCPSign:      "",
		DCPPublicKey: "",
		Payload:      payload,
	})

	if err != nil {
		return err
	}

	s.taskPublisher.Publish(bts)

	return nil
}

func (s *TaskService) sign(payload *entities.TaskPayload, privateKey string) string {
	var dccSign bytes.Buffer
	binary.Write(&dccSign, binary.BigEndian, payload)
	h := hmac.New(sha512.New, []byte(privateKey))
	h.Write(dccSign.Bytes())
	sha := hex.EncodeToString(h.Sum(nil))

	return strings.ToUpper(sha)
}
