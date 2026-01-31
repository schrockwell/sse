package keyfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"filippo.io/age"
)

const DefaultKeyFile = "master.key"
const MasterKeyEnvVar = "SSE_MASTER_KEY"

// Generate creates a new age X25519 identity and writes it to the specified file.
func Generate(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("key file %s already exists (use --force to overwrite)", path)
		}
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# public key: %s\n", identity.Recipient().String())
	fmt.Fprintf(f, "%s\n", identity.String())

	return nil
}

// Read reads the key file and returns the age identity and recipient.
func Read(path string) (*age.X25519Identity, *age.X25519Recipient, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open key file: %w", err)
	}
	defer f.Close()

	var identity *age.X25519Identity
	var recipient *age.X25519Recipient

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// Extract public key from comment if present
			if strings.HasPrefix(line, "# public key: ") {
				pubKey := strings.TrimPrefix(line, "# public key: ")
				r, err := age.ParseX25519Recipient(pubKey)
				if err == nil {
					recipient = r
				}
			}
			continue
		}

		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			id, err := age.ParseX25519Identity(line)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse secret key: %w", err)
			}
			identity = id
			// Derive recipient from identity if not already set
			if recipient == nil {
				recipient = identity.Recipient()
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to read key file: %w", err)
	}

	if identity == nil {
		return nil, nil, fmt.Errorf("no secret key found in key file")
	}

	return identity, recipient, nil
}

// ReadRecipient reads only the public key from the key file.
func ReadRecipient(path string) (*age.X25519Recipient, error) {
	_, recipient, err := Read(path)
	return recipient, err
}

// ReadIdentity reads only the secret key from the key file.
func ReadIdentity(path string) (*age.X25519Identity, error) {
	identity, _, err := Read(path)
	return identity, err
}

// LoadIdentity loads the identity from SSE_MASTER_KEY env var, or falls back to the default key file.
func LoadIdentity() (*age.X25519Identity, error) {
	if key := os.Getenv(MasterKeyEnvVar); key != "" {
		identity, err := age.ParseX25519Identity(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", MasterKeyEnvVar, err)
		}
		return identity, nil
	}
	return ReadIdentity(DefaultKeyFile)
}

// LoadRecipient loads the recipient (public key) from SSE_MASTER_KEY env var, or falls back to the default key file.
func LoadRecipient() (*age.X25519Recipient, error) {
	if key := os.Getenv(MasterKeyEnvVar); key != "" {
		identity, err := age.ParseX25519Identity(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", MasterKeyEnvVar, err)
		}
		return identity.Recipient(), nil
	}
	return ReadRecipient(DefaultKeyFile)
}
