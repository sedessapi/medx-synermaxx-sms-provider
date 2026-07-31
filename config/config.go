package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const DefaultSynermaxxBaseURL = "https://svr36.synermaxx.asia"

type Config struct {
	ServerPort string
	GinMode    string

	SynermaxxBaseURL    string
	SynermaxxAppKey     string
	SynermaxxOriginator string
}

func Load() (*Config, error) {
	_ = godotenv.Overload(".env")
	cfg := &Config{
		ServerPort:          getEnv("SERVER_PORT", "8064"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		SynermaxxBaseURL:    getEnv("SYNERMAXX_SMS_BASE_URL", DefaultSynermaxxBaseURL),
		SynermaxxAppKey:     getEnv("SYNERMAXX_SMS_APPKEY", ""),
		SynermaxxOriginator: getEnv("SYNERMAXX_SMS_ORIGINATOR", "22566100"),
	}
	if cfg.SynermaxxAppKey == "" {
		return nil, fmt.Errorf("SYNERMAXX_SMS_APPKEY is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
