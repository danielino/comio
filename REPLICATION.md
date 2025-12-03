# ComIO Cross-Site Replication

Sistema di replica asincrona tra istanze ComIO su siti geografici diversi.

## Architettura

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

## Caratteristiche

### ✅ Replica Asincrona
- **Non blocca le write**: gli oggetti vengono scritti immediatamente nel sito primario
- **Queue bufferizzata**: fino a 10,000 eventi in memoria
- **Batch processing**: invia fino a 100 eventi per batch ogni secondo
- **5 worker paralleli**: per throughput elevato

### 🔄 Retry Automatico
- **3 tentativi** per evento fallito
- **Delay configurabile** tra retry (default: 5s)
- **Error logging** dettagliato

### 📊 Eventi Replicati
1. **PUT Object** - Creazione/aggiornamento oggetto
2. **DELETE Object** - Eliminazione oggetto singolo  
3. **PURGE Bucket** - Eliminazione tutti gli oggetti di un bucket

### 🎯 Ottimizzazioni
- **Oggetti piccoli (<1MB)**: data inline nel payload
- **Oggetti grandi (≥1MB)**: replica fetcha dal sito primario via HTTP

## Configurazione

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
  enabled: false  # ⚠️ IMPORTANTE: disabilita per evitare loop!
```

## Setup

### 1. Genera Token di Autenticazione

```bash
# Genera token sicuro
openssl rand -hex 32
```

### 2. Configura Site A (Primary)

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

### 3. Configura Site B (Replica)

```bash
# Site B
cat > config.yaml <<EOF
replication:
  enabled: false
EOF

./bin/comio server --config config.yaml
```

### 4. Test Replica

```bash
# Site A: upload file
curl -X PUT http://site-a:8080/mybucket/myfile \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.dat

# Attendi 1-2 secondi (batch interval)

# Site B: verifica presenza
curl http://site-b:8080/mybucket/myfile
```

## Monitoraggio

### Status Replica

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

### Metriche Chiave

- **events_queued**: Totale eventi accodati
- **events_replicated**: Eventi replicati con successo
- **events_failed**: Eventi falliti dopo tutti i retry
- **last_replication**: Timestamp ultima replica riuscita

## Performance

### Throughput Stimato

Con configurazione default:
- **Batch size**: 100 eventi
- **Batch interval**: 1 secondo
- **Workers**: 5 paralleli

**Throughput teorico**: ~500 eventi/secondo

### Latenza Replica

| Dimensione Oggetto | Latenza Stimata |
|-------------------|-----------------|
| < 1MB (inline)    | 10-50ms         |
| 1-10MB            | 100-500ms       |
| 10-100MB          | 1-5s            |

## Best Practices

### 🔒 Sicurezza

1. **Usa HTTPS** in produzione per `remote_url`
2. **Token forte**: almeno 32 caratteri random
3. **Firewall**: limita accesso solo da Site A

### ⚡ Performance

1. **Aumenta workers** per throughput alto:
   ```yaml
   # Nel codice, modifica numWorkers in replicator.go
   numWorkers := 10  # default: 5
   ```

2. **Batch size ottimale** per tuo workload:
   - File piccoli: batch_size=200
   - File grandi: batch_size=50

3. **Rete veloce**: 1Gbps+ consigliato tra siti

### 💾 Storage

1. **Stessa capacità** su entrambi i siti
2. **NVMe consigliato** per latenza bassa
3. **Monitoraggio spazio**: evita disco pieno su replica

## Troubleshooting

### Replica non funziona

```bash
# Verifica configurazione
curl http://site-a:8080/admin/replication/status

# Check logs
grep -i "replication" server.log | tail -50

# Verifica connettività
curl -I https://site-b.example.com/admin/health
```

### Eventi falliti

Gli eventi falliti dopo tutti i retry vengono loggati:

```json
{"level":"error","msg":"Failed to replicate event",
 "event_id":"1701234567890-mybucket-myfile",
 "error":"remote returned 500: out of space"}
```

**Soluzioni:**
- Aumenta `retry_attempts` e `retry_delay`
- Verifica spazio disponibile su Site B
- Check network errors tra siti

### Queue piena

Se vedi warning `"Replication queue full, dropping event"`:

```go
// Aumenta buffer queue in replicator.go
queue: make(chan Event, 50000), // default: 10000
```

## Limitazioni

1. **No replica sincrona**: eventual consistency
2. **No conflict resolution**: last-write-wins
3. **No bidirectional**: unidirezionale A→B
4. **No ordered delivery**: eventi possono arrivare out-of-order

## Future Enhancements

- [ ] Multi-site replication (A→B→C)
- [ ] Replica sincrona (mode: sync)
- [ ] Conflict detection e resolution
- [ ] WAL persistence per recovery
- [ ] Compression per oggetti grandi
- [ ] Metrics export (Prometheus)

## Esempio Completo

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
