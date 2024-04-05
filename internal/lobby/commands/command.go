package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/sessions"
)

type Command struct {
	ParentCtx context.Context
	user      uuid.UUID
	Err       error
	done      chan struct{}
}

func NewCommand(ctx context.Context, user uuid.UUID) *Command {
	return &Command{
		ParentCtx: ctx,
		user:      user,
		Err:       nil,
		done:      make(chan struct{}),
	}
}
func (c *Command) GetUserId() uuid.UUID {
	return c.user
}

func (c *Command) SetError(err error) {
	select {
	case <-c.done:
	default:
		if c.Err != nil {
			err = fmt.Errorf("%w: %w", c.Err, err)
		}
		c.Err = err
		c.SetDone()
	}
}

func (c *Command) SetDone() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}
func (c *Command) Done() <-chan struct{} {
	return c.done
}

func (c *Command) Execute(_ *sessions.Session) {
	panic("Execute not implemented")
}

func (c *Command) WaitForDone() error {
	select {
	case <-c.Done():
		if c.Err != nil {
			return c.Err
		}
	}
	return nil
}
