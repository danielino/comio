package lifecycle

import (
	"testing"
)

func TestRule_Creation(t *testing.T) {
	rule := Rule{
		ID:     "rule1",
		Status: "Enabled",
		Filter: Filter{Prefix: "logs/"},
		Transitions: []Transition{
			{Days: 30, StorageClass: "GLACIER"},
		},
		Expiration: &Expiration{Days: 90},
	}

	if rule.ID != "rule1" {
		t.Errorf("ID = %s, want rule1", rule.ID)
	}
	if rule.Status != "Enabled" {
		t.Errorf("Status = %s, want Enabled", rule.Status)
	}
	if rule.Filter.Prefix != "logs/" {
		t.Errorf("Filter.Prefix = %s, want logs/", rule.Filter.Prefix)
	}
	if len(rule.Transitions) != 1 {
		t.Errorf("Transitions count = %d, want 1", len(rule.Transitions))
	}
	if rule.Expiration.Days != 90 {
		t.Errorf("Expiration.Days = %d, want 90", rule.Expiration.Days)
	}
}

func TestRule_Disabled(t *testing.T) {
	rule := Rule{
		ID:     "rule2",
		Status: "Disabled",
	}

	if rule.Status != "Disabled" {
		t.Errorf("Status = %s, want Disabled", rule.Status)
	}
}

func TestFilter(t *testing.T) {
	filter := Filter{Prefix: "archive/"}
	if filter.Prefix != "archive/" {
		t.Errorf("Prefix = %s, want archive/", filter.Prefix)
	}
}

func TestTransition(t *testing.T) {
	transition := Transition{
		Days:         60,
		StorageClass: "DEEP_ARCHIVE",
	}

	if transition.Days != 60 {
		t.Errorf("Days = %d, want 60", transition.Days)
	}
	if transition.StorageClass != "DEEP_ARCHIVE" {
		t.Errorf("StorageClass = %s, want DEEP_ARCHIVE", transition.StorageClass)
	}
}

func TestExpiration(t *testing.T) {
	exp := Expiration{Days: 365}
	if exp.Days != 365 {
		t.Errorf("Days = %d, want 365", exp.Days)
	}
}

func TestNoncurrentVersionExpiration(t *testing.T) {
	nve := NoncurrentVersionExpiration{NoncurrentDays: 30}
	if nve.NoncurrentDays != 30 {
		t.Errorf("NoncurrentDays = %d, want 30", nve.NoncurrentDays)
	}
}
