package kafka

import (
	"context"
)

type Message struct {
	Ctx   context.Context
	Key   string
	Val   string
	Topic string
}
