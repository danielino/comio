package replication

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielino/comio/internal/monitoring"
)

func init() {
	monitoring.InitLogger("test", "info", "console")
}

func TestReplicator_QueueEvent(t *testing.T) {
	config := Config{
		Enabled:       true,
		Mode:          ModeAsync,
		BatchSize:     10,
		BatchInterval: 100 * time.Millisecond,
	}

	replicator := NewReplicator(config)

	event := Event{
		Type:   EventPutObject,
		Bucket: "test",
		Key:    "file1",
		Data:   []byte("test data"),
	}

	replicator.QueueEvent(event)

	stats := replicator.GetStats()
	if stats.EventsQueued != 1 {
		t.Errorf("EventsQueued = %d, want 1", stats.EventsQueued)
	}
}

func TestReplicator_StartStop(t *testing.T) {
	config := Config{
		Enabled:       true,
		Mode:          ModeAsync,
		BatchSize:     10,
		BatchInterval: 100 * time.Millisecond,
	}

	replicator := NewReplicator(config)

	if err := replicator.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	replicator.Stop()
}

func TestReplicator_ReplicatePutObject(t *testing.T) {
	received := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/test/file1" {
			atomic.AddInt32(&received, 1)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		Enabled:       true,
		Mode:          ModeAsync,
		RemoteURL:     server.URL,
		BatchSize:     10,
		BatchInterval: 100 * time.Millisecond,
		RetryAttempts: 1,
		RetryDelay:    10 * time.Millisecond,
	}

	replicator := NewReplicator(config)
	replicator.Start()
	defer replicator.Stop()

	event := Event{
		Type:   EventPutObject,
		Bucket: "test",
		Key:    "file1",
		Data:   []byte("test data"),
	}

	replicator.QueueEvent(event)

	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("Expected 1 PUT request, got %d", received)
	}

	stats := replicator.GetStats()
	if stats.EventsReplicated != 1 {
		t.Errorf("EventsReplicated = %d, want 1", stats.EventsReplicated)
	}
}

func TestReplicator_ReplicateDeleteObject(t *testing.T) {
	received := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/test/file1" {
			atomic.AddInt32(&received, 1)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		Enabled:       true,
		RemoteURL:     server.URL,
		BatchSize:     10,
		BatchInterval: 100 * time.Millisecond,
		RetryAttempts: 1,
	}

	replicator := NewReplicator(config)
	replicator.Start()
	defer replicator.Stop()

	event := Event{
		Type:   EventDeleteObject,
		Bucket: "test",
		Key:    "file1",
	}

	replicator.QueueEvent(event)
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("Expected 1 DELETE request, got %d", received)
	}
}

func TestReplicator_DisabledNoOp(t *testing.T) {
	config := Config{
		Enabled: false,
	}

	replicator := NewReplicator(config)
	replicator.Start()

	event := Event{
		Type:   EventPutObject,
		Bucket: "test",
		Key:    "file1",
		Data:   []byte("test"),
	}

	replicator.QueueEvent(event)

	stats := replicator.GetStats()
	if stats.EventsQueued != 0 {
		t.Errorf("EventsQueued = %d, want 0 when disabled", stats.EventsQueued)
	}

	replicator.Stop()
}
