package replication

// Manager handles replication
type Manager struct {
	nodes []*Node
}

// NewManager creates a new replication manager
func NewManager() *Manager {
	return &Manager{
		nodes: make([]*Node, 0),
	}
}
