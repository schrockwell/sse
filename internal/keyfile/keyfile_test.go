package keyfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
)

func generateTestKey(t *testing.T) (keyFileContent string, secretKey string) {
	t.Helper()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}
	secretKey = identity.String()
	keyFileContent = fmt.Sprintf("# created: 2024-01-01T00:00:00Z\n# public key: %s\n%s\n", identity.Recipient().String(), secretKey)
	return
}

func TestGenerate(t *testing.T) {
	t.Run("creates new key file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		err := Generate(path, false)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("key file was not created")
		}

		// Verify file permissions
		info, _ := os.Stat(path)
		if info.Mode().Perm() != 0600 {
			t.Errorf("key file permissions = %v, want 0600", info.Mode().Perm())
		}

		// Verify contents
		data, _ := os.ReadFile(path)
		content := string(data)
		if !strings.Contains(content, "# created:") {
			t.Error("key file missing created comment")
		}
		if !strings.Contains(content, "# public key: age1") {
			t.Error("key file missing public key comment")
		}
		if !strings.Contains(content, "AGE-SECRET-KEY-") {
			t.Error("key file missing secret key")
		}
	})

	t.Run("fails if file exists without force", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		// Create the file first
		os.WriteFile(path, []byte("existing"), 0600)

		err := Generate(path, false)
		if err == nil {
			t.Fatal("Generate() should have failed for existing file")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error = %v, want 'already exists' message", err)
		}
	})

	t.Run("overwrites with force", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		// Create the file first
		os.WriteFile(path, []byte("existing"), 0600)

		err := Generate(path, true)
		if err != nil {
			t.Fatalf("Generate() with force error = %v", err)
		}

		// Verify new content
		data, _ := os.ReadFile(path)
		if string(data) == "existing" {
			t.Error("file was not overwritten")
		}
	})
}

func TestRead(t *testing.T) {
	t.Run("reads valid key file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		// Generate a key file
		Generate(path, false)

		identity, recipient, err := Read(path)
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
		if recipient == nil {
			t.Fatal("recipient is nil")
		}

		// Verify identity and recipient match
		if identity.Recipient().String() != recipient.String() {
			t.Error("recipient does not match identity.Recipient()")
		}
	})

	t.Run("reads key file without public key comment", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		// Generate a real key and write without public key comment
		_, secretKey := generateTestKey(t)
		content := fmt.Sprintf("# created: 2024-01-01T00:00:00Z\n%s\n", secretKey)
		os.WriteFile(path, []byte(content), 0600)

		identity, recipient, err := Read(path)
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
		// Recipient should be derived from identity
		if recipient == nil {
			t.Fatal("recipient should be derived from identity")
		}
	})

	t.Run("fails for non-existent file", func(t *testing.T) {
		_, _, err := Read("/nonexistent/path")
		if err == nil {
			t.Fatal("Read() should have failed")
		}
	})

	t.Run("fails for file without secret key", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.key")

		os.WriteFile(path, []byte("# just a comment\n"), 0600)

		_, _, err := Read(path)
		if err == nil {
			t.Fatal("Read() should have failed for file without secret key")
		}
		if !strings.Contains(err.Error(), "no secret key found") {
			t.Errorf("error = %v, want 'no secret key found' message", err)
		}
	})
}

func TestReadIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.key")
	Generate(path, false)

	identity, err := ReadIdentity(path)
	if err != nil {
		t.Fatalf("ReadIdentity() error = %v", err)
	}
	if identity == nil {
		t.Fatal("identity is nil")
	}
}

func TestReadRecipient(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.key")
	Generate(path, false)

	recipient, err := ReadRecipient(path)
	if err != nil {
		t.Fatalf("ReadRecipient() error = %v", err)
	}
	if recipient == nil {
		t.Fatal("recipient is nil")
	}
}

func TestParseKeyData(t *testing.T) {
	t.Run("parses full key file format", func(t *testing.T) {
		keyFileContent, _ := generateTestKey(t)
		identity, recipient, err := parseKeyData(keyFileContent)
		if err != nil {
			t.Fatalf("parseKeyData() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
		if recipient == nil {
			t.Fatal("recipient is nil")
		}
	})

	t.Run("parses just the secret key", func(t *testing.T) {
		_, secretKey := generateTestKey(t)
		identity, recipient, err := parseKeyData(secretKey)
		if err != nil {
			t.Fatalf("parseKeyData() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
		// Recipient should be derived
		if recipient == nil {
			t.Fatal("recipient should be derived from identity")
		}
	})

	t.Run("handles empty lines and whitespace", func(t *testing.T) {
		_, secretKey := generateTestKey(t)
		data := fmt.Sprintf("\n\n# comment\n\n%s\n\n", secretKey)
		identity, _, err := parseKeyData(data)
		if err != nil {
			t.Fatalf("parseKeyData() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
	})

	t.Run("fails for empty data", func(t *testing.T) {
		_, _, err := parseKeyData("")
		if err == nil {
			t.Fatal("parseKeyData() should have failed")
		}
	})

	t.Run("fails for invalid key", func(t *testing.T) {
		_, _, err := parseKeyData("AGE-SECRET-KEY-INVALID")
		if err == nil {
			t.Fatal("parseKeyData() should have failed for invalid key")
		}
	})
}

func TestLoadIdentity(t *testing.T) {
	t.Run("loads from environment variable", func(t *testing.T) {
		_, secretKey := generateTestKey(t)
		os.Setenv(MasterKeyEnvVar, secretKey)
		defer os.Unsetenv(MasterKeyEnvVar)

		identity, err := LoadIdentity()
		if err != nil {
			t.Fatalf("LoadIdentity() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
	})

	t.Run("loads full key file format from environment", func(t *testing.T) {
		keyFileContent, _ := generateTestKey(t)
		os.Setenv(MasterKeyEnvVar, keyFileContent)
		defer os.Unsetenv(MasterKeyEnvVar)

		identity, err := LoadIdentity()
		if err != nil {
			t.Fatalf("LoadIdentity() error = %v", err)
		}
		if identity == nil {
			t.Fatal("identity is nil")
		}
	})

	t.Run("fails for invalid environment variable", func(t *testing.T) {
		os.Setenv(MasterKeyEnvVar, "invalid-key")
		defer os.Unsetenv(MasterKeyEnvVar)

		_, err := LoadIdentity()
		if err == nil {
			t.Fatal("LoadIdentity() should have failed for invalid key")
		}
	})
}

func TestLoadRecipient(t *testing.T) {
	t.Run("loads from environment variable", func(t *testing.T) {
		_, secretKey := generateTestKey(t)
		os.Setenv(MasterKeyEnvVar, secretKey)
		defer os.Unsetenv(MasterKeyEnvVar)

		recipient, err := LoadRecipient()
		if err != nil {
			t.Fatalf("LoadRecipient() error = %v", err)
		}
		if recipient == nil {
			t.Fatal("recipient is nil")
		}
	})
}
