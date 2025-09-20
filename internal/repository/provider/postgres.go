package provider

import (
	"context"
	"fmt"
	"workshop/internal/repository"
	"workshop/internal/repository/driver/postgres"

	"gorm.io/gorm/logger"
)

type Provider struct{}

func New() *Provider { return &Provider{} }

type PostgresConfiguration struct {
	Host           string
	Database       string
	Username       string
	Password       string
	Params         string
	MigrationsPath string
	LogLevel       logger.LogLevel
}

func (p *Provider) NewPostgresDriver(ctx context.Context, cfg *PostgresConfiguration) (repository.Driver, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Database,
	)
	if cfg.Params != "" {
		connString += "?" + cfg.Params
	}
	return postgres.NewDriver(ctx, connString, cfg.MigrationsPath, cfg.LogLevel)
}
