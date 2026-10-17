package config

import (
	"fmt"
	"os"
)

// all the necessary configurations

type Config struct {
	Port       string
	DBHOST     string
	DBPORT     string
	DBPASSWORD string
	DBNAME     string
	DBUSER     string
	DBSSLMODE  string
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return fallback
	}
	return v
}
func Load() *Config {
	// get the the env and return them
	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHOST:     getEnv("DBHOST", "localhost"),
		DBPASSWORD: getEnv("DBPASSWORD", ""),
		DBNAME:     getEnv("DBNAME", "ledgerdb"),
		DBUSER:     getEnv("DBUSER", "ledger_user"),
		DBSSLMODE:  getEnv("DBSSLMODE", "false"),
		DBPORT:     getEnv("DBPORT", "5432"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s password=%s dbname=%s dbuser=%s sslmode=%s", c.DBHOST, c.DBPORT, c.DBPASSWORD, c.DBNAME, c.DBUSER, c.DBSSLMODE)
}
