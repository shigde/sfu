package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/stream"
	"github.com/stretchr/testify/assert"
)

func testWhipReqSetup(t *testing.T) (*mux.Router, string) {
	t.Helper()

	// Setup space
	lobbyManager := newTestLobbyManager()
	store := newTestStore()
	manager := stream.NewSpaceManager(lobbyManager, store, nil)
	space, _ := manager.GetOrCreateSpace(context.Background(), spaceId)

	// Setup Stream
	s := &stream.LiveStream{}
	streamId, _ := space.LiveStreamRepo.Add(context.Background(), s)
	router := NewRouter(securityConfig, rtpConfig, nil, manager)
	return router, streamId
}

func runWhipRequest(t *testing.T, router *mux.Router, streamId string) (*http.Cookie, string) {
	t.Helper()

	offer := []byte(testOffer)
	body := bytes.NewBuffer(offer)

	req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whip", spaceId, streamId), body, len(offer))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	session := rr.Result().Cookies()[0]
	csrfToken := rr.Header().Get(reqTokenHeaderName)
	return session, csrfToken
}

func TestWhipReq(t *testing.T) {
	router, streamId := testWhipReqSetup(t)
	resourceRxp := fmt.Sprintf("^resource/%s", resourceID)
	offer := []byte(testOffer)
	body := bytes.NewBuffer(offer)

	req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whip", spaceId, streamId), body, len(offer))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 201
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, testAnswerETag, rr.Header().Get("ETag"))
	assert.Equal(t, "application/sdp", rr.Header().Get("Content-Type"))
	assert.Equal(t, strconv.Itoa(len([]byte(testAnswer))), rr.Header().Get("Content-Length"))
	assert.Regexp(t, resourceRxp, rr.Header().Get("Location"))
	assert.Regexp(t, "^session.id=[a-zA-z0-9]+", rr.Header().Get("Set-Cookie"))
	assert.Equal(t, testAnswer, rr.Body.String())
}

func TestWhipDeleteReq(t *testing.T) {
	router, streamId := testWhipReqSetup(t)
	sessionCookie, reqToken := runWhipRequest(t, router, streamId)

	req := newSDPContentRequest("DELETE", fmt.Sprintf("/space/%s/stream/%s/whip", spaceId, streamId), nil, 0)
	req.AddCookie(sessionCookie)
	req.Header.Set(reqTokenHeaderName, reqToken)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)
}

func newSDPContentRequest(method string, url string, body io.Reader, len int) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(len))
	req.Header.Set("Authorization", bearer)
	return req
}
