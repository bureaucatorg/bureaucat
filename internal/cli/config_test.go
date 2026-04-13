package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndClearConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("BUREAUCAT_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))

	want := Config{URL: "http://localhost:1341", Token: "bcat_test"}
	path, err := SaveConfig(want)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config perms = %o, want 600", info.Mode().Perm())
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if got != want {
		t.Fatalf("LoadConfig() = %#v, want %#v", got, want)
	}

	if err := ClearConfig(); err != nil {
		t.Fatalf("ClearConfig() error = %v", err)
	}

	got, err = LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after clear error = %v", err)
	}
	if got != (Config{}) {
		t.Fatalf("LoadConfig() after clear = %#v, want empty", got)
	}
}

func TestGetCredentialsPrecedence(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("BUREAUCAT_CONFIG_PATH", cfgPath)

	if _, err := SaveConfig(Config{URL: "http://config", Token: "config-token"}); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	url, token, err := GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() error = %v", err)
	}
	if url != "http://config" || token != "config-token" {
		t.Fatalf("GetCredentials() = (%q, %q), want config values", url, token)
	}

	t.Setenv("BUREAUCAT_URL", "http://env")
	url, token, err = GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() with env URL error = %v", err)
	}
	if url != "http://env" || token != "config-token" {
		t.Fatalf("GetCredentials() partial env override = (%q, %q)", url, token)
	}

	t.Setenv("BUREAUCAT_TOKEN", "env-token")
	url, token, err = GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() with full env error = %v", err)
	}
	if url != "http://env" || token != "env-token" {
		t.Fatalf("GetCredentials() env override = (%q, %q), want env values", url, token)
	}
}
