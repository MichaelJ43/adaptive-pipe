package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL  string
	RedisURL     string
	ValidateURL  string
	FileURL      string
	JWTSecret    string
	ListenAddr   string
	SeedAdminPW  string // optional override for demo seed user
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ap:ap@localhost:5432/adaptivepipe?sslmode=disable"
	}
	rURL := os.Getenv("REDIS_URL")
	if rURL == "" {
		rURL = "redis://localhost:6379"
	}
	v := os.Getenv("VALIDATE_URL")
	if v == "" {
		v = "http://localhost:8081"
	}
	f := os.Getenv("FILE_URL")
	if f == "" {
		f = "http://localhost:8082"
	}
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		jwt = "dev-insecure-jwt-secret-change-in-production"
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if jwt == "dev-insecure-jwt-secret-change-in-production" {
		fmt.Fprintf(os.Stderr, "warning: using default JWT_SECRET (set JWT_SECRET in production)\n")
	}
	return &Config{
		DatabaseURL: dbURL,
		RedisURL:    rURL,
		ValidateURL: v,
		FileURL:     f,
		JWTSecret:   jwt,
		ListenAddr:  addr,
		SeedAdminPW: os.Getenv("SEED_DEMO_ADMIN_PASSWORD"),
	}, nil
}
