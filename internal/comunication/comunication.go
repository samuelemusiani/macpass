package comunication

import (
	"time"
)

type Request struct {
	User     string
	Mac      string
	Duration time.Duration
}
