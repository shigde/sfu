package media

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/stream"
	"github.com/stretchr/testify/assert"
)

func testWhepReqSetup(t *testing.T) (*mux.Router, string) {
	t.Helper()

	// Setup space
	lobbyManager := newTestLobbyManager()
	store := newTestStore()
	manager, _ := stream.NewSpaceManager(lobbyManager, store)
	space, _ := manager.GetOrCreateSpace(context.Background(), spaceId)

	// Setup Stream
	s := &stream.LiveStream{}
	streamId, _ := space.LiveStreamRepo.Add(context.Background(), s)
	router := NewRouter(securityConfig, rtpConfig, manager)
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

func TestWhepOfferReq(t *testing.T) {
	t.Run("Request to start WHEP, but have no active web session", func(t *testing.T) {
		router, streamId := testWhepReqSetup(t)

		req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", spaceId, streamId), nil, 0)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Then: status is 403 because no active web session
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Request to start and listen WHEP", func(t *testing.T) {
		router, streamId := testWhepReqSetup(t)
		sessionCookie, reqToken := runWhipRequest(t, router, streamId)

		startReq := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", spaceId, streamId), nil, 0)
		startReq.AddCookie(sessionCookie)
		startReq.Header.Set(reqTokenHeaderName, reqToken)
		startRr := httptest.NewRecorder()
		router.ServeHTTP(startRr, startReq)

		assert.Equal(t, http.StatusCreated, startRr.Code)
		assert.Equal(t, "application/sdp", startRr.Header().Get("Content-Type"))
		assert.Equal(t, strconv.Itoa(len([]byte(testOffer))), startRr.Header().Get("Content-Length"))
		assert.Equal(t, testOffer, startRr.Body.String())

		answer := []byte(testAnswer)
		body := bytes.NewBuffer(answer)
		reqToken = startRr.Header().Get(reqTokenHeaderName)

		listenReq := newSDPContentRequest("PATCH", fmt.Sprintf("/space/%s/stream/%s/whep", spaceId, streamId), body, len(answer))
		listenReq.AddCookie(sessionCookie)
		listenReq.Header.Set(reqTokenHeaderName, reqToken)
		listenRr := httptest.NewRecorder()
		router.ServeHTTP(listenRr, listenReq)

		assert.Equal(t, http.StatusCreated, listenRr.Code)
		assert.Equal(t, "application/sdp", listenRr.Header().Get("Content-Type"))
		assert.Equal(t, "0", listenRr.Header().Get("Content-Length"))
	})
}
