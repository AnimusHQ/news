package registry

import (
	"fmt"
	"os"

	"github.com/AnimusHQ/news/internal/models"
	"gopkg.in/yaml.v3"
)

// File is the YAML model registry format.
type File struct {
	Models []models.ModelRecord `yaml:"models" json:"models"`
}

// LoadFile loads and validates a model registry file.
func LoadFile(path string) ([]models.ModelRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read model registry: %w", err)
	}

	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode model registry: %w", err)
	}
	if err := Validate(file.Models); err != nil {
		return nil, err
	}
	return file.Models, nil
}

// Validate checks model registry records for required fields and safe defaults.
func Validate(records []models.ModelRecord) error {
	if len(records) == 0 {
		return fmt.Errorf("model registry must contain at least one model")
	}
	seen := map[string]struct{}{}
	for _, record := range records {
		if record.ID == "" {
			return fmt.Errorf("model id is required")
		}
		if _, ok := seen[record.ID]; ok {
			return fmt.Errorf("duplicate model id: %s", record.ID)
		}
		seen[record.ID] = struct{}{}
		if record.Provider == "" {
			return fmt.Errorf("model %s provider is required", record.ID)
		}
		if record.Version == "" {
			return fmt.Errorf("model %s version is required", record.ID)
		}
		if !validStatus(record.Status) {
			return fmt.Errorf("model %s has invalid status %q", record.ID, record.Status)
		}
		if !validPrivacy(record.PrivacyTier) {
			return fmt.Errorf("model %s has invalid privacy tier %q", record.ID, record.PrivacyTier)
		}
		if len(record.Modalities) == 0 {
			return fmt.Errorf("model %s must declare at least one modality", record.ID)
		}
		if len(record.Capabilities) == 0 {
			return fmt.Errorf("model %s must declare at least one capability", record.ID)
		}
	}
	return nil
}

func validStatus(status models.ModelStatus) bool {
	switch status {
	case models.ModelStatusActive, models.ModelStatusDegraded, models.ModelStatusDisabled:
		return true
	default:
		return false
	}
}

func validPrivacy(tier models.PrivacyTier) bool {
	switch tier {
	case models.PrivacyTierPublic, models.PrivacyTierInternalApproved, models.PrivacyTierRestricted, models.PrivacyTierLocalOnly:
		return true
	default:
		return false
	}
}
