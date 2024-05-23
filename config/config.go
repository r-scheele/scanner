package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// App Configuration
type AppConfig struct {
	ClamHost       string
	ClamPort       string
	ListenPort     string
	ClamAddress    string
	BucketName     string
	SubnetEndpoint string
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfig() *AppConfig {

	if err := godotenv.Load("config/app.env"); err != nil {
		log.Println("No .env file found")
	}

	return &AppConfig{
		ClamHost:       getEnv("CLAMD_HOST", "localhost"),
		ClamPort:       getEnv("CLAMD_PORT", "3310"),
		ListenPort:     getEnv("LISTEN_PORT", "8080"),
		BucketName:     getEnv("BUCKET_NAME", "subnet-filescan-test"),
		SubnetEndpoint: getEnv("SUBNET_ENDPOINT", "http://localhost:8080"),
	}
}
