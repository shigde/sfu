package media

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllStreamsReq(t *testing.T) {
	th, space, liveStream, _, bearer := testRouterSetup(t)

	// When: GET /streams is called
	req := newJsonContentRequest("GET", fmt.Sprintf("/space/%s/streams", space.Identifier), nil, bearer)
	rr := httptest.NewRecorder()
	th.router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`[{"uuid":"%s","title":"%s","user":"%s"}]%s`, liveStream.UUID, liveStream.Title, liveStream.User, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

func TestGetStreamReq(t *testing.T) {
	th, space, liveStream, _, bearer := testRouterSetup(t)

	// When: GET /catalog/products is called
	req := newJsonContentRequest("GET", fmt.Sprintf("/space/%s/stream/%s", space.Identifier, liveStream.UUID.String()), nil, bearer)
	rr := httptest.NewRecorder()
	th.router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`{"uuid":"%s","title":"%s","user":"%s"}%s`, liveStream.UUID, liveStream.Title, liveStream.User, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}

// I commented out the Delete/Update/Create Space Rest endpoints.
// Currently, spaces and streams created by activity pub endpoints

//func TestCreateStreamReq(t *testing.T) {
//	th, space, _, _, bearer := testRouterSetup(t)
//
//	url := fmt.Sprintf("/space/%s/stream", space.Identifier)
//	locationPrefix := fmt.Sprintf("%s/", url)
//	locationRx := fmt.Sprintf("^%s/[a-f0-9]+", url)
//
//	jsonStream, _ := json.Marshal(stream.LiveStream{})
//	body := bytes.NewBuffer(jsonStream)
//	req := newJsonContentRequest("POST", url, body, bearer)
//
//	rr := httptest.NewRecorder()
//	th.router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusCreated, rr.Code)
//	assert.Regexp(t, locationRx, rr.Header()["Location"][0])
//	newStreamId := strings.TrimPrefix(rr.Header()["Location"][0], locationPrefix)
//	assert.True(t, th.liveStreamRepo.Contains(context.Background(), newStreamId))
//}

//func TestUpdateStreamReq(t *testing.T) {
//	th, space, liveStream, _, bearer := testRouterSetup(t)
//
//	p, _ := th.liveStreamRepo.FindByUuid(context.Background(), liveStream.UUID.String())
//	jsonStream, _ := json.Marshal(p)
//	body := bytes.NewBuffer(jsonStream)
//	req := newJsonContentRequest("PUT", fmt.Sprintf("/space/%s/stream", space.Identifier), body, bearer)
//
//	rr := httptest.NewRecorder()
//	th.router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusNoContent, rr.Code)
//	assert.True(t, th.liveStreamRepo.Contains(context.Background(), liveStream.UUID.String()))
//}

//func TestDeleteStreamReq(t *testing.T) {
//	th, space, liveStream, _, bearer := testRouterSetup(t)
//
//	req := newJsonContentRequest("DELETE", fmt.Sprintf("/space/%s/stream/%s", space.Identifier, liveStream.UUID.String()), nil, bearer)
//	rr := httptest.NewRecorder()
//	th.router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusOK, rr.Code)
//	// assert.Equal(t, int64(0), th.liveStreamRepo.Len(context.Background()))
//}

func newJsonContentRequest(method string, url string, body io.Reader, bearer string) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)
	return req
}
