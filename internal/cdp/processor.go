package cdp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/deduplication"
	"go.uber.org/zap"
)

// EventProcessor processes file events and creates recovery points
type EventProcessor struct {
	config        CDPConfig
	dedupeManager deduplication.DeduplicationManager
	eventQueue    <-chan *FileEvent
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	isRunning     bool
	logger        *zap.Logger
	rateLimiter   *RateLimiter
}

// RateLimiter limits event processing rate
type RateLimiter struct {
	maxPerSecond int
	tokens       chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		maxPerSecond: maxPerSecond,
		tokens:       make(chan struct{}, maxPerSecond),
	}

	// Fill tokens
	for i := 0; i < maxPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	// Replenish tokens every second
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(config CDPConfig, dedupeMgr deduplication.DeduplicationManager) *EventProcessor {
	logger, _ := zap.NewDevelopment()

	return &EventProcessor{
		config:        config,
		dedupeManager: dedupeMgr,
		logger:        logger,
		rateLimiter:   NewRateLimiter(config.MaxEventsPerSecond),
	}
}

// Start begins processing events
func (ep *EventProcessor) Start(ctx context.Context, eventQueue <-chan *FileEvent, cdpEngine CDPEngine) error {
	if ep.isRunning {
		return fmt.Errorf("event processor is already running")
	}

	ep.ctx, ep.cancel = context.WithCancel(ctx)
	ep.eventQueue = eventQueue
	ep.isRunning = true

	ep.wg.Add(1)
	go ep.processEvents(cdpEngine)

	ep.logger.Info("Event processor started",
		zap.Int("max_events_per_second", ep.config.MaxEventsPerSecond))

	return nil
}

// Stop stops processing events
func (ep *EventProcessor) Stop() {
	if !ep.isRunning {
		return
	}

	ep.cancel()
	ep.wg.Wait()
	ep.isRunning = false

	ep.logger.Info("Event processor stopped")
}

// processEvents processes events from the queue
func (ep *EventProcessor) processEvents(cdpEngine CDPEngine) {
	defer ep.wg.Done()

	for {
		select {
		case <-ep.ctx.Done():
			return
		case event, ok := <-ep.eventQueue:
			if !ok {
				return
			}

			// Apply rate limiting
			if err := ep.rateLimiter.Wait(ep.ctx); err != nil {
				ep.logger.Warn("Rate limiter wait failed", zap.Error(err))
				continue
			}

			// Process event
			if err := cdpEngine.ProcessEvent(ep.ctx, event); err != nil {
				ep.logger.Error("Failed to process event",
					zap.String("event_id", event.ID),
					zap.String("path", event.Path),
					zap.Error(err))
			}
		}
	}
}

// ProcessEventSync processes a single event synchronously
func (ep *EventProcessor) ProcessEventSync(ctx context.Context, event *FileEvent, cdpEngine CDPEngine) error {
	return cdpEngine.ProcessEvent(ctx, event)
}

// BatchProcessEvents processes multiple events in batch
func (ep *EventProcessor) BatchProcessEvents(ctx context.Context, events []*FileEvent, cdpEngine CDPEngine) ([]error, error) {
	errors := make([]error, len(events))
	var wg sync.WaitGroup

	for i, event := range events {
		wg.Add(1)
		go func(idx int, ev *FileEvent) {
			defer wg.Done()
			if err := cdpEngine.ProcessEvent(ctx, ev); err != nil {
				errors[idx] = err
			}
		}(i, event)
	}

	wg.Wait()

	return errors, nil
}

// GetQueueSize returns the current queue size
func (ep *EventProcessor) GetQueueSize() int {
	return len(ep.eventQueue)
}

// GetStats returns processor statistics
func (ep *EventProcessor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"is_running":      ep.isRunning,
		"queue_size":      ep.GetQueueSize(),
		"max_per_second":  ep.config.MaxEventsPerSecond,
		"max_queue_size":  ep.config.MaxQueueSize,
		"rpo_target":      ep.config.RPOTarget.String(),
		"compression":     ep.config.CompressionEnabled,
		"encryption":      ep.config.EncryptionEnabled,
	}
}
