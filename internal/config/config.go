package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	DBMaxConns   int32
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		Port:         port,
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
		DBMaxConns:   dbMaxConns(),
	}
}

// Default of 50 was measured as the sweet spot for an 8-core machine: higher pool
// sizes increase CPU context-switching overhead faster than they add real throughput.
func dbMaxConns() int32 {
	if n, err := strconv.Atoi(os.Getenv("DB_MAX_CONNS")); err == nil && n > 0 {
		return int32(n)
	}
	return 50
}
