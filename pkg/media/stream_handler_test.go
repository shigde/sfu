package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/lobby"
	"github.com/shigde/sfu/pkg/stream"
	"github.com/stretchr/testify/assert"
)

const bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.h3ygBKXYiYVyGIwEMNYVuejBUCch2eysey4JqsXg9dk"

func testStreamsReqSetup(t *testing.T) (string, *mux.Router, *stream.LiveStreamRepository) {
	t.Helper()
	jwt := &auth.JwtToken{Enabled: true, Key: "SecretValueReplaceThis", DefaultExpireTime: 604800}
	config := &auth.AuthConfig{JWT: jwt}
	// Setup space
	lobbyManager := lobby.NewLobbyManager()
	store := newTestStore()
	manager, _ := stream.NewSpaceManager(lobbyManager, store)
	space, _ := manager.GetOrCreateSpace(context.Background(), spaceId)

	// Setup Stream
	s := &stream.LiveStream{}
	streamId, _ := space.LiveStreamRepo.Add(context.Background(), s)
	router := NewRouter(config, manager)

	return streamId, router, space.LiveStreamRepo
}

func TestGetAllStreamsReq(t *testing.T) {
	streamId, router, _ := testStreamsReqSetup(t)

	// When: GET /streams is called
	req := newJsonContentRequest("GET", fmt.Sprintf("/space/%s/streams", spaceId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`[{"Id":"%s"}]%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestGetStreamReq(t *testing.T) {
	streamId, router, _ := testStreamsReqSetup(t)

	// When: GET /catalog/products is called
	req := newJsonContentRequest("GET", fmt.Sprintf("/space/%s/stream/%s", spaceId, streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`{"Id":"%s"}%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestCreateStreamReq(t *testing.T) {
	_, router, repository := testStreamsReqSetup(t)
	url := fmt.Sprintf("/space/%s/stream", spaceId)
	locationPrefix := fmt.Sprintf("%s/", url)
	locationRx := fmt.Sprintf("^%s/[a-f0-9]+", url)

	jsonStream, _ := json.Marshal(stream.LiveStream{})
	body := bytes.NewBuffer(jsonStream)
	req := newJsonContentRequest("POST", url, body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Regexp(t, locationRx, rr.Header()["Location"][0])
	newStreamId := strings.TrimPrefix(rr.Header()["Location"][0], locationPrefix)
	assert.True(t, repository.Contains(context.Background(), newStreamId))
}

func TestUpdateStreamReq(t *testing.T) {
	streamId, router, repository := testStreamsReqSetup(t)

	p, _ := repository.FindById(context.Background(), streamId)
	jsonStream, _ := json.Marshal(p)
	body := bytes.NewBuffer(jsonStream)
	req := newJsonContentRequest("PUT", fmt.Sprintf("/space/%s/stream", spaceId), body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, repository.Contains(context.Background(), streamId))
}

func TestDeleteStreamReq(t *testing.T) {
	streamId, router, repository := testStreamsReqSetup(t)

	req := newJsonContentRequest("DELETE", fmt.Sprintf("/space/%s/stream/%s", spaceId, streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, int64(0), repository.Len(context.Background()))
}

func newJsonContentRequest(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)
	return req
}
