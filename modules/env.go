package modules

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv loads environment variables from .env files based on the runtime environment
func LoadEnv(env string) error {
	// Define the possible .env file names
	envFiles := []string{
		".env",
		".env" + "." + env,
	}

	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err := loadFile(file); err != nil {
				return err
			}
		}
	}

	return nil
}

// loadFile reads a .env file and sets the environment variables
func loadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}
