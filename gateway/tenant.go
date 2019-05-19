package gateway

import (
	"fmt"
)

type tenant struct {
	TenantID string `json:"tenant_id"`
	Audience string `json:"aud"`
	Version  uint8  `json:"version"`
}

// Valid returns an error in case this tenant had been blacklisted
func (t *tenant) Valid() error {
	if t.TenantID == "" {
		return fmt.Errorf("Tenant is empty")
	}

	return nil
}
