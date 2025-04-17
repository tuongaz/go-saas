package hooks

import (
	"context"
	"fmt"
	"sync"

	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/pkg/uid"
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

// New creates a new hook
func New[T any]() *Hook[T] {
	return &Hook[T]{
		handlers:      make([]*handlerPair[T], 0),
		asyncHandlers: make([]*handlerPair[T], 0),
	}
}

func (h *Hook[T]) Add(handler Handler[T]) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uid.ID()
	h.handlers = append(h.handlers, &handlerPair[T]{
		id:      id,
		handler: handler,
	})

	return id
}

func (h *Hook[T]) AddRsync(handler Handler[T]) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uid.ID()
	h.asyncHandlers = append(h.asyncHandlers, &handlerPair[T]{
		id:      id,
		handler: handler,
	})

	return id
}

func (h *Hook[T]) Trigger(ctx context.Context, event T, oneOffHandlers ...Handler[T]) error {
	h.mu.RLock()
	handlers := make([]*handlerPair[T], len(h.handlers))
	copy(handlers, h.handlers)
	asyncHandlers := make([]*handlerPair[T], len(h.asyncHandlers))
	copy(asyncHandlers, h.asyncHandlers)
	h.mu.RUnlock()

	// Execute synchronous handlers
	for _, pair := range handlers {
		if err := pair.handler(ctx, event); err != nil {
			return fmt.Errorf("handler %s: %w", pair.id, err)
		}
	}

	// Execute one-off handlers
	for _, handler := range oneOffHandlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("one-off handler: %w", err)
		}
	}

	// Execute asynchronous handlers
	for _, pair := range asyncHandlers {
		go func(p *handlerPair[T]) {
			if err := p.handler(ctx, event); err != nil {
				log.Default().Error("async handler error", "id", p.id, "err", err)
			}
		}(pair)
	}

	return nil
}
