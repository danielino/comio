package replication

// NodeStatus defines the status of a node
type NodeStatus string

const (
	NodeStatusUp   NodeStatus = "UP"
	NodeStatusDown NodeStatus = "DOWN"
)

// Node represents a cluster node
type Node struct {
	ID       string
	Address  string
	Status   NodeStatus
	Capacity int64
	Used     int64
}
