package postgres

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.log.Info().Msgf(format, v...)
}

func (l *Logger) Verbose() bool {
	return l.log.GetLevel() <= zerolog.DebugLevel
}

type Migrator struct {
	m   *migrate.Migrate
	log zerolog.Logger
}

func NewMigrator(migrationsPath, databaseURL string) (*Migrator, error) {
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return nil, err
	}

	logger := log.With().Str("module", "postgres_migrator").Logger()

	m.Log = &Logger{
		log: logger,
	}

	return &Migrator{
		m:   m,
		log: logger,
	}, nil
}

func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	m.log.Info().Msg("Migrations applied successfully!")
	return nil
}

func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	m.log.Info().Msg("Migrations rolled back successfully!")
	return nil
}

func (m *Migrator) Drop() error {
	if err := m.m.Drop(); err != nil {
		return err
	}
	m.log.Info().Msg("Migrations dropped successfully!")
	return nil
}
func (m *Migrator) Steps(steps int) error {
	if err := m.m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	m.log.Info().Msg("Migration steps applied successfully!")
	return nil
}

func (m *Migrator) Force(version int) error {
	if err := m.m.Force(version); err != nil {
		return err
	}
	m.log.Info().Msgf("Forced migration to version %d", version)
	return nil
}

func (m *Migrator) Close() {
	_, _ = m.m.Close()
}
