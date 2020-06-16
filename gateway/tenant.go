package gateway

import (
	"fmt"
)

type tenant struct {
	TenantID string `json:"tenant_id"`
	Audience string `json:"aud"`
	Version  uint8  `json:"version"`
}

// Valid returns an error if JWT payload is incomplete
func (t *tenant) Valid() error {
	if t.TenantID == "" {
		return fmt.Errorf("tenant is empty")
	}

	return nil
}
