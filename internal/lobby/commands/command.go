package commands

import (
	"context"

	"github.com/google/uuid"
)

type command struct {
	ctx  context.Context
	user uuid.UUID
	Err  chan error
}

func newCommand(
	ctx context.Context,
	user uuid.UUID,
) *command {
	return &command{
		ctx:  ctx,
		user: user,
		Err:  make(chan error),
	}
}
func (c *command) GetUserId() uuid.UUID {
	return c.user
}

func (c *command) Fail(err error) {
	select {
	case <-c.ctx.Done():
	default:
		c.Err <- err
	}
}
