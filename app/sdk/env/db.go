package env

import "github.com/ardanlabs/practical-ai-gcuk-2026/business/sdk/sqldb"

// DBConfig builds the database configuration from the environment so every
// binary in the project connects the same way.
func DBConfig() sqldb.Config {
	return sqldb.Config{
		User:         String("DB_USER", "postgres"),
		Password:     String("DB_PASSWORD", "postgres"),
		HostPort:     String("DB_HOST_PORT", "localhost:5432"),
		Name:         String("DB_NAME", "postgres"),
		MaxIdleConns: Int("DB_MAX_IDLE_CONNS", 2),
		MaxOpenConns: Int("DB_MAX_OPEN_CONNS", 0),
		DisableTLS:   Bool("DB_DISABLE_TLS", true),
	}
}
