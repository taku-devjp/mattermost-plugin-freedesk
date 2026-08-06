package service

import (
	"database/sql"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/utils"
)

// GetDesks returns desks for a location.
func (s *Service) GetDesks(locationID string, includeInactive bool) ([]*freedeskmodel.Desk, error) {
	if locationID == "" {
		loc, err := s.store.GetDefaultLocation()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default location")
		}
		locationID = loc.ID
	}
	return s.store.GetDesks(locationID, includeInactive)
}

// CreateDesk creates a new desk (Plugin Admin).
func (s *Service) CreateDesk(req *freedeskmodel.CreateDeskRequest) (*freedeskmodel.Desk, error) {
	if req.LocationID == "" || req.Name == "" {
		return nil, newAPIError("INVALID_REQUEST", "location_id と name は必須です。", 400, nil)
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := model.GetMillis()
	desk := &freedeskmodel.Desk{
		ID:         model.NewId(),
		LocationID: req.LocationID,
		Name:       req.Name,
		SortOrder:  sortOrder,
		IsActive:   isActive,
		CreateAt:   now,
		UpdateAt:   now,
		DeleteAt:   0,
	}

	if err := s.store.CreateDesk(desk); err != nil {
		return nil, errors.Wrap(err, "failed to create desk")
	}

	s.client.Log.Info("Desk created", "desk_id", desk.ID, "name", desk.Name)
	return desk, nil
}

// UpdateDesk updates a desk (Plugin Admin).
func (s *Service) UpdateDesk(deskID string, req *freedeskmodel.UpdateDeskRequest) (*freedeskmodel.Desk, error) {
	desk, err := s.store.GetDesk(deskID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get desk")
	}
	if desk == nil {
		return nil, newAPIError("DESK_NOT_FOUND", "デスクが見つかりません。", 404, nil)
	}

	if req.Name != nil {
		desk.Name = *req.Name
	}
	if req.SortOrder != nil {
		desk.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		desk.IsActive = *req.IsActive
	}
	desk.UpdateAt = model.GetMillis()

	if err := s.store.UpdateDesk(desk); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, newAPIError("DESK_NOT_FOUND", "デスクが見つかりません。", 404, nil)
		}
		return nil, errors.Wrap(err, "failed to update desk")
	}

	s.client.Log.Info("Desk updated", "desk_id", desk.ID)
	return desk, nil
}

// DeleteDesk soft-deletes a desk (Plugin Admin).
func (s *Service) DeleteDesk(deskID string) error {
	desk, err := s.store.GetDesk(deskID)
	if err != nil {
		return errors.Wrap(err, "failed to get desk")
	}
	if desk == nil {
		return newAPIError("DESK_NOT_FOUND", "デスクが見つかりません。", 404, nil)
	}

	count, err := s.store.CountFutureReservationsForDesk(deskID, utils.Today())
	if err != nil {
		return errors.Wrap(err, "failed to count future reservations")
	}
	if count > 0 {
		return newAPIError("DESK_HAS_RESERVATIONS", "未来の予約が残っているため削除できません。", 409, nil)
	}

	now := model.GetMillis()
	if err := s.store.DeleteDesk(deskID, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newAPIError("DESK_NOT_FOUND", "デスクが見つかりません。", 404, nil)
		}
		return errors.Wrap(err, "failed to delete desk")
	}

	s.client.Log.Info("Desk deleted", "desk_id", deskID)
	return nil
}
