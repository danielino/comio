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

func TestReplicator_PurgeBucket(t *testing.T) {
	received := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/admin/test/objects" {
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
		Type:   EventPurgeBucket,
		Bucket: "test",
	}

	replicator.QueueEvent(event)
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("Expected 1 PURGE request, got %d", received)
	}
}

func TestReplicator_LargeObjectWithURL(t *testing.T) {
	received := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/test/large-file" {
			atomic.AddInt32(&received, 1)
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "GET" && r.URL.Path == "/test/large-file" {
			w.Write([]byte("large data"))
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

	// Event with DataURL instead of inline Data
	event := Event{
		Type:    EventPutObject,
		Bucket:  "test",
		Key:     "large-file",
		DataURL: server.URL + "/test/large-file",
	}

	replicator.QueueEvent(event)
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("Expected 1 PUT request for large object, got %d", received)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Enabled {
		t.Error("DefaultConfig() Enabled = true, want false (default is disabled)")
	}

	if config.Mode != ModeAsync {
		t.Errorf("DefaultConfig() Mode = %s, want %s", config.Mode, ModeAsync)
	}

	if config.BatchSize != 100 {
		t.Errorf("DefaultConfig() BatchSize = %d, want 100", config.BatchSize)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("DefaultConfig() RetryAttempts = %d, want 3", config.RetryAttempts)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Error("NewManager() returned nil")
	}
}

func TestReplicator_FailedRetries(t *testing.T) {
	attempts := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := Config{
		Enabled:       true,
		RemoteURL:     server.URL,
		BatchSize:     10,
		BatchInterval: 50 * time.Millisecond,
		RetryAttempts: 3,
		RetryDelay:    10 * time.Millisecond,
	}

	replicator := NewReplicator(config)
	replicator.Start()
	defer replicator.Stop()

	event := Event{
		Type:   EventPutObject,
		Bucket: "test",
		Key:    "fail",
		Data:   []byte("data"),
	}

	replicator.QueueEvent(event)
	time.Sleep(500 * time.Millisecond)

	// RetryAttempts means total attempts (initial + retries)
	totalAttempts := atomic.LoadInt32(&attempts)
	if totalAttempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", totalAttempts)
	}

	stats := replicator.GetStats()
	if stats.EventsFailed != 1 {
		t.Errorf("EventsFailed = %d, want 1", stats.EventsFailed)
	}
}

func TestReplicator_QueueFull(t *testing.T) {
	config := Config{
		Enabled:       false, // Disabled so events just queue
		BatchSize:     10,
		BatchInterval: 1 * time.Second,
	}

	replicator := NewReplicator(config)

	// Queue multiple events when disabled
	for i := 0; i < 5; i++ {
		event := Event{
			Type:   EventPutObject,
			Bucket: "test",
			Key:    "file",
			Data:   []byte("data"),
		}
		replicator.QueueEvent(event)
	}

	// When disabled, queue should remain empty
	stats := replicator.GetStats()
	if stats.EventsQueued != 0 {
		t.Errorf("EventsQueued = %d, want 0 when disabled", stats.EventsQueued)
	}
}
