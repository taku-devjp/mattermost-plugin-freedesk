package service

import freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"

// notifyReservationCreated posts a channel notification (Phase 2 stub).
func (s *Service) notifyReservationCreated(reservation *freedeskmodel.Reservation, deskName string) {
	if !s.config.GetEnableNotifications() {
		return
	}
	channelID := s.config.GetNotificationChannelID()
	if channelID == "" {
		return
	}
	// Phase 2: async channel post
	s.client.Log.Info("Reservation created (notification pending Phase 2)",
		"reservation_id", reservation.ID,
		"desk", deskName,
		"date", reservation.ReserveDate,
		"channel_id", channelID,
	)
}

// notifyReservationDeleted posts a cancellation notification (Phase 2 stub).
func (s *Service) notifyReservationDeleted(reservation *freedeskmodel.Reservation, userName string) {
	if !s.config.GetEnableNotifications() {
		return
	}
	channelID := s.config.GetNotificationChannelID()
	if channelID == "" {
		return
	}
	s.client.Log.Info("Reservation cancelled (notification pending Phase 2)",
		"reservation_id", reservation.ID,
		"user", userName,
		"date", reservation.ReserveDate,
		"channel_id", channelID,
	)
}
