package service

import (
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/utils"
)

// GetConfig returns frontend configuration for the current user.
func (s *Service) GetConfig(userID string) *model.ConfigData {
	maxDays := s.config.GetMaxAdvanceDays()
	return &model.ConfigData{
		Timezone:            utils.Timezone,
		Today:               utils.Today(),
		MaxAdvanceDays:      maxDays,
		BookableUntil:       utils.BookableUntil(maxDays),
		OneDeskPerDay:       s.config.GetOneDeskPerDay(),
		NotificationEnabled: s.config.GetEnableNotifications(),
		IsPluginAdmin:       s.config.IsPluginAdmin(userID),
	}
}

// GetLocations returns all active locations.
func (s *Service) GetLocations() ([]*model.Location, error) {
	return s.store.GetLocations()
}
