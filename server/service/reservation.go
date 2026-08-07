package service

import (
	"database/sql"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/store/sqlstore"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/utils"
)

// APIError represents a business error with HTTP status and error code.
type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
	Details    map[string]any
}

func (e *APIError) Error() string {
	return e.Message
}

func newAPIError(code, message string, status int, details map[string]any) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Details:    details,
	}
}

// GetMatrix returns matrix data for the given month.
func (s *Service) GetMatrix(userID string, year, month int, locationID string) (*freedeskmodel.MatrixData, error) {
	today := utils.Today()
	maxMonths := s.config.GetMaxAdvanceMonths()
	bookableUntil := utils.BookableUntil(maxMonths)
	bookableYear, bookableMonth, err := parseYearMonth(bookableUntil)
	if err != nil {
		return nil, newAPIError("INVALID_REQUEST", "予約可能期間の算出に失敗しました。", 400, nil)
	}

	if year == 0 || month == 0 {
		year, month = utils.CurrentYearMonth()
	}

	curYear, curMonth := utils.CurrentYearMonth()
	if utils.IsBeforeMonth(year, month, curYear, curMonth) {
		return nil, newAPIError("MONTH_OUT_OF_RANGE", "表示対象月が当月より前です。", 400, nil)
	}
	if utils.IsAfterMonth(year, month, bookableYear, bookableMonth) {
		return nil, newAPIError("MONTH_OUT_OF_RANGE", "表示対象月が予約可能期間外です。", 400, nil)
	}

	if locationID == "" {
		loc, err := s.store.GetDefaultLocation()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default location")
		}
		locationID = loc.ID
	}

	monthStart := utils.FormatDate(utils.MonthStart(year, month))
	monthEnd := utils.FormatDate(utils.MonthEnd(year, month))

	dates, err := utils.DateRange(monthStart, monthEnd)
	if err != nil {
		return nil, newAPIError("INVALID_REQUEST", "日付の算出に失敗しました。", 400, nil)
	}

	desks, err := s.store.GetDesks(locationID, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get desks")
	}

	reservations, err := s.store.GetReservationsByDateRange(locationID, monthStart, monthEnd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get reservations")
	}

	matrixDesks := make([]freedeskmodel.MatrixDesk, 0, len(desks))
	for _, d := range desks {
		matrixDesks = append(matrixDesks, freedeskmodel.MatrixDesk{
			ID:        d.ID,
			Name:      d.Name,
			SortOrder: d.SortOrder,
			IsActive:  d.IsActive,
		})
	}

	matrixReservations := make([]freedeskmodel.MatrixReservation, 0, len(reservations))
	for _, r := range reservations {
		userName := s.resolveUserName(r.UserID)
		matrixReservations = append(matrixReservations, freedeskmodel.MatrixReservation{
			ID:          r.ID,
			DeskID:      r.DeskID,
			UserID:      r.UserID,
			UserName:    userName,
			ReserveDate: r.ReserveDate,
			IsMine:      r.UserID == userID,
		})
	}

	canGoPrev := utils.IsAfterMonth(year, month, curYear, curMonth)
	canGoNext := utils.IsBeforeMonth(year, month, bookableYear, bookableMonth)

	return &freedeskmodel.MatrixData{
		Year:          year,
		Month:         month,
		Timezone:      utils.Timezone,
		Today:         today,
		BookableUntil: bookableUntil,
		CanGoPrev:     canGoPrev,
		CanGoNext:     canGoNext,
		Desks:         matrixDesks,
		Dates:         dates,
		Reservations:  matrixReservations,
	}, nil
}

func parseYearMonth(date string) (int, int, error) {
	t, err := utils.ParseDate(date)
	if err != nil {
		return 0, 0, err
	}
	return t.Year(), int(t.Month()), nil
}

func (s *Service) resolveUserName(userID string) string {
	user, err := s.client.User.Get(userID)
	if err != nil || user == nil {
		return userID
	}
	name := formatUserFullName(user)
	if name == "" {
		name = user.Username
	}
	return name
}

func formatUserFullName(user *model.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return user.LastName + " " + user.FirstName
	}
	if user.FirstName != "" {
		return user.FirstName
	}
	if user.LastName != "" {
		return user.LastName
	}
	return ""
}

// GetMyReservations returns the user's reservations from today onward.
func (s *Service) GetMyReservations(userID string, limit int) ([]*freedeskmodel.Reservation, error) {
	return s.store.GetReservationsByUser(userID, utils.Today(), limit)
}

// GetReservation returns a reservation by ID with permission check.
func (s *Service) GetReservation(userID, reservationID string) (*freedeskmodel.Reservation, error) {
	res, err := s.store.GetReservation(reservationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get reservation")
	}
	if res == nil {
		return nil, newAPIError("RESERVATION_NOT_FOUND", "予約が見つかりません。", 404, nil)
	}
	if res.UserID != userID && !s.config.IsPluginAdmin(userID) {
		return nil, newAPIError("FORBIDDEN", "この予約を参照する権限がありません。", 403, nil)
	}
	res.UserName = s.resolveUserName(res.UserID)
	return res, nil
}

