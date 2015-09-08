package migration

import (
	"fmt"
	"os"
	"sort"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type Manager struct {
	logger         lager.Logger
	db             db.DB
	storeClient    etcd.StoreClient
	migrations     []Migration
	migrationsDone chan<- struct{}
}

func NewManager(
	logger lager.Logger,
	db db.DB,
	storeClient etcd.StoreClient,
	migrations Migrations,
	migrationsDone chan<- struct{},
) Manager {
	sort.Sort(migrations)

	return Manager{
		logger:         logger,
		db:             db,
		storeClient:    storeClient,
		migrations:     migrations,
		migrationsDone: migrationsDone,
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

	if version.TargetVersion > bbsMigrationVersion {
		return fmt.Errorf(
			"Existing DB target version (%d) exceeds final migration version (%d)",
			version.TargetVersion,
			bbsMigrationVersion,
		)
	}

	currentVersion := version.CurrentVersion
	targetVersion := version.TargetVersion
	for _, migration := range m.migrations {
		if currentVersion < migration.Version() {
			if targetVersion > migration.Version() {
				return fmt.Errorf(
					"Existing DB target version (%d) exceeds pending migration version (%d)",
					targetVersion,
					migration.Version(),
				)
			}
			targetVersion = migration.Version()

			err := m.db.SetVersion(m.logger, &models.Version{
				CurrentVersion: currentVersion,
				TargetVersion:  targetVersion,
			})
			if err != nil {
				return err
			}

			logger.Debug("running-migration", lager.Data{
				"CurrentVersion":   currentVersion,
				"TargetVersion":    targetVersion,
				"MigrationVersion": migration.Version(),
			})
			err = migration.Up(m.logger, m.storeClient)
			if err != nil {
				return err
			}

			currentVersion = migration.Version()
			err = m.db.SetVersion(m.logger, &models.Version{
				CurrentVersion: currentVersion,
				TargetVersion:  targetVersion,
			})
			if err != nil {
				return err
			}

			logger.Debug("completed-migration")
		}
	}

	close(ready)
	close(m.migrationsDone)

	logger.Debug("migrations-finished")

	select {
	case <-signals:
		return nil
	}
}

type Migrations []Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Migrations) Less(i, j int) bool { return m[i].Version() < m[j].Version() }

//go:generate counterfeiter -o migrationfakes/fake_migration.go . Migration
type Migration interface {
	Version() int64
	Up(logger lager.Logger, storeClient etcd.StoreClient) error
	Down(logger lager.Logger, storeClient etcd.StoreClient) error
}
