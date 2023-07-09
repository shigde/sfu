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
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/stream"
	"github.com/stretchr/testify/assert"
)

func testWhipReqSetup(t *testing.T) (*mux.Router, string) {
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

func TestWhipReq(t *testing.T) {
	router, streamId := testWhipReqSetup(t)
	offer := []byte(testOffer)
	body := bytes.NewBuffer(offer)
	req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whip", spaceId, streamId), body, len(offer))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 201
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "etag", rr.Header().Get("ETag"))
	assert.Equal(t, "application/sdp", rr.Header().Get("Content-Type"))
	assert.Equal(t, "1400", rr.Header().Get("Content-Length"))
	assert.Equal(t, "https://whip.example.com/resource/id", rr.Header().Get("Location"))
	assert.Equal(t, "session-id", rr.Header().Get("Cookie")) // ??? session cookie
	assert.Equal(t, testAnswer, rr.Body.String())
}

func newSDPContentRequest(method string, url string, body io.Reader, len int) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(len))
	req.Header.Set("Authorization", bearer)
	return req
}
