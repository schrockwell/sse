package secrets

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"filippo.io/age"
)

func generateTestIdentity(t *testing.T) *age.X25519Identity {
	t.Helper()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}
	return identity
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"ENC[abc123]", true},
		{"ENC[]", true},
		{"ENC[some long value with spaces]", true},
		{"plaintext", false},
		{"ENC[missing suffix", false},
		{"missing prefix]", false},
		{"", false},
		{"ENC", false},
		{"[]", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := IsEncrypted(tt.value)
			if result != tt.expected {
				t.Errorf("IsEncrypted(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestEncryptDecryptValue(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()

	t.Run("encrypts and decrypts successfully", func(t *testing.T) {
		plaintext := "my secret value"

		encrypted, err := EncryptValue(plaintext, recipient)
		if err != nil {
			t.Fatalf("EncryptValue() error = %v", err)
		}

		if !IsEncrypted(encrypted) {
			t.Error("encrypted value should have ENC[] wrapper")
		}

		if encrypted == plaintext {
			t.Error("encrypted value should differ from plaintext")
		}

		decrypted, err := DecryptValue(encrypted, identity)
		if err != nil {
			t.Fatalf("DecryptValue() error = %v", err)
		}

		if decrypted != plaintext {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		encrypted, err := EncryptValue("", recipient)
		if err != nil {
			t.Fatalf("EncryptValue() error = %v", err)
		}

		decrypted, err := DecryptValue(encrypted, identity)
		if err != nil {
			t.Fatalf("DecryptValue() error = %v", err)
		}

		if decrypted != "" {
			t.Errorf("decrypted = %q, want empty", decrypted)
		}
	})

	t.Run("handles special characters", func(t *testing.T) {
		plaintext := "password with 'quotes' and \"double quotes\" and $pecial chars!"

		encrypted, err := EncryptValue(plaintext, recipient)
		if err != nil {
			t.Fatalf("EncryptValue() error = %v", err)
		}

		decrypted, err := DecryptValue(encrypted, identity)
		if err != nil {
			t.Fatalf("DecryptValue() error = %v", err)
		}

		if decrypted != plaintext {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("DecryptValue returns unencrypted values as-is", func(t *testing.T) {
		plaintext := "not encrypted"

		result, err := DecryptValue(plaintext, identity)
		if err != nil {
			t.Fatalf("DecryptValue() error = %v", err)
		}

		if result != plaintext {
			t.Errorf("result = %q, want %q", result, plaintext)
		}
	})

	t.Run("fails with wrong identity", func(t *testing.T) {
		encrypted, _ := EncryptValue("secret", recipient)

		wrongIdentity := generateTestIdentity(t)
		_, err := DecryptValue(encrypted, wrongIdentity)
		if err == nil {
			t.Error("DecryptValue() should have failed with wrong identity")
		}
	})

	t.Run("fails with invalid base64", func(t *testing.T) {
		_, err := DecryptValue("ENC[!!!invalid-base64!!!]", identity)
		if err == nil {
			t.Error("DecryptValue() should have failed with invalid base64")
		}
	})
}

func TestEncryptDecryptEnvironment(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()

	t.Run("encrypts and decrypts environment", func(t *testing.T) {
		env := map[string]string{
			"API_KEY":     "secret123",
			"DB_PASSWORD": "dbpass456",
		}

		encrypted, err := EncryptEnvironment(env, recipient)
		if err != nil {
			t.Fatalf("EncryptEnvironment() error = %v", err)
		}

		// Verify all values are encrypted
		for key, value := range encrypted {
			if !IsEncrypted(value) {
				t.Errorf("value for %s should be encrypted", key)
			}
		}

		decrypted, err := DecryptEnvironment(encrypted, identity)
		if err != nil {
			t.Fatalf("DecryptEnvironment() error = %v", err)
		}

		// Verify decrypted matches original
		for key, value := range env {
			if decrypted[key] != value {
				t.Errorf("decrypted[%s] = %q, want %q", key, decrypted[key], value)
			}
		}
	})

	t.Run("skips already encrypted values", func(t *testing.T) {
		alreadyEncrypted, _ := EncryptValue("original", recipient)
		env := map[string]string{
			"NEW_KEY":       "plaintext",
			"EXISTING_KEY":  alreadyEncrypted,
		}

		encrypted, err := EncryptEnvironment(env, recipient)
		if err != nil {
			t.Fatalf("EncryptEnvironment() error = %v", err)
		}

		// The already encrypted value should be unchanged
		if encrypted["EXISTING_KEY"] != alreadyEncrypted {
			t.Error("already encrypted value should not be re-encrypted")
		}
	})

	t.Run("handles empty environment", func(t *testing.T) {
		env := map[string]string{}

		encrypted, err := EncryptEnvironment(env, recipient)
		if err != nil {
			t.Fatalf("EncryptEnvironment() error = %v", err)
		}

		if len(encrypted) != 0 {
			t.Errorf("encrypted length = %d, want 0", len(encrypted))
		}
	})
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()

	t.Run("saves and loads file", func(t *testing.T) {
		path := filepath.Join(dir, "test.toml")

		f := &File{
			Environments: map[string]map[string]string{
				"development": {"KEY1": "value1", "KEY2": "value2"},
				"production":  {"KEY1": "prod1"},
			},
		}

		err := f.Save(path)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify environments
		if len(loaded.Environments) != 2 {
			t.Errorf("loaded %d environments, want 2", len(loaded.Environments))
		}

		if loaded.Environments["development"]["KEY1"] != "value1" {
			t.Error("development.KEY1 mismatch")
		}
	})

	t.Run("Load fails for non-existent file", func(t *testing.T) {
		_, err := Load("/nonexistent/file.toml")
		if err == nil {
			t.Error("Load() should have failed")
		}
	})

	t.Run("Load fails for invalid TOML", func(t *testing.T) {
		path := filepath.Join(dir, "invalid.toml")
		os.WriteFile(path, []byte("not valid toml [[["), 0644)

		_, err := Load(path)
		if err == nil {
			t.Error("Load() should have failed for invalid TOML")
		}
	})
}

func TestGetEnvironment(t *testing.T) {
	f := &File{
		Environments: map[string]map[string]string{
			"development": {"KEY": "dev"},
			"production":  {"KEY": "prod"},
		},
	}

	t.Run("returns existing environment", func(t *testing.T) {
		env, err := f.GetEnvironment("development")
		if err != nil {
			t.Fatalf("GetEnvironment() error = %v", err)
		}
		if env["KEY"] != "dev" {
			t.Errorf("KEY = %q, want 'dev'", env["KEY"])
		}
	})

	t.Run("fails for non-existent environment", func(t *testing.T) {
		_, err := f.GetEnvironment("staging")
		if err == nil {
			t.Error("GetEnvironment() should have failed")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error = %v, want 'not found' message", err)
		}
	})
}

func TestToEnvList(t *testing.T) {
	env := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	}

	result := ToEnvList(env)

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}

	// Sort for consistent comparison
	sort.Strings(result)

	expected := []string{"KEY1=value1", "KEY2=value2"}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestCreateDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.toml")

	err := CreateDefault(path)
	if err != nil {
		t.Fatalf("CreateDefault() error = %v", err)
	}

	// Load and verify
	f, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := f.Environments["development"]; !ok {
		t.Error("missing development environment")
	}
	if _, ok := f.Environments["production"]; !ok {
		t.Error("missing production environment")
	}
}
