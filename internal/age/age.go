package age

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// Encrypt encrypts plaintext using the given recipient and returns armored ciphertext.
func Encrypt(plaintext []byte, recipient age.Recipient) ([]byte, error) {
	var buf bytes.Buffer

	armorWriter := armor.NewWriter(&buf)

	w, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption writer: %w", err)
	}

	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("failed to write plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encryption writer: %w", err)
	}

	if err := armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close armor writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Decrypt decrypts armored ciphertext using the given identity.
func Decrypt(ciphertext []byte, identity age.Identity) ([]byte, error) {
	armorReader := armor.NewReader(bytes.NewReader(ciphertext))

	r, err := age.Decrypt(armorReader, identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create decryption reader: %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read plaintext: %w", err)
	}

	return plaintext, nil
}

// EncryptFile reads a plaintext file, encrypts it, and writes to the output path.
func EncryptFile(inputPath, outputPath string, recipient age.Recipient) error {
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	ciphertext, err := Encrypt(plaintext, recipient)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, ciphertext, 0644); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// DecryptFile reads an encrypted file and writes the plaintext to the output path.
func DecryptFile(inputPath, outputPath string, identity age.Identity) error {
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	plaintext, err := Decrypt(ciphertext, identity)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// DecryptToMemory reads an encrypted file and returns the plaintext.
func DecryptToMemory(path string, identity age.Identity) ([]byte, error) {
	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	return Decrypt(ciphertext, identity)
}
