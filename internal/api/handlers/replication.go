package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danielino/comio/internal/replication"
)

type ReplicationHandler struct {
	replicator *replication.Replicator
}

func NewReplicationHandler(replicator *replication.Replicator) *ReplicationHandler {
	return &ReplicationHandler{
		replicator: replicator,
	}
}

func (h *ReplicationHandler) GetStatus(c *gin.Context) {
	if h.replicator == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
		})
		return
	}

	stats := h.replicator.GetStats()
	
	c.JSON(http.StatusOK, gin.H{
		"enabled":            true,
		"events_queued":      stats.EventsQueued,
		"events_replicated":  stats.EventsReplicated,
		"events_failed":      stats.EventsFailed,
		"last_replication":   stats.LastReplication,
	})
}
