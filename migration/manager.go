package migration

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const (
	migrationDuration = metric.Duration("MigrationDuration")
)

type Manager struct {
	logger         lager.Logger
	db             db.DB
	storeClient    etcd.StoreClient
	migrations     []Migration
	migrationsDone chan<- struct{}
	clock          clock.Clock
}

func NewManager(
	logger lager.Logger,
	db db.DB,
	storeClient etcd.StoreClient,
	migrations Migrations,
	migrationsDone chan<- struct{},
	clock clock.Clock,
) Manager {
	sort.Sort(migrations)

	return Manager{
		logger:         logger,
		db:             db,
		storeClient:    storeClient,
		migrations:     migrations,
		migrationsDone: migrationsDone,
		clock:          clock,
	}
}

func (m Manager) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := m.logger.Session("migration-manager")

	version, err := m.db.Version(logger)
	if err != nil {
		if models.ConvertError(err) == models.ErrResourceNotFound {
			version = &models.Version{}
			err = m.db.SetVersion(m.logger, version)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if version.TargetVersion < version.CurrentVersion {
		return fmt.Errorf(
			"Existing DB target version (%d) exceeds current version (%d)",
			version.TargetVersion,
			version.CurrentVersion,
		)
	}

	var bbsMigrationVersion int64
	if len(m.migrations) > 0 {
		bbsMigrationVersion = m.migrations[len(m.migrations)-1].Version()
	}

	if version.CurrentVersion > bbsMigrationVersion {
		return fmt.Errorf(
			"Existing DB version (%d) exceeds bbs version (%d)",
			version.CurrentVersion,
			bbsMigrationVersion,
		)
	}

	if version.TargetVersion != bbsMigrationVersion {
		if version.TargetVersion > bbsMigrationVersion {
			version.TargetVersion = bbsMigrationVersion
		}

		m.writeVersion(version.CurrentVersion, bbsMigrationVersion)
	}

	migrateStart := m.clock.Now()
	if version.CurrentVersion != bbsMigrationVersion {
		lastVersion := version.CurrentVersion
		nextVersion := version.CurrentVersion

		for _, currentMigration := range m.migrations {
			if lastVersion < currentMigration.Version() {
				if nextVersion > currentMigration.Version() {
					return fmt.Errorf(
						"Existing DB target version (%d) exceeds pending migration version (%d)",
						nextVersion,
						currentMigration.Version(),
					)
				}
				nextVersion = currentMigration.Version()

				logger.Debug("running-migration", lager.Data{
					"CurrentVersion":   lastVersion,
					"NextVersion":      nextVersion,
					"MigrationVersion": currentMigration.Version(),
				})
				currentMigration.SetStoreClient(m.storeClient)
				err = currentMigration.Up(m.logger)
				if err != nil {
					return err
				}

				lastVersion = currentMigration.Version()
				logger.Debug("completed-migration")
			}
		}

		err = m.writeVersion(lastVersion, nextVersion)
		if err != nil {
			return err
		}
	}

	close(ready)
	close(m.migrationsDone)

	logger.Debug("migrations-finished")
	migrationDuration.Send(time.Since(migrateStart))

	select {
	case <-signals:
		return nil
	}
}

func (m *Manager) writeVersion(currentVersion, targetVersion int64) error {
	return m.db.SetVersion(m.logger, &models.Version{
		CurrentVersion: currentVersion,
		TargetVersion:  targetVersion,
	})
}

type Migrations []Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Migrations) Less(i, j int) bool { return m[i].Version() < m[j].Version() }
