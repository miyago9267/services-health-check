package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ReadEnv loads env variables from a file based on ENV/env.
// It returns os.ErrNotExist when the file is missing.
func ReadEnv() error {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("env")))
	}
	filename := "./.env"
	switch env {
	case "prd", "prod", "production":
		filename = "./.env.production"
	case "bak", "backup":
		filename = "./.env.bak"
	case "dev", "development":
		filename = "./.env.development"
	case "local":
		filename = "./.env.local"
	}
	if _, err := os.Stat(filename); err != nil {
		return err
	}

	envMap, err := godotenv.Read(filename)
	if err != nil {
		return err
	}
	for k, v := range envMap {
		_ = os.Setenv(k, v)
	}
	return nil
}
