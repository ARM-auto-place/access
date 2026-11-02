package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/ARM-auto-place/access/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type ConnectManager interface {
	GetMasterConn() *pgxpool.Pool
	GetReplicaConn() *pgxpool.Pool
	Close()
}
type manager struct {
	log     zerolog.Logger
	master  *pgxpool.Pool
	replica *pgxpool.Pool
}

func (m *manager) GetMasterConn() *pgxpool.Pool {
	return m.master

}
func (m *manager) GetReplicaConn() *pgxpool.Pool {
	return m.replica
}
func (m *manager) Close() {
	if m.master != m.replica {
		m.replica.Close()
	}
	m.master.Close()
}
func New(log zerolog.Logger, cfg config.PostgresConfig) (ConnectManager, error) {
	master, err := dbConnect(cfg, cfg.Addr, cfg.User, cfg.Password)
	if err != nil {
		return nil, err
	}
	replica := master
	if cfg.ReplicaAddr != cfg.Addr || cfg.UserRO != cfg.User || cfg.PasswordRO != cfg.Password {
		replica, err = dbConnect(cfg, cfg.ReplicaAddr, cfg.UserRO, cfg.PasswordRO)
		if err != nil {
			master.Close()
			return nil, err
		}
	}
	return &manager{log: log, master: master, replica: replica}, nil
}
func dbConnect(cfg config.PostgresConfig, addr, user, password string) (*pgxpool.Pool, error) {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL address: %w", err)
	}
	dsn := &url.URL{Scheme: "postgres", Host: addr, Path: "/" + cfg.DB, User: url.UserPassword(user, password)}
	q := dsn.Query()
	q.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = q.Encode()
	pcfg, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL connection configuration")
	}
	pcfg.MaxConns = int32(cfg.MaxConn)
	pcfg.MaxConnIdleTime = cfg.MaxIdleLifetime
	pcfg.MaxConnLifetime = cfg.MaxLifetime
	pcfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	pcfg.ConnConfig.StatementCacheCapacity = cfg.PrepareCacheCap
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return pool, nil
}
