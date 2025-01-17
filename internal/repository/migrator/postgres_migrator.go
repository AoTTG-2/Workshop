package migrator

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type PostgresMigrator struct {
	m *migrate.Migrate
}

func NewPostgresMigrator(migrationsPath, databaseURL string) (*PostgresMigrator, error) {
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return nil, err
	}
	m.Log = &Logger{}

	return &PostgresMigrator{m: m}, nil
}

func (mig *PostgresMigrator) Up() error {
	if err := mig.m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Info().Msg("Migrations applied successfully!")
	return nil
}

func (mig *PostgresMigrator) Down() error {
	if err := mig.m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Info().Msg("Migrations rolled back successfully!")
	return nil
}

func (mig *PostgresMigrator) Drop() error {
	if err := mig.m.Drop(); err != nil {
		return err
	}
	log.Info().Msg("Migrations dropped successfully!")
	return nil
}
func (mig *PostgresMigrator) Steps(steps int) error {
	if err := mig.m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Info().Msg("Migration steps applied successfully!")
	return nil
}

func (mig *PostgresMigrator) Force(version int) error {
	if err := mig.m.Force(version); err != nil {
		return err
	}
	log.Info().Msgf("Forced migration to version: %d", version)
	return nil
}
