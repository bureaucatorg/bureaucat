package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"bereaucat/internal/store"
)

// SettingKey is the settings table key holding the assistant configuration.
const SettingKey = "assistant"

// Settings is the admin-pinned LLM configuration shared by every user.
type Settings struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key"`
	AWSRegion string `json:"aws_region"`
}

// Load reads the assistant settings; a missing row means disabled.
func Load(ctx context.Context, s store.Querier) (Settings, error) {
	var cfg Settings
	setting, err := s.GetSetting(ctx, SettingKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(setting.Value, &cfg)
	return cfg, err
}

// Validate checks the fields required by the chosen provider.
func (s Settings) Validate() error {
	if s.Model == "" {
		return errors.New("model is required")
	}
	switch s.Provider {
	case ProviderOpenAI, ProviderAnthropic, ProviderOpenRouter:
		if s.APIKey == "" {
			return errors.New("api_key is required")
		}
	case ProviderBedrock:
		if s.AWSRegion == "" {
			return errors.New("aws_region is required")
		}
	default:
		return fmt.Errorf("unknown provider %q", s.Provider)
	}
	return nil
}
