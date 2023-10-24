package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetSettingReq(t *testing.T) {
	th, _, _, _, bearer := testRouterSetup(t)

	// When: GET /streams is called
	req := newJsonContentRequest("GET", "/space/setting", nil, bearer)
	rr := httptest.NewRecorder()
	th.router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)
}

func getCsrfRequestToken(t *testing.T, router *mux.Router, bearer string) string {
	t.Helper()
	req := newJsonContentRequest("GET", "/space/setting", nil, bearer)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result().Cookies()[0].Value
}
