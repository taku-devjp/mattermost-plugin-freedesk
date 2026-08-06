package sqlstore

import (
	"database/sql"

	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
)

const locationColumns = `id, name, sort_order, create_at, update_at, delete_at`

// GetLocations returns all active locations.
func (s *SQLStore) GetLocations() ([]*freedeskmodel.Location, error) {
	query, args, err := s.builder.
		Select(locationColumns).
		From("freedesk_locations").
		Where("delete_at = 0").
		OrderBy("sort_order ASC").
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build locations query")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query locations")
	}
	defer func() { _ = rows.Close() }()

	var locations []*freedeskmodel.Location
	for rows.Next() {
		loc, scanErr := scanLocation(rows)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "failed to scan location")
		}
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

// GetDefaultLocation returns the first active location.
func (s *SQLStore) GetDefaultLocation() (*freedeskmodel.Location, error) {
	query, args, err := s.builder.
		Select(locationColumns).
		From("freedesk_locations").
		Where("delete_at = 0").
		OrderBy("sort_order ASC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build default location query")
	}

	row := s.db.QueryRow(query, args...)
	loc, err := scanLocation(row)
	if err == sql.ErrNoRows {
		return nil, errors.New("default location not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default location")
	}
	return loc, nil
}
