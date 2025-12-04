# ComIO Cross-Site Replication

Asynchronous replication system between ComIO instances on different geographic sites.

## Architecture

```
Site A (Primary)                    Site B (Replica)
┌──────────────────┐               ┌──────────────────┐
│ ComIO Server     │               │ ComIO Server     │
│ Port: 8080       │               │ Port: 8080       │
│                  │               │                  │
│ ┌──────────────┐ │   HTTPS       │ ┌──────────────┐ │
│ │ Replicator   │─┼───────────────>│ │ HTTP API     │ │
│ │ Queue: 10k   │ │   Async       │ │              │ │
│ │ Workers: 5   │ │   Batched     │ │              │ │
│ └──────────────┘ │               │ └──────────────┘ │
│                  │               │                  │
│ /dev/nvme0n1     │               │ /dev/nvme0n1     │
└──────────────────┘               └──────────────────┘
```

## Features

### ✅ Asynchronous Replication
- **Non-blocking writes**: objects are written immediately to the primary site
- **Buffered Queue**: up to 10,000 events in memory
- **Batch processing**: sends up to 100 events per batch every second
- **5 parallel workers**: for high throughput

### 🔄 Automatic Retry
- **3 attempts** for failed events
- **Configurable delay** between retries (default: 5s)
- **Detailed error logging**

### 📊 Replicated Events
1. **PUT Object** - Object creation/update
2. **DELETE Object** - Single object deletion
3. **PURGE Bucket** - Deletion of all objects in a bucket

### 🎯 Optimizations
- **Small objects (<1MB)**: data inline in the payload
- **Large objects (≥1MB)**: replica fetches from the primary site via HTTP

## Configuration

### Site A (Primary) - config.yaml

```yaml
server:
  port: 8080
  storage_path: /dev/nvme0n1
  storage_size: 1TB

replication:
  enabled: true
  mode: async
  remote_url: https://site-b.example.com
  remote_token: "secret-token-xyz123"
  batch_size: 100
  batch_interval: 1s
  retry_attempts: 3
  retry_delay: 5s
```

### Site B (Replica) - config.yaml

```yaml
server:
  port: 8080
  storage_path: /dev/nvme0n1
  storage_size: 1TB

replication:
  enabled: false  # ⚠️ IMPORTANT: disable to avoid loops!
```

## Setup

### 1. Generate Authentication Token

```bash
# Generate secure token
openssl rand -hex 32
```

### 2. Configure Site A (Primary)

```bash
# Site A
cat > config.yaml <<EOF
replication:
  enabled: true
  remote_url: https://site-b.example.com
  remote_token: "abc123..."
EOF

./bin/comio server --config config.yaml
```

### 3. Configure Site B (Replica)

```bash
# Site B
cat > config.yaml <<EOF
replication:
  enabled: false
EOF

./bin/comio server --config config.yaml
```

### 4. Test Replication

```bash
# Site A: upload file
curl -X PUT http://site-a:8080/mybucket/myfile \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.dat

# Wait 1-2 seconds (batch interval)

# Site B: verify presence
curl http://site-b:8080/mybucket/myfile
```

## Monitoring

### Replication Status

```bash
curl http://site-a:8080/admin/replication/status
```

**Response:**
```json
{
  "enabled": true,
  "events_queued": 1523,
  "events_replicated": 1500,
  "events_failed": 23,
  "last_replication": "2025-12-03T21:00:00Z"
}
```

### Key Metrics

- **events_queued**: Total events queued
- **events_replicated**: Events successfully replicated
- **events_failed**: Events failed after all retries
- **last_replication**: Timestamp of last successful replication

## Performance

### Estimated Throughput

With default configuration:
- **Batch size**: 100 events
- **Batch interval**: 1 second
- **Workers**: 5 parallel

**Theoretical throughput**: ~500 events/second

### Replication Latency

| Object Size       | Estimated Latency |
|-------------------|-------------------|
| < 1MB (inline)    | 10-50ms           |
| 1-10MB            | 100-500ms         |
| 10-100MB          | 1-5s              |

## Best Practices

### 🔒 Security

1. **Use HTTPS** in production for `remote_url`
2. **Strong Token**: at least 32 random characters
3. **Firewall**: limit access only from Site A

### ⚡ Performance

1. **Increase workers** for high throughput:
   ```yaml
   # In code, modify numWorkers in replicator.go
   numWorkers := 10  # default: 5
   ```

2. **Optimal batch size** for your workload:
   - Small files: batch_size=200
   - Large files: batch_size=50

3. **Fast Network**: 1Gbps+ recommended between sites

### 💾 Storage

1. **Same capacity** on both sites
2. **NVMe recommended** for low latency
3. **Space monitoring**: avoid full disk on replica

## Troubleshooting

### Replication not working

```bash
# Verify configuration
curl http://site-a:8080/admin/replication/status

# Check logs
grep -i "replication" server.log | tail -50

# Verify connectivity
curl -I https://site-b.example.com/admin/health
```

### Failed Events

Events failed after all retries are logged:

```json
{"level":"error","msg":"Failed to replicate event",
 "event_id":"1701234567890-mybucket-myfile",
 "error":"remote returned 500: out of space"}
```

**Solutions:**
- Increase `retry_attempts` and `retry_delay`
- Check available space on Site B
- Check network errors between sites

### Queue Full

If you see warning `"Replication queue full, dropping event"`:

```go
// Increase queue buffer in replicator.go
queue: make(chan Event, 50000), // default: 10000
```

## Limitations

1. **No synchronous replication**: eventual consistency
2. **No conflict resolution**: last-write-wins
3. **No bidirectional**: unidirectional A→B
4. **No ordered delivery**: events may arrive out-of-order

## Future Enhancements

- [ ] Multi-site replication (A→B→C)
- [ ] Synchronous replication (mode: sync)
- [ ] Conflict detection and resolution
- [ ] WAL persistence for recovery
- [ ] Compression for large objects
- [ ] Metrics export (Prometheus)

## Complete Example

```bash
# Site A: Primary
./bin/comio server \
  --storage /dev/nvme0n1 \
  --size 1TB \
  --replication-enabled \
  --replication-remote https://site-b.example.com \
  --replication-token "abc123..."

# Site B: Replica
./bin/comio server \
  --storage /dev/nvme0n1 \
  --size 1TB

# Test workload
for i in {1..1000}; do
  dd if=/dev/urandom bs=1M count=1 | \
    curl -X PUT http://site-a:8080/test/file$i \
      -H "Content-Type: application/octet-stream" \
      --data-binary @-
done

# Check replica status
watch -n1 'curl -s http://site-a:8080/admin/replication/status | jq'
```
