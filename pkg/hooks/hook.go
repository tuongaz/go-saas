package hooks

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/pkg/uid"
)

type Handler[T any] func(context.Context, T) error

type handlerPair[T any] struct {
	id      string
	handler Handler[T]
}

type Hook[T any] struct {
	mu            sync.RWMutex
	handlers      []*handlerPair[T]
	asyncHandlers []*handlerPair[T]
}

func (h *Hook[T]) Add(handler Handler[T]) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uid.ID()

	h.handlers = append(h.handlers, &handlerPair[T]{id, handler})

	return id
}

func (h *Hook[T]) AddRsync(handler Handler[T]) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uid.ID()

	h.asyncHandlers = append(h.asyncHandlers, &handlerPair[T]{id, handler})

	return id
}

func (h *Hook[T]) Trigger(ctx context.Context, event T) error {
	h.mu.RLock()
	handlers := make([]*handlerPair[T], len(h.handlers))
	handlers = append(handlers, h.handlers...)
	h.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.handler(ctx, event); err != nil {
			return errors.Wrap(err, "failed to trigger hook")
		}
	}

	for _, handler := range h.asyncHandlers {
		go func(handler *handlerPair[T]) {
			if err := handler.handler(ctx, event); err != nil {
				log.Error("failed to trigger async hook: %v", err)
			}
		}(handler)
	}

	return nil
}
