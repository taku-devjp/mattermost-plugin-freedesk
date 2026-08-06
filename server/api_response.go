package main

import (
	"encoding/json"
	"net/http"

	"github.com/taku-devjp/mattermost-plugin-freedesk/server/service"
)

type apiResponse struct {
	Data any `json:"data,omitempty"`
}

type apiErrorBody struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, apiResponse{Data: data})
}

func writeServiceError(w http.ResponseWriter, err error) {
	var apiErr *service.APIError
	if e, ok := err.(*service.APIError); ok {
		apiErr = e
	} else {
		apiErr = &service.APIError{
			Code:       "INTERNAL_ERROR",
			Message:    "内部エラーが発生しました。",
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	writeJSON(w, apiErr.HTTPStatus, apiErrorBody{
		Error: apiErrorDetail{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		},
	})
}

func userIDFromRequest(r *http.Request) string {
	return r.Header.Get("Mattermost-User-ID")
}
