package entities

import "time"

const (
	TaskStatusNew         Status = "new"
	TaskStatusInitialized Status = "initialized"
	TaskStatusRunning     Status = "running"
	TaskStatusFinished    Status = "finished"
)

type TaskPayload struct {
	Link         string       `json:"link"`
	PriceLimit   uint         `json:"price_limit"`
	Requirements Requirements `json:"requirements"`
	Status       Status       `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	ExpiredAt    time.Time    `json:"expired_at"`
}

type Task struct {
	DCCSign      string       `json:"dcc_sign"`
	DCCPublicKey string       `json:"dcc_public_key"`
	DCPSign      string       `json:"dcp_sign"`
	DCPPublicKey string       `json:"dcp_public_key"`
	Payload      *TaskPayload `json:"payload"`
}

type Status string

type Requirements struct {
	VCPU    uint8 `json:"vcpu"`
	RAM     uint8 `json:"ram"`
	Storage uint8 `json:"storage"`
	Network uint8 `json:"network"`
}
