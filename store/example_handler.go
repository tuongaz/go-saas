package store

import (
	"context"
	"fmt"

	"github.com/tuongaz/go-saas/store/events"
)

// ExampleHandler is an example of how to implement a database event handler
type ExampleHandler struct{}

// NewExampleHandler creates a new example handler
func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

// OnBeforeCreate is called before a record is created
func (h *ExampleHandler) OnBeforeCreate(ctx context.Context, event *events.OnBeforeRecordCreatedEvent) error {
	fmt.Printf("Before creating record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// OnAfterCreate is called after a record is created
func (h *ExampleHandler) OnAfterCreate(ctx context.Context, event *events.OnAfterRecordCreatedEvent) error {
	fmt.Printf("After creating record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// OnBeforeUpdate is called before a record is updated
func (h *ExampleHandler) OnBeforeUpdate(ctx context.Context, event *events.OnBeforeRecordUpdatedEvent) error {
	fmt.Printf("Before updating record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// OnAfterUpdate is called after a record is updated
func (h *ExampleHandler) OnAfterUpdate(ctx context.Context, event *events.OnAfterRecordUpdatedEvent) error {
	fmt.Printf("After updating record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// OnBeforeDelete is called before a record is deleted
func (h *ExampleHandler) OnBeforeDelete(ctx context.Context, event *events.OnBeforeRecordDeletedEvent) error {
	fmt.Printf("Before deleting record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// OnAfterDelete is called after a record is deleted
func (h *ExampleHandler) OnAfterDelete(ctx context.Context, event *events.OnAfterRecordDeletedEvent) error {
	fmt.Printf("After deleting record in table %s: %+v\n", event.Table, event.Record)
	return nil
}

// Example usage:
/*
func main() {
    logger := log.New(os.Stdout, "", log.LstdFlags)
    handler := NewExampleHandler(logger)

    store := store.New("postgres://...")
    store.AddEventHandler(handler)

    // Now all operations will trigger the events
    record := store.Record{
        "name": "John Doe",
        "email": "john@example.com",
    }

    _, err := store.Collection("users").CreateRecord(context.Background(), record)
    if err != nil {
        log.Fatal(err)
    }
}
*/
