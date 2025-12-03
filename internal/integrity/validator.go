package integrity

import (
	"fmt"
	"io"
)

// Validator validates data integrity
type Validator struct {
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate verifies the checksum of the data
func (v *Validator) Validate(r io.Reader, expectedChecksum Checksum) error {
	calculated, err := CalculateChecksum(r, expectedChecksum.Algorithm)
	if err != nil {
		return err
	}

	if calculated != expectedChecksum.Value {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum.Value, calculated)
	}

	return nil
}
