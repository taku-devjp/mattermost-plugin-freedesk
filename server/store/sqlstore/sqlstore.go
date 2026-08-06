package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/store"
)

const migrationVersion = 1

const (
	driverPostgres = "postgres"
	driverMysql    = "mysql"
	driverSqlite   = "sqlite3"
)

// SQLStore implements store.Store using the Mattermost plugin SQL driver.
type SQLStore struct {
	db         *sql.DB
	driverName string
	builder    sq.StatementBuilderType
}

// New creates a SQLStore backed by the plugin API store service.
func New(client *pluginapi.Client) (store.Store, error) {
	db, err := client.Store.GetMasterDB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get master db")
	}

	driverName := client.Store.DriverName()
	builder := sq.StatementBuilder.PlaceholderFormat(placeholderFormat(driverName))

	return &SQLStore{
		db:         db,
		driverName: driverName,
		builder:    builder,
	}, nil
}

func placeholderFormat(driverName string) sq.PlaceholderFormat {
	if driverName == model.DatabaseDriverPostgres {
		return sq.Dollar
	}
	return sq.Question
}

func (s *SQLStore) isPostgres() bool {
	return s.driverName == model.DatabaseDriverPostgres
}

// selectReserveDateColumn returns reserve_date as text for consistent scanning across drivers.
func (s *SQLStore) selectReserveDateColumn() string {
	switch s.driverName {
	case driverPostgres:
		return "TO_CHAR(r.reserve_date, 'YYYY-MM-DD') AS reserve_date"
	case driverMysql:
		return "DATE_FORMAT(r.reserve_date, '%Y-%m-%d') AS reserve_date"
	default:
		return "r.reserve_date"
	}
}

func (s *SQLStore) tableExists(tableName string) (bool, error) {
	var query string
	switch s.driverName {
	case driverPostgres:
		query = `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`
	case driverMysql:
		query = `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`
	default:
		query = `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`
	}

	var count int
	if err := s.db.QueryRow(query, tableName).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLStore) getAppliedVersion() (int, error) {
	exists, err := s.tableExists("freedesk_migrations")
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	var version int
	err = s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM freedesk_migrations`).Scan(&version)
	return version, err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Migrate applies pending schema migrations.
func (s *SQLStore) Migrate() error {
	version, err := s.getAppliedVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get applied migration version")
	}
	if version >= migrationVersion {
		return nil
	}

	stmts, err := s.migrationV001Statements()
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin migration transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, stmt := range stmts {
		if _, execErr := tx.Exec(stmt); execErr != nil {
			err = errors.Wrapf(execErr, "failed to execute migration statement: %s", truncate(stmt, 80))
			return err
		}
	}

	now := model.GetMillis()
	if s.isPostgres() {
		_, err = tx.Exec(`INSERT INTO freedesk_migrations (version, applied_at) VALUES ($1, $2)`, migrationVersion, now)
	} else {
		_, err = tx.Exec(`INSERT INTO freedesk_migrations (version, applied_at) VALUES (?, ?)`, migrationVersion, now)
	}
	if err != nil {
		return errors.Wrap(err, "failed to record migration")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit migration")
	}
	return nil
}

func (s *SQLStore) indexExists(indexName string) (bool, error) {
	var query string
	switch s.driverName {
	case driverPostgres:
		query = `SELECT COUNT(*) FROM pg_indexes WHERE indexname = $1`
	case driverMysql:
		query = `SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND index_name = ?`
	default:
		query = `SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?`
	}

	var count int
	err := s.db.QueryRow(query, indexName).Scan(&count)
	return count > 0, err
}

// SyncOneDeskPerDayIndex creates or drops the user-date unique index per configuration.
func (s *SQLStore) SyncOneDeskPerDayIndex(enabled bool) error {
	const indexName = "uq_freedesk_reservations_user_date"

	exists, err := s.indexExists(indexName)
	if err != nil {
		return errors.Wrap(err, "failed to check user_date index")
	}

	if enabled && !exists {
		if err := s.createUserDateIndex(); err != nil {
			return err
		}
	} else if !enabled && exists {
		if err := s.dropIndex(indexName); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) createUserDateIndex() error {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT user_id, reserve_date, COUNT(*) AS cnt
			FROM freedesk_reservations
			WHERE delete_at = 0
			GROUP BY user_id, reserve_date
			HAVING COUNT(*) > 1
		) dup`).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "failed to check duplicate reservations")
	}
	if count > 0 {
		return errors.New("cannot enable OneDeskPerDay: duplicate active user+date reservations exist")
	}

	stmt := s.userDateIndexDDL()
	_, err = s.db.Exec(stmt)
	return errors.Wrap(err, "failed to create user_date index")
}

