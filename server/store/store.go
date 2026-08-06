package store

import (
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
)

// Store defines database operations for the plugin.
type Store interface {
	Migrate() error
	SyncOneDeskPerDayIndex(enabled bool) error
	SeedInitialData() error

	GetLocations() ([]*model.Location, error)
	GetDefaultLocation() (*model.Location, error)

	GetDesks(locationID string, includeInactive bool) ([]*model.Desk, error)
	GetDesk(id string) (*model.Desk, error)
	CreateDesk(desk *model.Desk) error
	UpdateDesk(desk *model.Desk) error
	DeleteDesk(id string, now int64) error
	CountFutureReservationsForDesk(deskID, today string) (int, error)

	GetReservation(id string) (*model.Reservation, error)
	GetReservationsByDateRange(locationID, startDate, endDate string) ([]*model.Reservation, error)
	GetReservationsByUser(userID, today string, limit int) ([]*model.Reservation, error)
	CreateReservation(reservation *model.Reservation) error
	DeleteReservation(id string, now int64) error
}
