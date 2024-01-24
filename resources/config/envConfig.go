package config

import "os"

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	NATSClientID  string
	NATSClusterID string
	NATSURL       string
	Port          string
}

func NewConfig() *Config {
	return &Config{
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		NATSClientID:  os.Getenv("NATS_CLIENT_ID"),
		NATSClusterID: os.Getenv("NATS_CLUSTER_ID"),
		NATSURL:       os.Getenv("NATS_URL"),
		Port:          os.Getenv("PORT"),
	}
}
