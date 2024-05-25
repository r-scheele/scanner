package config

import (
	"os"
)

// App Configuration
type AppConfig struct {
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

	return &AppConfig{
		ClamAddress:    getEnv("CLAM_ADDRESS", "tcp://localhost:3310"),
		ListenPort:     getEnv("LISTEN_PORT", "8080"),
		BucketName:     getEnv("BUCKET_NAME", "subnet-filescan-test"),
		SubnetEndpoint: getEnv("SUBNET_ENDPOINT", "http://localhost:8080"),
	}
}
