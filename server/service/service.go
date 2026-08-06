package service

import (
	"github.com/mattermost/mattermost/server/public/pluginapi"

	"github.com/taku-devjp/mattermost-plugin-freedesk/server/store"
)

// ConfigProvider supplies plugin configuration.
type ConfigProvider interface {
	GetNotificationChannelID() string
	GetEnableNotifications() bool
	GetMaxAdvanceDays() int
	GetOneDeskPerDay() bool
	GetPluginAdminUserIDs() []string
	IsPluginAdmin(userID string) bool
}

// Service holds business logic dependencies.
type Service struct {
	store     store.Store
	client    *pluginapi.Client
	config    ConfigProvider
	botUserID string
}

// New creates a Service.
func New(s store.Store, client *pluginapi.Client, config ConfigProvider) *Service {
	return &Service{
		store:  s,
		client: client,
		config: config,
	}
}
