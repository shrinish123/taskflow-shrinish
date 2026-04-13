package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	origins := os.Getenv("CORS_ORIGINS")
	var allowedOrigins []string
	if origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	} else {
		allowedOrigins = []string{"*"}
	}

	return &Config{
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		Port:           port,
		AllowedOrigins: allowedOrigins,
	}, nil
}
