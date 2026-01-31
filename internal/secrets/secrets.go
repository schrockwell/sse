package secrets

import (
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strings"

	"filippo.io/age"
	"github.com/BurntSushi/toml"
	ageutil "github.com/schrockwell/sse/internal/age"
)

const (
	DefaultFile        = "env.toml"
	DefaultEnvironment = "development"
	EncryptedPrefix    = "ENC["
	EncryptedSuffix    = "]"
)

// File represents an env.toml file with multiple environments.
type File struct {
	Environments map[string]map[string]string
}

// Load reads and parses an env.toml file.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	var raw map[string]map[string]string
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	return &File{Environments: raw}, nil
}

// Save writes the secrets file to disk.
func (f *File) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create secrets file: %w", err)
	}
	defer file.Close()

	// Sort environment names for consistent output
	envNames := make([]string, 0, len(f.Environments))
	for name := range f.Environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	// Write each environment section
	for i, envName := range envNames {
		if i > 0 {
			file.WriteString("\n")
		}
		fmt.Fprintf(file, "[%s]\n", envName)

		env := f.Environments[envName]
		// Sort keys for consistent output
		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Fprintf(file, "%s = %q\n", key, env[key])
		}
	}

	return nil
}

// GetEnvironment returns the secrets for a specific environment.
func (f *File) GetEnvironment(name string) (map[string]string, error) {
	env, ok := f.Environments[name]
	if !ok {
		return nil, fmt.Errorf("environment %q not found", name)
	}
	return env, nil
}

// IsEncrypted checks if a value is encrypted.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix) && strings.HasSuffix(value, EncryptedSuffix)
}

// EncryptValue encrypts a plaintext value.
func EncryptValue(plaintext string, recipient age.Recipient) (string, error) {
	ciphertext, err := ageutil.Encrypt([]byte(plaintext), recipient)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return EncryptedPrefix + encoded + EncryptedSuffix, nil
}

// DecryptValue decrypts an encrypted value.
func DecryptValue(encrypted string, identity age.Identity) (string, error) {
	if !IsEncrypted(encrypted) {
		return encrypted, nil // Return as-is if not encrypted
	}

	encoded := strings.TrimPrefix(encrypted, EncryptedPrefix)
	encoded = strings.TrimSuffix(encoded, EncryptedSuffix)

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	plaintext, err := ageutil.Decrypt(ciphertext, identity)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// DecryptEnvironment decrypts all values in an environment.
func DecryptEnvironment(env map[string]string, identity age.Identity) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range env {
		decrypted, err := DecryptValue(value, identity)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", key, err)
		}
		result[key] = decrypted
	}
	return result, nil
}

// EncryptEnvironment encrypts all plaintext values in an environment.
func EncryptEnvironment(env map[string]string, recipient age.Recipient) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range env {
		if IsEncrypted(value) {
			result[key] = value // Already encrypted
		} else {
			encrypted, err := EncryptValue(value, recipient)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt %s: %w", key, err)
			}
			result[key] = encrypted
		}
	}
	return result, nil
}

// ToEnvList converts a map of environment variables to a slice of KEY=value strings.
func ToEnvList(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// CreateDefault creates a default env.toml with empty development and production sections.
func CreateDefault(path string) error {
	f := &File{
		Environments: map[string]map[string]string{
			"development": {},
			"production":  {},
		},
	}
	return f.Save(path)
}
