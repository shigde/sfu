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
	"github.com/stretchr/testify/assert"
)

const bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.h3ygBKXYiYVyGIwEMNYVuejBUCch2eysey4JqsXg9dk"

var router *mux.Router
var repository *StreamRepository
var streamId string

func init() {
	jwt := &auth.JwtToken{Enabled: true, Key: "SecretValueReplaceThis", DefaultExpireTime: 604800}
	config := &auth.AuthConfig{JWT: jwt}
	// Setup mocks
	repository = newStreamRepository()
	s := StreamResource{}
	streamId = repository.AddStream(s)
	router = newRouter(config, repository)
}

func TestGetAllStreamsReq(t *testing.T) {

	// When: GET /streams is called
	req := newRequest("GET", "/streams", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`[{"Id":"%s"}]%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestGetStreamReq(t *testing.T) {

	// When: GET /catalog/products is called
	req := newRequest("GET", fmt.Sprintf("/stream/%s", streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`{"Id":"%s"}%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestCreateStreamReq(t *testing.T) {
	jsonStream, _ := json.Marshal(StreamResource{})
	body := bytes.NewBuffer(jsonStream)
	req := newRequest("POST", "/stream", body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Regexp(t, "/stream/[a-f0-9]+", rr.Header()["Location"][0])
	newStreamId := strings.TrimPrefix(rr.Header()["Location"][0], "/stream/")
	assert.True(t, repository.Contains(newStreamId))
}

func TestUpdateStreamReq(t *testing.T) {
	p, _ := repository.StreamById(streamId)
	jsonStream, _ := json.Marshal(p)
	body := bytes.NewBuffer(jsonStream)
	req := newRequest("PUT", "/stream", body)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, repository.Contains(streamId))
}

func TestDeleteStreamReq(t *testing.T) {
	req := newRequest("DELETE", fmt.Sprintf("/stream/%s", streamId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, 0, len(repository.AllStreams()))
}

func newRequest(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)
	return req
}
