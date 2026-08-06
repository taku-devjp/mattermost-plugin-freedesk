package main

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"

	"github.com/taku-devjp/mattermost-plugin-freedesk/server/service"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/store"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/store/sqlstore"
)

// Plugin implements the interface expected by the Mattermost server.
type Plugin struct {
	plugin.MattermostPlugin

	client *pluginapi.Client
	store  store.Store
	service *service.Service
	router *mux.Router

	configurationLock sync.RWMutex
	configuration   *configuration
}

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	sqlStore, err := sqlstore.New(p.client)
	if err != nil {
		return errors.Wrap(err, "failed to create sql store")
	}
	p.store = sqlStore

	if err := p.store.Migrate(); err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	if err := p.store.SeedInitialData(); err != nil {
		return errors.Wrap(err, "failed to seed initial data")
	}

	if err := p.OnConfigurationChange(); err != nil {
		return errors.Wrap(err, "failed to load configuration")
	}

	if err := p.store.SyncOneDeskPerDayIndex(p.getConfiguration().GetOneDeskPerDay()); err != nil {
		return errors.Wrap(err, "failed to sync OneDeskPerDay index on activate")
	}

	p.service = service.New(p.store, p.client, p)
	p.router = p.initRouter()

	p.API.LogInfo("Free Desk plugin activated")
	return nil
}

// ConfigProvider methods — delegate to active configuration.

func (p *Plugin) GetNotificationChannelID() string {
	return p.getConfiguration().GetNotificationChannelID()
}

func (p *Plugin) GetEnableNotifications() bool {
	return p.getConfiguration().GetEnableNotifications()
}

func (p *Plugin) GetMaxAdvanceDays() int {
	return p.getConfiguration().GetMaxAdvanceDays()
}

func (p *Plugin) GetOneDeskPerDay() bool {
	return p.getConfiguration().GetOneDeskPerDay()
}

func (p *Plugin) GetPluginAdminUserIDs() []string {
	return p.getConfiguration().GetPluginAdminUserIDs()
}

func (p *Plugin) IsPluginAdmin(userID string) bool {
	return p.getConfiguration().IsPluginAdmin(userID)
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	p.API.LogInfo("Free Desk plugin deactivated")
	return nil
}
