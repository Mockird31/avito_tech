//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Mockird31/avito_tech/config"
	"github.com/jackc/pgx/v5"
)

func postgresConfigFromDSN(dsn string) config.PostgresConfig {
	clean := strings.TrimPrefix(dsn, "postgres://")
	parts := strings.Split(clean, "@")
	up := strings.Split(parts[0], ":")
	hostPort := strings.Split(parts[1], "/")[0]
	hp := strings.Split(hostPort, ":")
	dbNamePart := strings.Split(parts[1], "/")[1]
	dbName := strings.Split(dbNamePart, "?")[0]
	return config.PostgresConfig{
		PostgresUser:     up[0],
		PostgresPassword: up[1],
		PostgresHost:     hp[0],
		PostgresPort:     hp[1],
		PostgresDB:       dbName,
	}
}

func WaitForPostgres(ctx context.Context, dsn string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		conn, err := pgx.Connect(dialCtx, dsn)
		if err == nil {
			var one int
			if err = conn.QueryRow(dialCtx, "SELECT 1").Scan(&one); err == nil && one == 1 {
				_ = conn.Close(dialCtx)
				cancel()
				return nil
			}
			_ = conn.Close(dialCtx)
		}
		cancel()
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("postgres not ready: %w", lastErr)
}
