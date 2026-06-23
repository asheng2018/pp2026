package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/webhook/event"
)

type Queue interface {
	Publish(subject string, data interface{}) error
	Subscribe(queue, subject string, handler func(subject string, data []byte) error) error
}

type EventDispatcher struct {
	queue    Queue
	handlers map[string]event.EventHandler
}

func NewEventDispatcher(queue Queue) *EventDispatcher {
	return &EventDispatcher{
		queue:    queue,
		handlers: make(map[string]event.EventHandler),
	}
}

func (d *EventDispatcher) Register(eventType string, handler event.EventHandler) {
	d.handlers[eventType] = handler
	log.Info().Str("event_type", eventType).Msg("event handler registered")
}

func (d *EventDispatcher) Dispatch(ctx context.Context, gateway, eventType string, rawData []byte) error {
	evt := &event.PaymentEvent{
		ID:        fmt.Sprintf("evt_%s", time.Now().Format("20060102150405")),
		Gateway:   gateway,
		EventType: eventType,
		RawData:   rawData,
		CreatedAt: time.Now(),
	}

	// Publish to message queue for async processing
	subject := fmt.Sprintf("events.%s.%s", gateway, eventType)
	if err := d.queue.Publish(subject, evt); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	// Also process synchronously if handler exists
	if handler, ok := d.handlers[eventType]; ok {
		// Clone context with timeout to prevent goroutine leaks
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := handler.Handle(bgCtx, evt); err != nil {
				log.Error().Err(err).Str("event_type", eventType).Msg("handler error")
				// Publish to retry queue for later reprocessing
				d.queue.Publish("events.retry."+eventType, evt)
			}
		}()
	}

	return nil
}

func (d *EventDispatcher) Start(ctx context.Context) error {
	// Subscribe to all event types
	for eventType := range d.handlers {
		subject := fmt.Sprintf("events.*.%s", eventType)
		queueName := fmt.Sprintf("webhook_processor_%s", eventType)
		if err := d.queue.Subscribe(queueName, subject, d.messageHandler); err != nil {
			return fmt.Errorf("subscribe %s: %w", subject, err)
		}
	}
	log.Info().Msg("event dispatcher started")
	return nil
}

func (d *EventDispatcher) messageHandler(subject string, data []byte) error {
	var evt event.PaymentEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if handler, ok := d.handlers[evt.EventType]; ok {
		ctx := context.Background()
		return handler.Handle(ctx, &evt)
	}
	return nil
}