// CreateReservation creates a new reservation.
func (s *Service) CreateReservation(userID string, req *freedeskmodel.CreateReservationRequest) (*freedeskmodel.Reservation, error) {
	if req.DeskID == "" || req.ReserveDate == "" {
		return nil, newAPIError("INVALID_REQUEST", "desk_id と reserve_date は必須です。", 400, nil)
	}
	if !utils.IsValidDateFormat(req.ReserveDate) {
		return nil, newAPIError("INVALID_REQUEST", "reserve_date の形式が不正です。", 400, nil)
	}

	today := utils.Today()
	bookableUntil := utils.BookableUntil(s.config.GetMaxAdvanceMonths())

	if utils.CompareDates(req.ReserveDate, today) < 0 {
		return nil, newAPIError("DATE_OUT_OF_RANGE", "昨日以前の日付は予約できません。", 400, nil)
	}
	if utils.CompareDates(req.ReserveDate, bookableUntil) > 0 {
		return nil, newAPIError("DATE_OUT_OF_RANGE", "予約可能期間外です。", 400, nil)
	}

	desk, err := s.store.GetDesk(req.DeskID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get desk")
	}
	if desk == nil {
		return nil, newAPIError("DESK_NOT_FOUND", "デスクが見つかりません。", 404, nil)
	}
	if !desk.IsActive {
		return nil, newAPIError("DESK_INACTIVE", "無効化されたデスクです。", 400, nil)
	}

	if s.config.GetOneDeskPerDay() {
		existing, listErr := s.store.GetReservationsByUser(userID, req.ReserveDate, 10)
		if listErr != nil {
			return nil, errors.Wrap(listErr, "failed to check existing reservations")
		}
		for _, r := range existing {
			if r.ReserveDate == req.ReserveDate {
				return nil, newAPIError("USER_ALREADY_RESERVED", "同日に既に予約があります。", 409, map[string]any{
					"reserve_date": req.ReserveDate,
				})
			}
		}
	}

	now := model.GetMillis()
	reservation := &freedeskmodel.Reservation{
		ID:          model.NewId(),
		DeskID:      req.DeskID,
		UserID:      userID,
		ReserveDate: req.ReserveDate,
		CreateAt:    now,
		UpdateAt:    now,
		DeleteAt:    0,
		DeskName:    desk.Name,
	}

	if err := s.store.CreateReservation(reservation); err != nil {
		if errors.Is(err, sqlstore.ErrUniqueViolation) {
			return nil, mapUniqueViolation(err, req)
		}
		return nil, errors.Wrap(err, "failed to create reservation")
	}

	s.notifyReservationCreated(reservation, desk.Name)

	return reservation, nil
}

func mapUniqueViolation(err error, req *freedeskmodel.CreateReservationRequest) *APIError {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "user_date") || strings.Contains(msg, "user_id") {
		return newAPIError("USER_ALREADY_RESERVED", "同日に既に予約があります。", 409, map[string]any{
			"reserve_date": req.ReserveDate,
		})
	}
	return newAPIError("DESK_ALREADY_RESERVED", "指定されたデスクは既に予約されています。", 409, map[string]any{
		"desk_id":      req.DeskID,
		"reserve_date": req.ReserveDate,
	})
}

// DeleteReservation cancels a reservation.
func (s *Service) DeleteReservation(userID, reservationID string) error {
	res, err := s.store.GetReservation(reservationID)
	if err != nil {
		return errors.Wrap(err, "failed to get reservation")
	}
	if res == nil {
		return newAPIError("RESERVATION_NOT_FOUND", "予約が見つかりません。", 404, nil)
	}

	today := utils.Today()
	if utils.CompareDates(res.ReserveDate, today) < 0 {
		return newAPIError("DATE_OUT_OF_RANGE", "昨日以前の予約は取消できません。", 400, nil)
	}

	if res.UserID != userID && !s.config.IsPluginAdmin(userID) {
		return newAPIError("FORBIDDEN", "この予約を取消する権限がありません。", 403, nil)
	}

	now := model.GetMillis()
	if err := s.store.DeleteReservation(reservationID, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newAPIError("RESERVATION_NOT_FOUND", "予約が見つかりません。", 404, nil)
		}
		return errors.Wrap(err, "failed to delete reservation")
	}

	s.notifyReservationDeleted(res, s.resolveUserName(res.UserID))
	s.client.Log.Info("Reservation cancelled", "reservation_id", reservationID, "user_id", userID, "proxy", res.UserID != userID)

	return nil
}
