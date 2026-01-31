package age

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestEncryptDecrypt(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()

	t.Run("encrypts and decrypts successfully", func(t *testing.T) {
		plaintext := []byte("hello, world!")

		ciphertext, err := Encrypt(plaintext, recipient)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		if bytes.Equal(ciphertext, plaintext) {
			t.Error("ciphertext should not equal plaintext")
		}

		decrypted, err := Decrypt(ciphertext, identity)
		if err != nil {
			t.Fatalf("Decrypt() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("handles empty plaintext", func(t *testing.T) {
		plaintext := []byte("")

		ciphertext, err := Encrypt(plaintext, recipient)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		decrypted, err := Decrypt(ciphertext, identity)
		if err != nil {
			t.Fatalf("Decrypt() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %q, want empty", decrypted)
		}
	})

	t.Run("handles binary data", func(t *testing.T) {
		plaintext := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}

		ciphertext, err := Encrypt(plaintext, recipient)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		decrypted, err := Decrypt(ciphertext, identity)
		if err != nil {
			t.Fatalf("Decrypt() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %v, want %v", decrypted, plaintext)
		}
	})

	t.Run("handles unicode", func(t *testing.T) {
		plaintext := []byte("こんにちは世界 🌍")

		ciphertext, err := Encrypt(plaintext, recipient)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		decrypted, err := Decrypt(ciphertext, identity)
		if err != nil {
			t.Fatalf("Decrypt() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("fails with wrong identity", func(t *testing.T) {
		plaintext := []byte("secret message")

		ciphertext, err := Encrypt(plaintext, recipient)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}

		// Use a different identity to decrypt
		wrongIdentity := generateTestIdentity(t)
		_, err = Decrypt(ciphertext, wrongIdentity)
		if err == nil {
			t.Error("Decrypt() should have failed with wrong identity")
		}
	})

	t.Run("fails with invalid ciphertext", func(t *testing.T) {
		_, err := Decrypt([]byte("not valid ciphertext"), identity)
		if err == nil {
			t.Error("Decrypt() should have failed with invalid ciphertext")
		}
	})
}

func TestEncryptFile(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()
	dir := t.TempDir()

	t.Run("encrypts file successfully", func(t *testing.T) {
		inputPath := filepath.Join(dir, "input.txt")
		outputPath := filepath.Join(dir, "output.age")
		plaintext := []byte("file contents")

		os.WriteFile(inputPath, plaintext, 0644)

		err := EncryptFile(inputPath, outputPath, recipient)
		if err != nil {
			t.Fatalf("EncryptFile() error = %v", err)
		}

		// Verify output exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Fatal("output file was not created")
		}

		// Verify output is different from input
		ciphertext, _ := os.ReadFile(outputPath)
		if bytes.Equal(ciphertext, plaintext) {
			t.Error("output should be encrypted")
		}
	})

	t.Run("fails for non-existent input", func(t *testing.T) {
		err := EncryptFile("/nonexistent", filepath.Join(dir, "out.age"), recipient)
		if err == nil {
			t.Error("EncryptFile() should have failed")
		}
	})
}

func TestDecryptFile(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()
	dir := t.TempDir()

	t.Run("decrypts file successfully", func(t *testing.T) {
		plaintext := []byte("secret file contents")

		// Encrypt first
		ciphertext, _ := Encrypt(plaintext, recipient)
		encryptedPath := filepath.Join(dir, "encrypted.age")
		outputPath := filepath.Join(dir, "decrypted.txt")
		os.WriteFile(encryptedPath, ciphertext, 0644)

		err := DecryptFile(encryptedPath, outputPath, identity)
		if err != nil {
			t.Fatalf("DecryptFile() error = %v", err)
		}

		// Verify output
		decrypted, _ := os.ReadFile(outputPath)
		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("fails for non-existent input", func(t *testing.T) {
		err := DecryptFile("/nonexistent", filepath.Join(dir, "out.txt"), identity)
		if err == nil {
			t.Error("DecryptFile() should have failed")
		}
	})
}

func TestDecryptToMemory(t *testing.T) {
	identity := generateTestIdentity(t)
	recipient := identity.Recipient()
	dir := t.TempDir()

	t.Run("decrypts to memory successfully", func(t *testing.T) {
		plaintext := []byte("in-memory secret")

		ciphertext, _ := Encrypt(plaintext, recipient)
		path := filepath.Join(dir, "encrypted.age")
		os.WriteFile(path, ciphertext, 0644)

		decrypted, err := DecryptToMemory(path, identity)
		if err != nil {
			t.Fatalf("DecryptToMemory() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("fails for non-existent file", func(t *testing.T) {
		_, err := DecryptToMemory("/nonexistent", identity)
		if err == nil {
			t.Error("DecryptToMemory() should have failed")
		}
	})
}
