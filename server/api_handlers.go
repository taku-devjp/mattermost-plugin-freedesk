package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	freedeskmodel "github.com/taku-devjp/mattermost-plugin-freedesk/server/model"
	"github.com/taku-devjp/mattermost-plugin-freedesk/server/service"
)

func (p *Plugin) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	data := p.service.GetConfig(userID)
	writeData(w, http.StatusOK, data)
}

func (p *Plugin) handleGetMatrix(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	q := r.URL.Query()

	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	locationID := q.Get("location_id")

	data, err := p.service.GetMatrix(userID, year, month, locationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, data)
}

func (p *Plugin) handleGetMyReservations(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	reservations, err := p.service.GetMyReservations(userID, limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"reservations": reservations})
}

func (p *Plugin) handleGetReservation(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := mux.Vars(r)["id"]

	res, err := p.service.GetReservation(userID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func (p *Plugin) handleCreateReservation(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)

	var req freedeskmodel.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, invalidRequestError("リクエストボディが不正です。"))
		return
	}

	res, err := p.service.CreateReservation(userID, &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, res)
}

func (p *Plugin) handleDeleteReservation(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := mux.Vars(r)["id"]

	if err := p.service.DeleteReservation(userID, id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (p *Plugin) handleGetDesks(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	q := r.URL.Query()
	locationID := q.Get("location_id")
	includeInactive := q.Get("include_inactive") == "true"

	if includeInactive && !p.getConfiguration().IsPluginAdmin(userID) {
		writeServiceError(w, forbiddenError("無効デスクの取得権限がありません。"))
		return
	}

	desks, err := p.service.GetDesks(locationID, includeInactive)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"desks": desks})
}

func (p *Plugin) handleGetLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := p.service.GetLocations()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"locations": locations})
}

func (p *Plugin) handleCreateDesk(w http.ResponseWriter, r *http.Request) {
	var req freedeskmodel.CreateDeskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, invalidRequestError("リクエストボディが不正です。"))
		return
	}

	desk, err := p.service.CreateDesk(&req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, desk)
}

func (p *Plugin) handleUpdateDesk(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req freedeskmodel.UpdateDeskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceError(w, invalidRequestError("リクエストボディが不正です。"))
		return
	}

	desk, err := p.service.UpdateDesk(id, &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, desk)
}

func (p *Plugin) handleDeleteDesk(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := p.service.DeleteDesk(id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func invalidRequestError(message string) *service.APIError {
	return &service.APIError{Code: "INVALID_REQUEST", Message: message, HTTPStatus: http.StatusBadRequest}
}

func forbiddenError(message string) *service.APIError {
	return &service.APIError{Code: "FORBIDDEN", Message: message, HTTPStatus: http.StatusForbidden}
}
