package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/taku-devjp/mattermost-plugin-freedesk/server/service"
)

func TestUnauthorizedRequest(t *testing.T) {
	plugin := Plugin{}
	plugin.router = plugin.initRouter()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)

	plugin.ServeHTTP(nil, w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestGetConfigUnauthorizedBody(t *testing.T) {
	plugin := Plugin{
		configuration: &configuration{},
	}
	plugin.router = plugin.initRouter()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	r.Header.Set("Mattermost-User-ID", "user-1")

	plugin.service = service.New(nil, nil, &plugin)

	plugin.ServeHTTP(nil, w, r)

	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)

	var resp struct {
		Data struct {
			Timezone string `json:"timezone"`
			Today    string `json:"today"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, "Asia/Tokyo", resp.Data.Timezone)
	assert.NotEmpty(t, resp.Data.Today)
}
