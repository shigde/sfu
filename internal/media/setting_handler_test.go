package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/stream"
	"github.com/stretchr/testify/assert"
)

func testSettingReqSetup(t *testing.T) *mux.Router {
	t.Helper()
	// Setup space
	lobbyManager := newTestLobbyManager()
	store := newTestStore()
	manager, _ := stream.NewSpaceManager(lobbyManager, store)

	router := NewRouter(securityConfig, rtpConfig, manager)
	return router
}

func TestGetSettingReq(t *testing.T) {
	router := testSettingReqSetup(t)

	// When: GET /streams is called
	req := newJsonContentRequest("GET", "/space/setting", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)
}

func getCsrfRequestToken(t *testing.T, router *mux.Router) string {
	t.Helper()
	req := newJsonContentRequest("GET", "/space/setting", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result().Cookies()[0].Value
}
