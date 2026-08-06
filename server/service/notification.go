package service

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
)

// SetBotUserID sets the bot user ID used for channel notifications.
func (s *Service) SetBotUserID(botUserID string) {
	s.botUserID = botUserID
}

func (s *Service) notifyReservationCreated(reservation *freedeskmodel.Reservation, deskName string) {
	if !s.config.GetEnableNotifications() {
		return
	}
	channelID := s.config.GetNotificationChannelID()
	if channelID == "" || s.botUserID == "" {
		return
	}

	userName := s.resolveUserName(reservation.UserID)
	message := fmt.Sprintf("%s が %s を %s に予約しました。", userName, deskName, reservation.ReserveDate)
	s.postNotificationAsync(channelID, message)
}

func (s *Service) notifyReservationDeleted(reservation *freedeskmodel.Reservation, userName string) {
	if !s.config.GetEnableNotifications() {
		return
	}
	channelID := s.config.GetNotificationChannelID()
	if channelID == "" || s.botUserID == "" {
		return
	}

	deskName := reservation.DeskName
	if deskName == "" {
		if desk, err := s.store.GetDesk(reservation.DeskID); err == nil && desk != nil {
			deskName = desk.Name
		}
	}

	message := fmt.Sprintf("%s が %s の %s 予約を取り消しました。", userName, reservation.ReserveDate, deskName)
	s.postNotificationAsync(channelID, message)
}

func (s *Service) postNotificationAsync(channelID, message string) {
	go func() {
		post := &model.Post{
			UserId:    s.botUserID,
			ChannelId: channelID,
			Message:   message,
		}
		if err := s.client.Post.CreatePost(post); err != nil {
			s.client.Log.Info("Failed to send channel notification", "error", err.Error(), "channel_id", channelID)
			return
		}
		s.client.Log.Info("Channel notification sent", "channel_id", channelID)
	}()
}
