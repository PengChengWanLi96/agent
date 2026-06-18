package config

import (
	"os"
)

type Config struct {
	ServerAddr string
	Docker     DockerConfig
}

type DockerConfig struct {
	Host       string
	APIVersion string
	TLSVerify  bool
	CertPath   string
	KeyPath    string
	CAPath     string
}

func Load() *Config {
	return &Config{
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		Docker: DockerConfig{
			Host:       getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
			APIVersion: getEnv("DOCKER_API_VERSION", "v1.43"),
			TLSVerify:  getEnv("DOCKER_TLS_VERIFY", "") != "",
			CertPath:   getEnv("DOCKER_CERT_PATH", ""),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
