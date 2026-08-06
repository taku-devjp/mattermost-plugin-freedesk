package sqlstore

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
)

// GetReservation returns a reservation by ID.
func (s *SQLStore) GetReservation(id string) (*freedeskmodel.Reservation, error) {
	query, args, err := s.builder.
		Select("r.id", "r.desk_id", "r.user_id", "r.reserve_date", "r.note", "r.create_at", "r.update_at", "r.delete_at", "d.name AS desk_name").
		From("freedesk_reservations r").
		Join("freedesk_desks d ON d.id = r.desk_id").
		Where("r.id = ? AND r.delete_at = 0", id).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build get reservation query")
	}

	row := s.db.QueryRow(query, args...)
	res, err := scanReservationWithDesk(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get reservation")
	}
	return res, nil
}

func scanReservationWithDesk(scanner interface{ Scan(...any) error }) (*freedeskmodel.Reservation, error) {
	var res freedeskmodel.Reservation
	var note sql.NullString
	if err := scanner.Scan(
		&res.ID, &res.DeskID, &res.UserID, &res.ReserveDate, &note,
		&res.CreateAt, &res.UpdateAt, &res.DeleteAt, &res.DeskName,
	); err != nil {
		return nil, err
	}
	if note.Valid {
		res.Note = &note.String
	}
	return &res, nil
}

// GetReservationsByDateRange returns active reservations in a date range for active desks.
func (s *SQLStore) GetReservationsByDateRange(locationID, startDate, endDate string) ([]*freedeskmodel.Reservation, error) {
	q := s.builder.
		Select("r.id", "r.desk_id", "r.user_id", "r.reserve_date", "r.note", "r.create_at", "r.update_at", "r.delete_at", "d.name AS desk_name").
		From("freedesk_reservations r").
		Join("freedesk_desks d ON d.id = r.desk_id").
		Where("r.delete_at = 0").
		Where("d.delete_at = 0").
		Where("d.is_active = ?", true).
		Where("r.reserve_date >= ?", startDate).
		Where("r.reserve_date <= ?", endDate).
		OrderBy("r.reserve_date ASC", "d.sort_order ASC")

	if locationID != "" {
		q = q.Where("d.location_id = ?", locationID)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build reservations by date range query")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query reservations")
	}
	defer rows.Close()

	var reservations []*freedeskmodel.Reservation
	for rows.Next() {
		res, scanErr := scanReservationWithDesk(rows)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "failed to scan reservation")
		}
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}

// GetReservationsByUser returns a user's active reservations from today onward.
func (s *SQLStore) GetReservationsByUser(userID, today string, limit int) ([]*freedeskmodel.Reservation, error) {
	if limit <= 0 {
		limit = 50
	}

	query, args, err := s.builder.
		Select("r.id", "r.desk_id", "r.user_id", "r.reserve_date", "r.note", "r.create_at", "r.update_at", "r.delete_at", "d.name AS desk_name").
		From("freedesk_reservations r").
		Join("freedesk_desks d ON d.id = r.desk_id").
		Where("r.user_id = ? AND r.delete_at = 0 AND r.reserve_date >= ?", userID, today).
		OrderBy("r.reserve_date ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build user reservations query")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user reservations")
	}
	defer rows.Close()

	var reservations []*freedeskmodel.Reservation
	for rows.Next() {
		res, scanErr := scanReservationWithDesk(rows)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "failed to scan reservation")
		}
		reservations = append(reservations, res)
	}
	return reservations, rows.Err()
}

// CreateReservation inserts a new reservation.
func (s *SQLStore) CreateReservation(reservation *freedeskmodel.Reservation) error {
	query, args, err := s.builder.
		Insert("freedesk_reservations").
		Columns("id", "desk_id", "user_id", "reserve_date", "note", "create_at", "update_at", "delete_at").
		Values(reservation.ID, reservation.DeskID, reservation.UserID, reservation.ReserveDate, reservation.Note, reservation.CreateAt, reservation.UpdateAt, reservation.DeleteAt).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build create reservation query")
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %v", ErrUniqueViolation, err)
		}
		return errors.Wrap(err, "failed to create reservation")
	}
	return nil
}

// DeleteReservation soft-deletes a reservation.
func (s *SQLStore) DeleteReservation(id string, now int64) error {
	query, args, err := s.builder.
		Update("freedesk_reservations").
		Set("delete_at", now).
		Set("update_at", now).
		Where("id = ? AND delete_at = 0", id).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build delete reservation query")
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete reservation")
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

// ErrUniqueViolation indicates a unique constraint violation.
var ErrUniqueViolation = errors.New("unique constraint violation")
