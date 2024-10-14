package comunication

import (
	"time"
)

type Request struct {
	User     string        `json:"user"`
	Mac      string        `json:"mac"`
	Duration time.Duration `json:"duration"`
}
