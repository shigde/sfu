package media

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/stream"
	"github.com/stretchr/testify/assert"
)

func testWhepReqSetup(t *testing.T) (*mux.Router, string) {
	t.Helper()
	jwt := &auth.JwtToken{Enabled: true, Key: "SecretValueReplaceThis", DefaultExpireTime: 604800}
	config := &auth.AuthConfig{JWT: jwt}

	// Setup space
	lobbyManager := newTestLobbyManager()
	store := newTestStore()
	manager, _ := stream.NewSpaceManager(lobbyManager, store)
	space, _ := manager.GetOrCreateSpace(context.Background(), spaceId)

	// Setup Stream
	s := &stream.LiveStream{}
	streamId, _ := space.LiveStreamRepo.Add(context.Background(), s)
	router := NewRouter(config, manager)
	return router, streamId
}

func runWhipRequest(t *testing.T, router *mux.Router, streamId string) string {
	offer := []byte(testOffer)
	body := bytes.NewBuffer(offer)

	req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whip", spaceId, streamId), body, len(offer))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	xx := rr.Result().Cookies()
	fmt.Printf("%v", xx)
	return rr.Header().Get("Set-Cookie")
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

	t.Run("Request to start WHEP", func(t *testing.T) {
		router, streamId := testWhepReqSetup(t)
		sessionCookie := runWhipRequest(t, router, streamId)

		req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", spaceId, streamId), nil, 0)
		req.Header.Set("Cookie", sessionCookie)
		// req.AddCookie(sessionCookie)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Then: status is 403 because no active web session
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

//func TestWhepReq(t *testing.T) {
//	router, streamId := testWhepReqSetup(t)
//	offer := []byte(testOffer)
//	body := bytes.NewBuffer(offer)
//
//	req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", spaceId, streamId), body, len(offer))
//	rr := httptest.NewRecorder()
//	router.ServeHTTP(rr, req)
//
//	// Then: status is 201
//	assert.Equal(t, http.StatusCreated, rr.Code)
//	assert.Equal(t, testAnswerETag, rr.Header().Get("ETag"))
//	assert.Equal(t, "application/sdp", rr.Header().Get("Content-Type"))
//	assert.Equal(t, strconv.Itoa(len([]byte(testAnswer))), rr.Header().Get("Content-Length"))
//	assert.Regexp(t, "^session.id=[a-zA-z0-9]+", rr.Header().Get("Set-Cookie"))
//	assert.Equal(t, testAnswer, rr.Body.String())
//}
