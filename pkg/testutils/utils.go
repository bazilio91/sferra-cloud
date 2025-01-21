// pkg/testutils/utils.go
package testutils

import (
	"log"
	"os"
	"testing"
)

func LoadEnv() error {
	// Assuming that the .env file is not needed in tests, or you can add logic to load it
	return nil
}

func SetupTest(m *testing.M) {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}
