package replication

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/danielino/comio/internal/monitoring"
)

const (
	// Small objects (<1MB) are replicated inline
	InlineDataThreshold = 1024 * 1024
)

type Replicator struct {
	config    Config
	client    *http.Client
	queue     chan Event
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	stats     Stats
}

type Stats struct {
	EventsQueued    int64
	EventsReplicated int64
	EventsFailed    int64
	LastReplication time.Time
}

func NewReplicator(config Config) *Replicator {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Replicator{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		queue:  make(chan Event, 10000), // Buffer 10k events
		ctx:    ctx,
		cancel: cancel,
	}
}

func (r *Replicator) Start() error {
	if !r.config.Enabled {
		monitoring.Log.Info("Replication disabled")
		return nil
	}

	monitoring.Log.Info("Starting replicator",
		zap.String("remote", r.config.RemoteURL),
		zap.String("mode", string(r.config.Mode)))

	// Start worker goroutines
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		r.wg.Add(1)
		go r.worker(i)
	}

	return nil
}

func (r *Replicator) Stop() {
	monitoring.Log.Info("Stopping replicator")
	r.cancel()
	close(r.queue)
	r.wg.Wait()
	monitoring.Log.Info("Replicator stopped")
}

func (r *Replicator) QueueEvent(event Event) {
	if !r.config.Enabled {
		return
	}

	// Generate ID if not set
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d-%s-%s", time.Now().UnixNano(), event.Bucket, event.Key)
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case r.queue <- event:
		r.mu.Lock()
		r.stats.EventsQueued++
		r.mu.Unlock()
	default:
		monitoring.Log.Warn("Replication queue full, dropping event",
			zap.String("event_id", event.ID))
		r.mu.Lock()
		r.stats.EventsFailed++
		r.mu.Unlock()
	}
}

func (r *Replicator) worker(id int) {
	defer r.wg.Done()

	monitoring.Log.Info("Replication worker started", zap.Int("worker_id", id))

	batch := make([]Event, 0, r.config.BatchSize)
	ticker := time.NewTicker(r.config.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			// Flush remaining events
			if len(batch) > 0 {
				r.sendBatch(batch)
			}
			return

		case event, ok := <-r.queue:
			if !ok {
				return
			}
			batch = append(batch, event)
			
			if len(batch) >= r.config.BatchSize {
				r.sendBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				r.sendBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (r *Replicator) sendBatch(events []Event) {
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		if err := r.sendEvent(event); err != nil {
			monitoring.Log.Error("Failed to replicate event",
				zap.String("event_id", event.ID),
				zap.Error(err))
			r.mu.Lock()
			r.stats.EventsFailed++
			r.mu.Unlock()
		} else {
			r.mu.Lock()
			r.stats.EventsReplicated++
			r.stats.LastReplication = time.Now()
			r.mu.Unlock()
		}
	}
}

func (r *Replicator) sendEvent(event Event) error {
	var err error
	for attempt := 0; attempt <= r.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(r.config.RetryDelay)
			monitoring.Log.Info("Retrying event replication",
				zap.String("event_id", event.ID),
				zap.Int("attempt", attempt))
		}

		switch event.Type {
		case EventPutObject:
			err = r.replicatePutObject(event)
		case EventDeleteObject:
			err = r.replicateDeleteObject(event)
		case EventPurgeBucket:
			err = r.replicatePurgeBucket(event)
		default:
			return fmt.Errorf("unknown event type: %s", event.Type)
		}

		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", r.config.RetryAttempts+1, err)
}

func (r *Replicator) replicatePutObject(event Event) error {
	url := fmt.Sprintf("%s/%s/%s", r.config.RemoteURL, event.Bucket, event.Key)

	var body io.Reader
	if len(event.Data) > 0 {
		// Inline data
		body = bytes.NewReader(event.Data)
	} else if event.DataURL != "" {
		// Fetch from local URL
		resp, err := http.Get(event.DataURL)
		if err != nil {
			return fmt.Errorf("failed to fetch object data: %w", err)
		}
		defer resp.Body.Close()
		body = resp.Body
	} else {
		return fmt.Errorf("no data or data URL provided")
	}

	req, err := http.NewRequestWithContext(r.ctx, "PUT", url, body)
	if err != nil {
		return err
	}

	if r.config.RemoteToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.config.RemoteToken)
	}

	if contentType, ok := event.Metadata["content_type"].(string); ok {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (r *Replicator) replicateDeleteObject(event Event) error {
	url := fmt.Sprintf("%s/%s/%s", r.config.RemoteURL, event.Bucket, event.Key)

	req, err := http.NewRequestWithContext(r.ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	if r.config.RemoteToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.config.RemoteToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (r *Replicator) replicatePurgeBucket(event Event) error {
	url := fmt.Sprintf("%s/admin/%s/objects", r.config.RemoteURL, event.Bucket)

	req, err := http.NewRequestWithContext(r.ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	if r.config.RemoteToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.config.RemoteToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (r *Replicator) GetStats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats
}
