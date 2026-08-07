package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
)

const deskColumns = `id, location_id, name, description, sort_order, is_active, create_at, update_at, delete_at`

// GetDesks returns desks for a location.
func (s *SQLStore) GetDesks(locationID string, includeInactive bool) ([]*freedeskmodel.Desk, error) {
	q := s.builder.
		Select(deskColumns).
		From("freedesk_desks").
		Where("delete_at = 0").
		OrderBy("sort_order ASC")

	if locationID != "" {
		q = q.Where("location_id = ?", locationID)
	}
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build desks query")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query desks")
	}
	defer func() { _ = rows.Close() }()

	var desks []*freedeskmodel.Desk
	for rows.Next() {
		desk, scanErr := scanDesk(rows)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "failed to scan desk")
		}
		desks = append(desks, desk)
	}
	return desks, rows.Err()
}

// GetDesk returns a desk by ID.
func (s *SQLStore) GetDesk(id string) (*freedeskmodel.Desk, error) {
	query, args, err := s.builder.
		Select(deskColumns).
		From("freedesk_desks").
		Where("id = ? AND delete_at = 0", id).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build get desk query")
	}

	row := s.db.QueryRow(query, args...)
	desk, err := scanDesk(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get desk")
	}
	return desk, nil
}

// CreateDesk inserts a new desk.
func (s *SQLStore) CreateDesk(desk *freedeskmodel.Desk) error {
	query, args, err := s.builder.
		Insert("freedesk_desks").
		Columns("id", "location_id", "name", "description", "sort_order", "is_active", "create_at", "update_at", "delete_at").
		Values(desk.ID, desk.LocationID, desk.Name, desk.Description, desk.SortOrder, desk.IsActive, desk.CreateAt, desk.UpdateAt, desk.DeleteAt).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build create desk query")
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %v", ErrUniqueViolation, err)
		}
		return errors.Wrap(err, "failed to create desk")
	}
	return nil
}

// UpdateDesk updates desk fields.
func (s *SQLStore) UpdateDesk(desk *freedeskmodel.Desk) error {
	query, args, err := s.builder.
		Update("freedesk_desks").
		Set("name", desk.Name).
		Set("sort_order", desk.SortOrder).
		Set("is_active", desk.IsActive).
		Set("update_at", desk.UpdateAt).
		Where("id = ? AND delete_at = 0", desk.ID).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build update desk query")
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %v", ErrUniqueViolation, err)
		}
		return errors.Wrap(err, "failed to update desk")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteDesk soft-deletes a desk.
func (s *SQLStore) DeleteDesk(id string, now int64) error {
	query, args, err := s.builder.
		Update("freedesk_desks").
		Set("delete_at", now).
		Set("update_at", now).
		Where("id = ? AND delete_at = 0", id).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete desk query")
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete desk")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountFutureReservationsForDesk counts active reservations on or after today.
func (s *SQLStore) CountFutureReservationsForDesk(deskID, today string) (int, error) {
	query, args, err := s.builder.
		Select("COUNT(*)").
		From("freedesk_reservations").
		Where("desk_id = ? AND delete_at = 0 AND reserve_date >= ?", deskID, today).
		ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build count reservations query")
	}

	var count int
	if err := s.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to count future reservations")
	}
	return count, nil
}
