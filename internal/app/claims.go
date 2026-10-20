package app

import (
	"fmt"
	"strings"
)

type claim struct {
	Name string `yaml:"name"`
}

// @impl CON-004 CLM-001-AC4 CLM-001-AC6 CLM-002-AC3
// Validate a presentation copy without normalizing or rewriting saved YAML.
func validateClaims(claims []claim) error {
	for i := range claims {
		claims[i].Name = strings.TrimSpace(claims[i].Name)
		if !validName(claims[i].Name) {
			return fmt.Errorf("invalid claimant name")
		}
		for _, prior := range claims[:i] {
			if strings.EqualFold(prior.Name, claims[i].Name) {
				return fmt.Errorf("duplicate claimant name")
			}
		}
	}
	return nil
}
