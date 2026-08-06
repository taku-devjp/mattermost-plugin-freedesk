package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) initRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(p.MattermostAuthorizationRequired)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/matrix", p.handleGetMatrix).Methods(http.MethodGet)
	api.HandleFunc("/reservations/mine", p.handleGetMyReservations).Methods(http.MethodGet)
	api.HandleFunc("/reservations/{id}", p.handleGetReservation).Methods(http.MethodGet)
	api.HandleFunc("/reservations", p.handleCreateReservation).Methods(http.MethodPost)
	api.HandleFunc("/reservations/{id}", p.handleDeleteReservation).Methods(http.MethodDelete)
	api.HandleFunc("/desks", p.handleGetDesks).Methods(http.MethodGet)
	api.HandleFunc("/locations", p.handleGetLocations).Methods(http.MethodGet)
	api.HandleFunc("/config", p.handleGetConfig).Methods(http.MethodGet)

	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(p.PluginAdminAuthorizationRequired)
	admin.HandleFunc("/desks", p.handleCreateDesk).Methods(http.MethodPost)
	admin.HandleFunc("/desks/{id}", p.handleUpdateDesk).Methods(http.MethodPut)
	admin.HandleFunc("/desks/{id}", p.handleDeleteDesk).Methods(http.MethodDelete)

	return router
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) MattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) PluginAdminAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		if !p.getConfiguration().IsPluginAdmin(userID) {
			writeServiceError(w, forbiddenError("プラグイン管理者権限が必要です。"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
