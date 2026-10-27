package config

import "testing"

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
}

func TestLoad_PortFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
}

func TestLoad_DatabaseURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	cfg := Load()
	if cfg.DatabaseURL != "postgres://u:p@localhost:5432/db" {
		t.Errorf("DatabaseURL = %q, want the env value", cfg.DatabaseURL)
	}
}

func TestLoad_KafkaBrokersFromEnv(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	cfg := Load()
	if cfg.KafkaBrokers != "localhost:9092" {
		t.Errorf("KafkaBrokers = %q, want the env value", cfg.KafkaBrokers)
	}
}

func TestLoad_DBMaxConnsDefault(t *testing.T) {
	t.Setenv("DB_MAX_CONNS", "")
	cfg := Load()
	if cfg.DBMaxConns != 50 {
		t.Errorf("DBMaxConns = %d, want 50", cfg.DBMaxConns)
	}
}

func TestLoad_DBMaxConnsFromEnv(t *testing.T) {
	t.Setenv("DB_MAX_CONNS", "20")
	cfg := Load()
	if cfg.DBMaxConns != 20 {
		t.Errorf("DBMaxConns = %d, want 20", cfg.DBMaxConns)
	}
}
