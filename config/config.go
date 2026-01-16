package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	GitHubClientID     string
	GitHubClientSecret string
	GithubRedirectURL  string
	JWTSecret          string
	EncryptionKey      string
	Production         bool
	// Add more config vars as needed
}

var Cfg *Config

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		// panic("Error loading .env file")
		fmt.Println("[WARN] Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	production := os.Getenv("PRODUCTION")

	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-prod"
	}
	Cfg = &Config{
		Port:               port,
		GitHubClientID:     githubClientID,
		GitHubClientSecret: githubClientSecret,
		GithubRedirectURL:  githubRedirectURL,
		JWTSecret:          jwtSecret,
		EncryptionKey:      encryptionKey,
		Production:         production == "true",
	}
	return Cfg
}