func (s *SQLStore) dropIndex(indexName string) error {
	var stmt string
	switch s.driverName {
	case driverPostgres:
		stmt = fmt.Sprintf(`DROP INDEX IF EXISTS %s`, indexName)
	case driverMysql:
		stmt = fmt.Sprintf(`DROP INDEX %s ON freedesk_reservations`, indexName)
	default:
		stmt = fmt.Sprintf(`DROP INDEX IF EXISTS %s`, indexName)
	}
	_, err := s.db.Exec(stmt)
	return errors.Wrapf(err, "failed to drop index %s", indexName)
}

func (s *SQLStore) userDateIndexDDL() string {
	if s.isPostgres() {
		return `CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_user_date
			ON freedesk_reservations (user_id, reserve_date) WHERE delete_at = 0`
	}
	return `CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_user_date
		ON freedesk_reservations (user_id, reserve_date, delete_at)`
}

func (s *SQLStore) migrationV001Statements() ([]string, error) {
	switch s.driverName {
	case driverPostgres:
		return postgresMigrationV001(), nil
	case driverMysql:
		return mysqlMigrationV001(), nil
	case driverSqlite:
		return sqliteMigrationV001(), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", s.driverName)
	}
}

func postgresMigrationV001() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS freedesk_migrations (
			version     INTEGER      NOT NULL PRIMARY KEY,
			applied_at  BIGINT       NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS freedesk_locations (
			id          VARCHAR(26)  NOT NULL PRIMARY KEY,
			name        VARCHAR(128) NOT NULL,
			sort_order  INTEGER      NOT NULL DEFAULT 0,
			create_at   BIGINT       NOT NULL,
			update_at   BIGINT       NOT NULL,
			delete_at   BIGINT       NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_locations_delete_at ON freedesk_locations (delete_at)`,
		`CREATE TABLE IF NOT EXISTS freedesk_desks (
			id          VARCHAR(26)  NOT NULL PRIMARY KEY,
			location_id VARCHAR(26)  NOT NULL REFERENCES freedesk_locations(id),
			name        VARCHAR(64)  NOT NULL,
			description VARCHAR(256),
			sort_order  INTEGER      NOT NULL DEFAULT 0,
			is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
			create_at   BIGINT       NOT NULL,
			update_at   BIGINT       NOT NULL,
			delete_at   BIGINT       NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_desks_location_id ON freedesk_desks (location_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_desks_location_name
			ON freedesk_desks (location_id, name) WHERE delete_at = 0`,
		`CREATE TABLE IF NOT EXISTS freedesk_reservations (
			id           VARCHAR(26) NOT NULL PRIMARY KEY,
			desk_id      VARCHAR(26) NOT NULL REFERENCES freedesk_desks(id),
			user_id      VARCHAR(26) NOT NULL,
			reserve_date DATE        NOT NULL,
			note         VARCHAR(256),
			create_at    BIGINT      NOT NULL,
			update_at    BIGINT      NOT NULL,
			delete_at    BIGINT      NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_desk_date
			ON freedesk_reservations (desk_id, reserve_date) WHERE delete_at = 0`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_user_date
			ON freedesk_reservations (user_id, reserve_date) WHERE delete_at = 0`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_date ON freedesk_reservations (reserve_date, delete_at)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_user_id ON freedesk_reservations (user_id, delete_at)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_desk_id ON freedesk_reservations (desk_id)`,
	}
}

func mysqlMigrationV001() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS freedesk_migrations (
			version     INT          NOT NULL PRIMARY KEY,
			applied_at  BIGINT       NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS freedesk_locations (
			id          VARCHAR(26)  NOT NULL PRIMARY KEY,
			name        VARCHAR(128) NOT NULL,
			sort_order  INT          NOT NULL DEFAULT 0,
			create_at   BIGINT       NOT NULL,
			update_at   BIGINT       NOT NULL,
			delete_at   BIGINT       NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX idx_freedesk_locations_delete_at ON freedesk_locations (delete_at)`,
		`CREATE TABLE IF NOT EXISTS freedesk_desks (
			id          VARCHAR(26)  NOT NULL PRIMARY KEY,
			location_id VARCHAR(26)  NOT NULL,
			name        VARCHAR(64)  NOT NULL,
			description VARCHAR(256),
			sort_order  INT          NOT NULL DEFAULT 0,
			is_active   TINYINT(1)   NOT NULL DEFAULT 1,
			create_at   BIGINT       NOT NULL,
			update_at   BIGINT       NOT NULL,
			delete_at   BIGINT       NOT NULL DEFAULT 0,
			INDEX idx_freedesk_desks_location_id (location_id)
		)`,
		`CREATE UNIQUE INDEX uq_freedesk_desks_location_name ON freedesk_desks (location_id, name, delete_at)`,
		`CREATE TABLE IF NOT EXISTS freedesk_reservations (
			id           VARCHAR(26) NOT NULL PRIMARY KEY,
			desk_id      VARCHAR(26) NOT NULL,
			user_id      VARCHAR(26) NOT NULL,
			reserve_date DATE        NOT NULL,
			note         VARCHAR(256),
			create_at    BIGINT      NOT NULL,
			update_at    BIGINT      NOT NULL,
			delete_at    BIGINT      NOT NULL DEFAULT 0,
			INDEX idx_freedesk_reservations_date (reserve_date, delete_at),
			INDEX idx_freedesk_reservations_user_id (user_id, delete_at),
			INDEX idx_freedesk_reservations_desk_id (desk_id)
		)`,
		`CREATE UNIQUE INDEX uq_freedesk_reservations_desk_date ON freedesk_reservations (desk_id, reserve_date, delete_at)`,
		`CREATE UNIQUE INDEX uq_freedesk_reservations_user_date ON freedesk_reservations (user_id, reserve_date, delete_at)`,
	}
}

func sqliteMigrationV001() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS freedesk_migrations (
			version     INTEGER      NOT NULL PRIMARY KEY,
			applied_at  INTEGER      NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS freedesk_locations (
			id          TEXT         NOT NULL PRIMARY KEY,
			name        TEXT         NOT NULL,
			sort_order  INTEGER      NOT NULL DEFAULT 0,
			create_at   INTEGER      NOT NULL,
			update_at   INTEGER      NOT NULL,
			delete_at   INTEGER      NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_locations_delete_at ON freedesk_locations (delete_at)`,
		`CREATE TABLE IF NOT EXISTS freedesk_desks (
			id          TEXT         NOT NULL PRIMARY KEY,
			location_id TEXT         NOT NULL,
			name        TEXT         NOT NULL,
			description TEXT,
			sort_order  INTEGER      NOT NULL DEFAULT 0,
			is_active   INTEGER      NOT NULL DEFAULT 1,
			create_at   INTEGER      NOT NULL,
			update_at   INTEGER      NOT NULL,
			delete_at   INTEGER      NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_desks_location_id ON freedesk_desks (location_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_desks_location_name ON freedesk_desks (location_id, name, delete_at)`,
		`CREATE TABLE IF NOT EXISTS freedesk_reservations (
			id           TEXT    NOT NULL PRIMARY KEY,
			desk_id      TEXT    NOT NULL,
			user_id      TEXT    NOT NULL,
			reserve_date TEXT    NOT NULL,
			note         TEXT,
			create_at    INTEGER NOT NULL,
			update_at    INTEGER NOT NULL,
			delete_at    INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_desk_date ON freedesk_reservations (desk_id, reserve_date, delete_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_freedesk_reservations_user_date ON freedesk_reservations (user_id, reserve_date, delete_at)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_date ON freedesk_reservations (reserve_date, delete_at)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_user_id ON freedesk_reservations (user_id, delete_at)`,
		`CREATE INDEX IF NOT EXISTS idx_freedesk_reservations_desk_id ON freedesk_reservations (desk_id)`,
	}
}

// SeedInitialData creates the default location and 7 desks when missing.
func (s *SQLStore) SeedInitialData() error {
	now := model.GetMillis()

	var locationCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM freedesk_locations WHERE delete_at = 0`).Scan(&locationCount); err != nil {
		return errors.Wrap(err, "failed to count locations")
	}

	var locationID string
	if locationCount == 0 {
		locationID = model.NewId()
		if s.isPostgres() {
			_, err := s.db.Exec(
				`INSERT INTO freedesk_locations (id, name, sort_order, create_at, update_at, delete_at) VALUES ($1, $2, $3, $4, $5, $6)`,
				locationID, "Default", 0, now, now, 0,
			)
			if err != nil {
				return errors.Wrap(err, "failed to seed default location")
			}
		} else {
			_, err := s.db.Exec(
				`INSERT INTO freedesk_locations (id, name, sort_order, create_at, update_at, delete_at) VALUES (?, ?, ?, ?, ?, ?)`,
				locationID, "Default", 0, now, now, 0,
			)
			if err != nil {
				return errors.Wrap(err, "failed to seed default location")
			}
		}
	} else {
		loc, err := s.GetDefaultLocation()
		if err != nil {
			return err
		}
		locationID = loc.ID
	}

	var deskCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM freedesk_desks WHERE delete_at = 0`).Scan(&deskCount); err != nil {
		return errors.Wrap(err, "failed to count desks")
	}
	if deskCount > 0 {
		return nil
	}

	deskNames := []string{
		"社員フリー1", "社員フリー2", "社員フリー3",
		"フリーデスク4", "フリーデスク5", "フリーデスク6", "フリーデスク7",
	}

	for i, name := range deskNames {
		desk := &freedeskmodel.Desk{
			ID:         model.NewId(),
			LocationID: locationID,
			Name:       name,
			SortOrder:  i,
			IsActive:   true,
			CreateAt:   now,
			UpdateAt:   now,
			DeleteAt:   0,
		}
		if err := s.CreateDesk(desk); err != nil {
			return errors.Wrapf(err, "failed to seed desk %s", name)
		}
	}
	return nil
}

func scanDesk(scanner interface{ Scan(...any) error }) (*freedeskmodel.Desk, error) {
	var desk freedeskmodel.Desk
	var description sql.NullString
	if err := scanner.Scan(
		&desk.ID, &desk.LocationID, &desk.Name, &description,
		&desk.SortOrder, &desk.IsActive, &desk.CreateAt, &desk.UpdateAt, &desk.DeleteAt,
	); err != nil {
		return nil, err
	}
	if description.Valid {
		desk.Description = &description.String
	}
	return &desk, nil
}

func scanLocation(scanner interface{ Scan(...any) error }) (*freedeskmodel.Location, error) {
	var loc freedeskmodel.Location
	if err := scanner.Scan(&loc.ID, &loc.Name, &loc.SortOrder, &loc.CreateAt, &loc.UpdateAt, &loc.DeleteAt); err != nil {
		return nil, err
	}
	return &loc, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}
