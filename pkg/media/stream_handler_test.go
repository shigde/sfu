package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/engine"
	"github.com/stretchr/testify/assert"
)

const bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.h3ygBKXYiYVyGIwEMNYVuejBUCch2eysey4JqsXg9dk"
const spaceId = "abc123"

func testSetup(t *testing.T) (string, *mux.Router, *engine.RtpStreamRepository) {
	t.Helper()
	jwt := &auth.JwtToken{Enabled: true, Key: "SecretValueReplaceThis", DefaultExpireTime: 604800}
	config := &auth.AuthConfig{JWT: jwt}
	// Setup engine  mocks
	manager := engine.NewSpaceManager()
	space := manager.GetOrCreateSpace(spaceId)

	s := engine.RtpStream{}
	streamId := space.PublicStreamRepo.Add(&s)
	router := NewRouter(config, manager)

	return streamId, router, space.PublicStreamRepo
}

func TestGetAllStreamsReq(t *testing.T) {
	streamId, router, _ := testSetup(t)

	// When: GET /streams is called
	req := newRequest("GET", "/space/abc123/streams", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`[{"Id":"%s"}]%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestGetStreamReq(t *testing.T) {
	streamId, router, _ := testSetup(t)

	// When: GET /catalog/products is called
	req := newRequest("GET", fmt.Sprintf("/space/abc123/stream/%s", streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`{"Id":"%s"}%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestCreateStreamReq(t *testing.T) {
	_, router, repository := testSetup(t)

	jsonStream, _ := json.Marshal(engine.RtpStream{})
	body := bytes.NewBuffer(jsonStream)
	req := newRequest("POST", "/space/abc123/stream", body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Regexp(t, "/stream/[a-f0-9]+", rr.Header()["Location"][0])
	newStreamId := strings.TrimPrefix(rr.Header()["Location"][0], "/stream/")
	assert.True(t, repository.Contains(newStreamId))
}

func TestUpdateStreamReq(t *testing.T) {
	streamId, router, repository := testSetup(t)

	p, _ := repository.FindById(streamId)
	jsonStream, _ := json.Marshal(p)
	body := bytes.NewBuffer(jsonStream)
	req := newRequest("PUT", "/space/abc123/stream", body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, repository.Contains(streamId))
}

func TestDeleteStreamReq(t *testing.T) {
	streamId, router, repository := testSetup(t)

	req := newRequest("DELETE", fmt.Sprintf("/stream/%s", streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, 0, len(repository.All()))
}

func newRequest(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)
	return req
}
