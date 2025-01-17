package provider

import (
	"fmt"
	"workshop/internal/repository"
	"workshop/internal/repository/driver/postgres"
)

type Provider struct {
}

func NewRepositoryProvider() *Provider {
	return &Provider{}
}

type PostgresConfiguration struct {
	Host           string
	Database       string
	Username       string
	Password       string
	Params         string
	MigrationsPath string
}

func (r *Provider) GetPostgresRepository(cfg *PostgresConfiguration) (repository.Repository, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Database,
	)

	if cfg.Params != "" {
		connString += "?" + cfg.Params
	}

	return postgres.NewGORMDriver(connString, cfg.MigrationsPath)
}
