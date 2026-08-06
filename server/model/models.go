package model

// Location represents a freedesk_locations row.
type Location struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	CreateAt  int64  `json:"create_at" db:"create_at"`
	UpdateAt  int64  `json:"update_at" db:"update_at"`
	DeleteAt  int64  `json:"delete_at" db:"delete_at"`
}

// Desk represents a freedesk_desks row.
type Desk struct {
	ID          string  `json:"id" db:"id"`
	LocationID  string  `json:"location_id" db:"location_id"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
	SortOrder   int     `json:"sort_order" db:"sort_order"`
	IsActive    bool    `json:"is_active" db:"is_active"`
	CreateAt    int64   `json:"create_at" db:"create_at"`
	UpdateAt    int64   `json:"update_at" db:"update_at"`
	DeleteAt    int64   `json:"delete_at" db:"delete_at"`
}

// Reservation represents a freedesk_reservations row.
type Reservation struct {
	ID          string  `json:"id" db:"id"`
	DeskID      string  `json:"desk_id" db:"desk_id"`
	UserID      string  `json:"user_id" db:"user_id"`
	ReserveDate string  `json:"reserve_date" db:"reserve_date"`
	Note        *string `json:"note,omitempty" db:"note"`
	CreateAt    int64   `json:"create_at" db:"create_at"`
	UpdateAt    int64   `json:"update_at" db:"update_at"`
	DeleteAt    int64   `json:"delete_at" db:"delete_at"`

	DeskName string `json:"desk_name,omitempty" db:"desk_name"`
	UserName string `json:"user_name,omitempty" db:"-"`
}

// MatrixDesk is a desk entry in the matrix API response.
type MatrixDesk struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

// MatrixReservation is a reservation entry in the matrix API response.
type MatrixReservation struct {
	ID          string `json:"id"`
	DeskID      string `json:"desk_id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ReserveDate string `json:"reserve_date"`
	IsMine      bool   `json:"is_mine"`
}

// MatrixData is the GET /matrix response payload.
type MatrixData struct {
	Year          int                 `json:"year"`
	Month         int                 `json:"month"`
	Timezone      string              `json:"timezone"`
	Today         string              `json:"today"`
	BookableUntil string              `json:"bookable_until"`
	CanGoPrev     bool                `json:"can_go_prev"`
	CanGoNext     bool                `json:"can_go_next"`
	Desks         []MatrixDesk        `json:"desks"`
	Dates         []string            `json:"dates"`
	Reservations  []MatrixReservation `json:"reservations"`
}

// ConfigData is the GET /config response payload.
type ConfigData struct {
	Timezone            string `json:"timezone"`
	Today               string `json:"today"`
	MaxAdvanceDays      int    `json:"max_advance_days"`
	BookableUntil       string `json:"bookable_until"`
	OneDeskPerDay       bool   `json:"one_desk_per_day"`
	NotificationEnabled bool   `json:"notification_enabled"`
	IsPluginAdmin       bool   `json:"is_plugin_admin"`
}

// CreateReservationRequest is the POST /reservations body.
type CreateReservationRequest struct {
	DeskID      string `json:"desk_id"`
	ReserveDate string `json:"reserve_date"`
}

// CreateDeskRequest is the POST /admin/desks body.
type CreateDeskRequest struct {
	LocationID string `json:"location_id"`
	Name       string `json:"name"`
	SortOrder  *int   `json:"sort_order"`
	IsActive   *bool  `json:"is_active"`
}

// UpdateDeskRequest is the PUT /admin/desks/{id} body.
type UpdateDeskRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}
