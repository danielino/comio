package lifecycle

// Rule represents a lifecycle rule
type Rule struct {
	ID                   string
	Status               string // Enabled, Disabled
	Filter               Filter
	Transitions          []Transition
	Expiration           *Expiration
	NoncurrentVersions   *NoncurrentVersionExpiration
}

type Filter struct {
	Prefix string
}

type Transition struct {
	Days         int
	StorageClass string
}

type Expiration struct {
	Days int
}

type NoncurrentVersionExpiration struct {
	NoncurrentDays int
}
