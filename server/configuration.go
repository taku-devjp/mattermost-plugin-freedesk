package main

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/taku-devjp/mattermost-plugin-freedesk/server/utils"
)

type configuration struct {
	NotificationChannelID string  `json:"NotificationChannelID"`
	EnableNotifications   *bool   `json:"EnableNotifications"`
	MaxAdvanceDays        *int    `json:"MaxAdvanceDays"`
	OneDeskPerDay         *bool   `json:"OneDeskPerDay"`
	PluginAdminUserIDs    string  `json:"PluginAdminUserIDs"`
}

func (c *configuration) Clone() *configuration {
	clone := *c
	return &clone
}

func (c *configuration) GetNotificationChannelID() string {
	if c == nil {
		return ""
	}
	return c.NotificationChannelID
}

func (c *configuration) GetEnableNotifications() bool {
	if c == nil || c.EnableNotifications == nil {
		return true
	}
	return *c.EnableNotifications
}

func (c *configuration) GetMaxAdvanceDays() int {
	if c == nil || c.MaxAdvanceDays == nil || *c.MaxAdvanceDays <= 0 {
		return utils.DefaultMaxDays
	}
	return *c.MaxAdvanceDays
}

func (c *configuration) GetOneDeskPerDay() bool {
	if c == nil || c.OneDeskPerDay == nil {
		return true
	}
	return *c.OneDeskPerDay
}

func (c *configuration) GetPluginAdminUserIDs() []string {
	if c == nil || c.PluginAdminUserIDs == "" {
		return nil
	}
	parts := strings.Split(c.PluginAdminUserIDs, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		id := strings.TrimSpace(p)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c *configuration) IsPluginAdmin(userID string) bool {
	for _, id := range c.GetPluginAdminUserIDs() {
		if id == userID {
			return true
		}
	}
	return false
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}
	return p.configuration
}

func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}
		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

func (p *Plugin) OnConfigurationChange() error {
	prev := p.getConfiguration()
	prevOneDesk := true
	if prev != nil {
		prevOneDesk = prev.GetOneDeskPerDay()
	}

	configuration := new(configuration)
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	if p.store != nil && prevOneDesk != configuration.GetOneDeskPerDay() {
		if err := p.store.SyncOneDeskPerDayIndex(configuration.GetOneDeskPerDay()); err != nil {
			p.API.LogError("Failed to sync OneDeskPerDay index", "error", err.Error())
			return errors.Wrap(err, "failed to sync OneDeskPerDay index")
		}
	}

	return nil
}
